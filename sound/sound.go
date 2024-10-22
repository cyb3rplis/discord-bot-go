package sound

import (
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cyb3rplis/discord-bot-go/utils"

	"github.com/cyb3rplis/discord-bot-go/logger"
	"github.com/cyb3rplis/discord-bot-go/model"

	"github.com/bwmarrin/discordgo"
)

var buffer = make([][]byte, 0)
var botSpeaking = false
var stopChannel = make(chan struct{})

var userLastInteraction = make(map[string]time.Time)
var userInteractionCount = make(map[string]int)
var mu sync.Mutex

const maxInteractions = 15             // Maximum allowed interactions before timeout
const resetDuration = 15 * time.Second // Duration to reset the interaction count

type Entry struct {
	ID   int
	Name string
}

// LoadSound attempts to load an encoded sound file from disk.
func LoadSound(soundName string) error {
	var opusLen int16
	file, err := os.Open(soundName)
	if err != nil {
		logger.ErrorLog.Println("Error opening dca file :", err)
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
			logger.ErrorLog.Println("Error reading from dca file :", err)
			return err
		}

		// Read encoded pcm from dca file.
		InBuf := make([]byte, opusLen)
		err = binary.Read(file, binary.LittleEndian, &InBuf)
		if err != nil {
			logger.ErrorLog.Println("Error reading from dca file :", err)
			return err
		}

		// Append encoded pcm data to the buffer.
		buffer = append(buffer, InBuf)
	}
}

// PlaySound plays the current buffer to the provided channel.
func PlaySound(s *discordgo.Session, m *discordgo.MessageCreate, st *discordgo.Message, guildID, channelID, soundFile, soundName string) (err error) {

	// check if the bot is currently speaking, and exit early to avoid corrupted sound buffer
	if botSpeaking {
		stopChannel <- struct{}{}
		time.Sleep(150 * time.Millisecond) // Give some time for the current sound to stop
	}

	if soundName == "tts" {
		soundFile = model.Bot.Config.TTSOutput
	}

	// Load the sound file.
	err = LoadSound(soundFile)
	if err != nil {
		logger.ErrorLog.Printf("Error loading sound %s, %v ", soundFile, err)
		_, err = s.ChannelMessageSend(m.ChannelID, "> Sound does not exist\n> Sikerim")
		if err != nil {
			logger.ErrorLog.Println("Error loading sound:", err)
		}
		return err
	}

	// Join the provided voice channel.
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return err
	}
	botSpeaking = true

	// Sleep for a specified amount of time before playing the sound
	time.Sleep(150 * time.Millisecond)

	// Start speaking.
	_ = vc.Speaking(true)

	// Send the buffer data.
	for _, buff := range buffer {
		select {
		case <-stopChannel:
			// Stop sending buffer data if stop signal is received
			_ = vc.Speaking(false)
			botSpeaking = false
			buffer = make([][]byte, 0)
			return nil
		default:
			vc.OpusSend <- buff
		}
	}

	// Stop speaking
	_ = vc.Speaking(false)

	// Sleep for a specified amount of time before ending.
	time.Sleep(250 * time.Millisecond)

	// Disconnect from the provided voice channel.
	//vc.Disconnect()

	// empty buffer to not play older sounds
	buffer = make([][]byte, 0)
	botSpeaking = false

	utils.StopButtonRoutine(s)
	return nil
}

// BuildSoundButtons creates a list of buttons for the provided category
func BuildSoundButtons(sounds []string, category string) ([]discordgo.MessageComponent, error) {
	content := []discordgo.MessageComponent{}
	row := discordgo.ActionsRow{}
	for i, soundName := range sounds {
		soundName = strings.TrimSuffix(soundName, ".dca")
		// only 5 buttons per row - discord does not allow more
		if i > 0 && i%5 == 0 {
			content = append(content, row)
			row = discordgo.ActionsRow{}
		}
		row.Components = append(row.Components, discordgo.Button{
			Label:    soundName,
			Style:    discordgo.SecondaryButton,
			CustomID: fmt.Sprintf("play_sound_%s_%s", category, soundName),
		})
	}
	// Append the last row if it has any components
	if len(row.Components) > 0 {
		content = append(content, row)
	}

	return content, nil
}

// GetCategories returns a list of sound categories (from DB)
func GetCategories() ([]string, error) {
	rows, err := model.Bot.Db.Query("SELECT name FROM categories")
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var category string
		err := rows.Scan(&category)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, category)
	}
	return categories, nil
}

