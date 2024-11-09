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

// getSound retrieves sound data from the database.
func (m *Model) getSound(soundName string) ([]byte, error) {
	var fileData []byte
	err := m.Db.QueryRow("SELECT file FROM sounds WHERE name = ?", soundName).Scan(&fileData)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("sound not found: %s", soundName)
		}
		dlog.ErrorLog.Println("error querying sound file from database:", err)
		return nil, err
	}
	return fileData, nil
}

// openSound opens a sound file from the filesystem.
func openSound(fileName string) (*os.File, error) {
	file, err := os.Open(fileName)
	if err != nil {
		dlog.ErrorLog.Println("error opening sound file:", err)
		return nil, err
	}
	return file, nil
}

// encodeSound reads encoded PCM data from a file reader and returns the buffer of frames.
func encodeSound(file io.Reader) ([][]byte, error) {
	var opusLen int16
	var buffer [][]byte

	for {
		// Read opus frame length from file
		err := binary.Read(file, binary.LittleEndian, &opusLen)
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			// End of file reached
			break
		}
		if err != nil {
			dlog.ErrorLog.Println("error reading opus frame length:", err)
			return nil, err
		}

		if opusLen <= 0 {
			dlog.ErrorLog.Println("invalid opus frame length:", opusLen)
			return nil, fmt.Errorf("invalid opus frame length: %d", opusLen)
		}

		// Read encoded PCM data
		inBuf := make([]byte, opusLen)
		err = binary.Read(file, binary.LittleEndian, &inBuf)
		if err != nil {
			dlog.ErrorLog.Println("error reading PCM data:", err)
			return nil, err
		}

		buffer = append(buffer, inBuf)
	}

	return buffer, nil
}

// LoadSound loads a sound from the database and returns the buffer.
func (m *Model) LoadSound(soundName string) ([][]byte, error) {
	fileData, err := m.getSound(soundName)
	if err != nil {
		return nil, err
	}
	file := bytes.NewReader(fileData)
	encodedFile, err := encodeSound(file)
	if err != nil {
		return nil, err
	}
	return encodedFile, nil
}

// LoadSoundFS loads a sound from the filesystem and returns the buffer.
func (m *Model) LoadSoundFS(soundName string) ([][]byte, error) {
	file, err := openSound(soundName)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	encodedFile, err := encodeSound(file)
	if err != nil {
		return nil, err
	}
	return encodedFile, nil
}

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
		return nil
	}

	_, err = m.Db.Exec(`
    INSERT INTO sounds (name, category_id, hash, file)
    VALUES (?, ?, ?, ?)
    ON CONFLICT(name) DO UPDATE SET
        hash = excluded.hash,
        file = excluded.file,
        category_id = excluded.category_id
	`, fileName, categoryID, fileHash, fileData)

	if err != nil {
		dlog.ErrorLog.Printf("Error inserting sound into database: %v", err)
	}

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
	rows, err := m.Db.Query("SELECT sounds.name FROM sounds ORDER BY sounds.name LIMIT 25")
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
			sounds = append(sounds, sound.String)
		}
	}
	return sounds, nil
}

// GetCategoryByID returns the ID of a category by its name.
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

// GetSoundsM returns a map of sounds with their hashes (from DB)
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

// AddCategory adds a category to the database.
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
