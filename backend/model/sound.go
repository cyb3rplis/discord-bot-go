package model

import (
	"bytes"
	"database/sql"
	"encoding/binary"
	"fmt"
	"github.com/cyb3rplis/discord-bot-go/dlog"
	"io"
	"os"
)

var Buffer = make([][]byte, 0)

// AddSound adds a sound to the database.
func (m *Model) AddSound(categoryID int, fileName, fileHash string, fileData []byte) error {
	// Check if the sound with the same hash already exists
	var existingID int
	err := m.Db.QueryRow("SELECT id FROM sounds WHERE hash = ?", fileHash).Scan(&existingID)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to check existing sound: %w", err)
	}
	// If the sound already exists, skip the insertion
	if existingID != 0 {
		//dlog.InfoLog.Printf("Sound with hash %s already exists, skipping insertion", fileHash)
		return nil
	}
	alias := RemoveFileExtension(fileName) // Or any other default value, e.g., ""
	_, err = m.Db.Exec("INSERT INTO sounds (name, alias, category_id, hash, file) VALUES (?, ?, ?, ?, ?)", fileName, alias, categoryID, fileHash, fileData)
	return err
}

// RemoveCategory removes a category from the database.
func (m *Model) RemoveCategory(categoryID int) error {
	// ON DELETE CASCADE - sounds will get deleted automatically when the category is deleted
	_, err := m.Db.Exec("DELETE FROM categories WHERE id = ?", categoryID)
	return err
}

// RemoveSound removes a sound from the database.
func (m *Model) RemoveSound(categoryID int, fileName string) error {
	_, err := m.Db.Exec("DELETE FROM sounds WHERE category_id = ? AND name = ?", categoryID, fileName)
	return err
}

// LoadSound attempts to load an encoded sound file from disk.
func (m *Model) LoadSound(soundName string) error {
	var opusLen int16
	var fileData []byte

	// get sounds from database
	err := m.Db.QueryRow("SELECT file FROM sounds WHERE name = ?", soundName).Scan(&fileData)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("sound not found: %s", soundName)
		}
		dlog.ErrorLog.Println("error querying sound file from database:", err)
		return err
	}
	// Create a reader for the file data
	file := bytes.NewReader(fileData)
	for {
		// Read opus frame length from the file data.
		err = binary.Read(file, binary.LittleEndian, &opusLen)

		// If this is the end of the file, just return.
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return nil
		}
		if err != nil {
			dlog.ErrorLog.Println("error reading from file data:", err)
			return err
		}
		// Read encoded PCM from the file data.
		InBuf := make([]byte, opusLen)
		err = binary.Read(file, binary.LittleEndian, &InBuf)
		if err != nil {
			dlog.ErrorLog.Println("error reading from file data:", err)
			return err
		}
		// Append encoded PCM data to the buffer.
		Buffer = append(Buffer, InBuf)
	}
}

// GetSounds returns a list of sounds in the specified category (from DB)
func (m *Model) GetSounds(category string) ([]string, error) {
	rows, err := m.Db.Query("SELECT sounds.name FROM sounds LEFT JOIN categories ON sounds.category_id = categories.id WHERE categories.name = ? ORDER BY sounds.name ASC", category)
	if err != nil {
		return nil, fmt.Errorf("failed to query sounds in category: %w", err)
	}
	defer rows.Close()

	var sounds []string
	for rows.Next() {
		var sound string
		err := rows.Scan(&sound)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sound: %w", err)
		}
		sounds = append(sounds, sound)
	}
	return sounds, nil
}

// GetSoundsAll returns a list of sounds in the specified category (from DB)
func (m *Model) GetSoundsAll() ([]string, error) {
	rows, err := m.Db.Query("SELECT sounds.name FROM sounds ORDER BY sounds.name")
	if err != nil {
		return nil, fmt.Errorf("failed to query sounds: %w", err)
	}
	defer rows.Close()

	var sounds []string
	for rows.Next() {
		var sound sql.NullString
		err := rows.Scan(&sound)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sound: %w", err)
		}
		if sound.Valid {
			if sound.String != "" {
				sounds = append(sounds, sound.String)
			}
		}
	}
	return sounds, nil
}

func (m *Model) GetCategoryByID(folderName string) int {
	var categoryID int
	err := m.Db.QueryRow("SELECT id FROM categories WHERE name = ?", folderName).Scan(&categoryID)
	if err != nil {
		if err == sql.ErrNoRows {
			// This should not happen because the category should exist by this point.
			dlog.InfoLog.Printf("Category %s not found in database", folderName)
		} else {
			dlog.FatalLog.Fatal(err)
		}
	}
	return categoryID
}

func (m *Model) GetSoundsM() map[int]map[string]string {
	rows, err := m.Db.Query("SELECT category_id, name, hash FROM sounds")
	if err != nil {
		dlog.FatalLog.Fatal(err)
	}
	defer rows.Close()

	sounds := make(map[int]map[string]string)
	for rows.Next() {
		var categoryID int
		var fileName, fileHash string
		if err := rows.Scan(&categoryID, &fileName, &fileHash); err != nil {
			dlog.FatalLog.Fatal(err)
		}
		if sounds[categoryID] == nil {
			sounds[categoryID] = make(map[string]string)
		}
		sounds[categoryID][fileName] = fileHash
	}
	return sounds
}

func (m *Model) AddCategory(folderName string) error {
	// Check if the category with the same name already exists
	var existingID int
	err := m.Db.QueryRow("SELECT id FROM categories WHERE name = ?", folderName).Scan(&existingID)
	if err != nil && err != sql.ErrNoRows {
		return fmt.Errorf("failed to check existing category: %w", err)
	}
	// If the category already exists, skip the insertion
	if existingID != 0 {
		//dlog.InfoLog.Printf("Category with name %s already exists, skipping insertion", folderName)
		return nil
	}
	_, err = m.Db.Exec("INSERT INTO categories (name) VALUES (?)", folderName)
	return err
}

// LoadSoundFS loads the sound file from the filesystem.
func (m *Model) LoadSoundFS(soundName string) error {
	var opusLen int16
	file, err := os.Open(soundName)
	if err != nil {
		dlog.ErrorLog.Println("error opening dca file :", err)
		return err
	}
	defer file.Close()
	for {
		// Read opus frame length from dca file.
		err = binary.Read(file, binary.LittleEndian, &opusLen)

		// If this is the end of the file, just return.
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			err := file.Close()
			if err != nil {
				return err
			}
			return nil
		}
		if err != nil {
			dlog.ErrorLog.Println("Error reading from dca file :", err)
			return err
		}

		// Read encoded pcm from dca file.
		InBuf := make([]byte, opusLen)
		err = binary.Read(file, binary.LittleEndian, &InBuf)
		if err != nil {
			dlog.ErrorLog.Println("Error reading from dca file :", err)
			return err
		}

		// Append encoded pcm data to the buffer.
		Buffer = append(Buffer, InBuf)
	}
}