func LoadYouTubeAudio(videoURL string) error {
	// Create ffmpeg and dcaenc pipeline to convert YouTube stream to DCA format
	cmd := exec.Command("bash", "-c", fmt.Sprintf("yt-dlp -x --audio-format mp3 --force-overwrites -o ../dist/yt.mp3 %s", videoURL))
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to download youtube audio: %w", err)
	}

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("error downloading youtube audio: %w", err)
	}

	cmd = exec.Command("bash", "-c", "ffmpeg -i ../dist/yt.mp3 -f s16le -ar 48000 -ac 2 pipe:1 | ../dca > ../dist/yt.dca")
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to to convert youtube audio from mp3 to dca: %w", err)
	}

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("error converting youtube audio: %w", err)
	}

	return nil
}

// getSounds returns a list of sounds in the specified category (from DB)
func getSounds(category string) ([]string, error) {
	rows, err := model.Bot.Db.Query("SELECT sounds.name FROM sounds LEFT JOIN categories ON sounds.category_id = categories.id WHERE categories.name = ? ORDER BY sounds.name ASC", category)
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

// SyncDatabaseWithFileSystem will sync the database with the filesystem.
func SyncDatabaseWithFileSystem(folderMap map[string][]string) error {
	existingCategories := fetchCategories() // Get current folders/categories in DB
	existingSounds := getSoundsM()          // Get current files/sounds in DB

	for folder, files := range folderMap {
		var categoryID int

		// Check if the folder (category) exists in the database
		if dbCategoryID, exists := existingCategories[folder]; exists {
			categoryID = dbCategoryID // The folder already exists
		} else {
			// The folder doesn't exist in the database, so we need to add it
			if err := addCategory(folder); err != nil {
				logger.InfoLog.Printf("Failed to add category %s: %v", folder, err)
				continue
			}

			// Fetch the new category ID after insertion
			categoryID = fetchCategoryID(folder)
		}

		// Add new sounds for this category
		for _, file := range files {
			soundPath := filepath.Join(model.Bot.Config.SoundsDir, folder, file+".dca")
			fileData, err := os.ReadFile(soundPath)
			if err != nil {
				return fmt.Errorf("failed to read sound file: %w", err)
			}

			fileHash, err := computeFileHash(soundPath)
			if err != nil {
				logger.ErrorLog.Printf("Failed to compute hash for file %s: %v", file, err)
				continue
			}

			if fileExistsInDB(existingSounds, categoryID, file, fileHash) {
				// File exists and hasn't changed, skip
				continue
			}

			// File does not exist in the DB, add it
			if err := addSound(categoryID, file, fileHash, fileData); err != nil {
				logger.InfoLog.Printf("Failed to add sound %s to category %s: %v", file, folder, err)
			}
		}
	}

	// Remove entries that no longer exist on the filesystem
	for folder, categoryID := range existingCategories {
		if _, exists := folderMap[folder]; !exists {
			// Folder exists in the database but not in the filesystem
			if err := removeCategory(categoryID); err != nil {
				logger.InfoLog.Printf("Failed to remove category %s (ID: %d): %v", folder, categoryID, err)
			}

			logger.InfoLog.Printf("Removed category %s (ID: %d)", folder, categoryID)
		} else {
			// For existing folders, remove missing files
			dbFiles := existingSounds[categoryID] // Files in the DB
			fsFiles := folderMap[folder]          // Files in the filesystem

			for dbFile := range dbFiles {
				if !fileExistsInFS(fsFiles, dbFile) {
					// File exists in the database but not in the filesystem
					if err := removeSound(categoryID, dbFile); err != nil {
						logger.InfoLog.Printf("Failed to remove sound %s from category %s: %v", dbFile, folder, err)
					}

					logger.InfoLog.Printf("Removed sound %s from category %s", dbFile, folder)
				}
			}
		}
	}

	return nil
}

func fetchCategories() map[string]int {
	rows, err := model.Bot.Db.Query("SELECT id, name FROM categories")
	if err != nil {
		logger.FatalLog.Fatal(err)
	}
	defer rows.Close()

	categories := make(map[string]int)
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			logger.FatalLog.Fatal(err)
		}
		categories[name] = id
	}
	return categories
}

func fetchCategoryID(folderName string) int {
	var categoryID int
	err := model.Bot.Db.QueryRow("SELECT id FROM categories WHERE name = ?", folderName).Scan(&categoryID)
	if err != nil {
		if err == sql.ErrNoRows {
			// This should not happen because the category should exist by this point.
			logger.InfoLog.Printf("Category %s not found in database", folderName)
		} else {
			logger.FatalLog.Fatal(err)
		}
	}
	return categoryID
}

func getSoundsM() map[int]map[string]string {
	rows, err := model.Bot.Db.Query("SELECT category_id, name, hash FROM sounds")
	if err != nil {
		logger.FatalLog.Fatal(err)
	}
	defer rows.Close()

	sounds := make(map[int]map[string]string)
	for rows.Next() {
		var categoryID int
		var fileName, fileHash string
		if err := rows.Scan(&categoryID, &fileName, &fileHash); err != nil {
			logger.FatalLog.Fatal(err)
		}
		if sounds[categoryID] == nil {
			sounds[categoryID] = make(map[string]string)
		}
		sounds[categoryID][fileName] = fileHash
	}
	return sounds
}

func addCategory(folderName string) error {
	_, err := model.Bot.Db.Exec("INSERT INTO categories (name) VALUES (?)", folderName)
	return err
}

func addSound(categoryID int, fileName, fileHash string, fileData []byte) error {
	alias := utils.RemoveFileExtension(fileName) // Or any other default value, e.g., ""
	_, err := model.Bot.Db.Exec("INSERT INTO sounds (name, alias, category_id, hash, file) VALUES (?, ?, ?, ?, ?)", fileName, alias, categoryID, fileHash, fileData)
	return err
}

func removeCategory(categoryID int) error {
	// ON DELETE CASCADE - sounds will get deleted automatically when the category is deleted
	_, err := model.Bot.Db.Exec("DELETE FROM categories WHERE id = ?", categoryID)
	return err
}

func removeSound(categoryID int, fileName string) error {
	_, err := model.Bot.Db.Exec("DELETE FROM sounds WHERE category_id = ? AND name = ?", categoryID, fileName)
	return err
}

func fileExistsInDB(existingSounds map[int]map[string]string, categoryID int, fileName string, fileHash string) bool {
	if soundsInCategory, exists := existingSounds[categoryID]; exists {
		if dbHash, fileExists := soundsInCategory[fileName]; fileExists {
			// Check if the hash matches
			return dbHash == fileHash
		}
	}
	return false
}

func fileExistsInFS(fsFiles []string, fileName string) bool {
	for _, fsFile := range fsFiles {
		if fsFile == fileName {
			return true
		}
	}
	return false
}

// computeFileHash computes the SHA-256 hash of a given file
func computeFileHash(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func PlayYoutubeAudio(s *discordgo.Session, m *discordgo.MessageCreate) (err error) {
	// Find the channel that the interaction came from
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		logger.ErrorLog.Println("Error finding channel:", err)
		return
	}

	// Find the guild for that channel
	g, err := s.State.Guild(c.GuildID)
	if err != nil {
		logger.ErrorLog.Println("Error finding guild:", err)
		return
	}

	// Look for the interaction user in that guild's current voice states
	for _, vs := range g.VoiceStates {
		// Error here with m.Member.User.ID nil pointer dereference
		if vs.UserID == m.Member.User.ID {
			fmt.Println("vs: ", vs)
			// add user and user statistics
			userID, err := strconv.Atoi(m.Member.User.ID)
			if err != nil {
				logger.ErrorLog.Println("Error converting user ID to int:", err)
				return err
			} else {
				err = utils.AddUser(userID, m.Member.User.GlobalName)
				if err != nil {
					logger.ErrorLog.Println("Error adding user:", err)
					return err
				}

				err = utils.AddUserStatistics(userID, "youtube")
				if err != nil {
					logger.ErrorLog.Println("Error adding user statistics:", err)
					return err
				}
			}

			content := []discordgo.MessageComponent{}
			row := discordgo.ActionsRow{}
			row.Components = append(row.Components, discordgo.Button{
				Label:    "Stop Sound",
				Style:    discordgo.DangerButton,
				CustomID: "stop_sound",
			})
			content = append(content, row)

			message := &discordgo.MessageSend{
				Content:    "➡ Currently Playing Youtube Sound by <@" + m.Member.User.ID + ">: ",
				Components: content,
			}

			st := utils.NewComplexMessageRoutine(".stopbutton", m.ChannelID, m.ID, message, s, true)

			logger.InfoLog.Printf("User: %s played youtube sound", m.Member.User.GlobalName)
			soundFile := "../dist/yt.dca"

			// Play the sound
			err = PlaySound(s, &discordgo.MessageCreate{Message: m.Message}, st, g.ID, vs.ChannelID, soundFile, "youtube")
			if err != nil {
				logger.ErrorLog.Println("Error playing sound:", err)
			}
		}
	}

	return nil
}
