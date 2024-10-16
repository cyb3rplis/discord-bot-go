package sound

import (
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cyb3rplis/discord-bot-go/logger"
	"github.com/cyb3rplis/discord-bot-go/model"

	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/config"
)

var cfg = config.GetConfig()

var buffer = make([][]byte, 0)
var botSpeaking = false
var stopChannel = make(chan struct{})

var userLastInteraction = make(map[string]time.Time)
var userInteractionCount = make(map[string]int)
var mu sync.Mutex

const maxInteractions = 15             // Maximum allowed interactions before timeout
const resetDuration = 15 * time.Second // Duration to reset the interaction count

var lastMessageID string
var lastChannelID string

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
		// delete the last message and set the new value to the last sent message
		err = s.ChannelMessageDelete(lastChannelID, lastMessageID)
		if err != nil {
			logger.ErrorLog.Println("Error deleting message with command to play new sound:", err)
			return err
		}
		stopChannel <- struct{}{}
		time.Sleep(150 * time.Millisecond) // Give some time for the current sound to stop
	}

	lastChannelID = st.ChannelID
	lastMessageID = st.ID

	if soundName == "tts" {
		soundFile = cfg.TTSOutput
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

	err = addSoundStatistics(soundName)
	if err != nil {
		logger.ErrorLog.Printf("Error inserting statistics: %v", err)
	}

	// Delete the initial "Now Playing" message
	err = s.ChannelMessageDelete(st.ChannelID, st.ID)
	if err != nil {
		logger.ErrorLog.Println("Error deleting message after sound finished:", err)
		return err
	}
	return nil
}

// InteractionHandler handles interaction events (e.g., button clicks)
func InteractionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}
	customID := i.MessageComponentData().CustomID

	// Acknowledge the interaction
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
	if err != nil {
		logger.ErrorLog.Println("Failed to respond to interaction:", err)
		return
	}

	// Check if the user is spamming the buttons and limit the interactions
	mu.Lock()
	lastInteraction, exists := userLastInteraction[i.Member.User.ID] // Get the last interaction time
	if exists && time.Since(lastInteraction) < resetDuration {       // Check if the user has interacted recently
		userInteractionCount[i.Member.User.ID]++
	} else {
		userInteractionCount[i.Member.User.ID] = 1 // Reset the interaction count
	}
	userLastInteraction[i.Member.User.ID] = time.Now()            // Update the last interaction time
	if userInteractionCount[i.Member.User.ID] > maxInteractions { // Check if the user has exceeded the interaction limit
		mu.Unlock()
		_, err := s.ChannelMessageSend(i.ChannelID, "Stop spamming the buttons <@"+i.Member.User.ID+"> you fucking idiot!!!")
		if err != nil {
			logger.ErrorLog.Println("Error sending message:", err)
		}
		return
	}
	mu.Unlock()

	switch {
	case strings.HasPrefix(customID, "play_sound_"):
		HandlePlaySoundInteraction(s, i, customID)
	case strings.HasPrefix(customID, "list_sounds_"):
		handleListSoundsInteraction(s, i, customID)
	case strings.HasPrefix(customID, "stop_sound"):
		handleStopSoundInteraction(s)
	default:
		logger.ErrorLog.Println("unknown interaction:", customID)
	}
}

func handleStopSoundInteraction(s *discordgo.Session) {
	// check if the bot is currently speaking, and exit
	if botSpeaking {
		stopChannel <- struct{}{}

		// Delete the last "Now Playing" message
		// This should not be needed, since the actual function will finish normally when emptying the buffer
		// and this deletes the old message anyway. Keeping it here just to make sure
		_ = s.ChannelMessageDelete(lastChannelID, lastMessageID)
		time.Sleep(150 * time.Millisecond) // Give some time for the current sound to stop
	}
}

func HandlePlaySoundInteraction(s *discordgo.Session, i *discordgo.InteractionCreate, customID string) {
	cfg := config.GetConfig()
	ttsOutput := cfg.TTSOutput
	// Extract the subfolder and sound name from the custom ID
	parts := strings.SplitN(strings.TrimPrefix(customID, "play_sound_"), "_", 2)
	if len(parts) != 2 {
		logger.ErrorLog.Println("Invalid custom ID format")
		return
	}
	subfolder := parts[0]
	soundName := parts[1]

	// Find the channel that the interaction came from
	c, err := s.State.Channel(i.ChannelID)
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
		if vs.UserID == i.Member.User.ID {
			if customID == "play_sound_temp_tts" {
				content := []discordgo.MessageComponent{}
				row := discordgo.ActionsRow{}
				row.Components = append(row.Components, discordgo.Button{
					Label:    "Stop Sound",
					Style:    discordgo.DangerButton,
					CustomID: "stop_sound",
				})
				content = append(content, row)
				st, err := s.ChannelMessageSendComplex(i.ChannelID, &discordgo.MessageSend{
					Content:    "➡ Text2Speech playing by <@" + i.Member.User.ID + ">",
					Components: content,
				})
				if err != nil {
					logger.ErrorLog.Println("Error sending message:", err)
				}
				logger.InfoLog.Printf("User: %s played sound: %s", i.Member.User.GlobalName, soundName)

				// Play the sound
				err = PlaySound(s, &discordgo.MessageCreate{Message: i.Message}, st, g.ID, vs.ChannelID, ttsOutput, "tts")
				if err != nil {
					logger.ErrorLog.Println("Error playing sound:", err)
				}
				_ = s.ChannelMessageDelete(st.ChannelID, st.ID)
				return
			} else {
				content := []discordgo.MessageComponent{}
				row := discordgo.ActionsRow{}
				row.Components = append(row.Components, discordgo.Button{
					Label:    "Stop Sound",
					Style:    discordgo.DangerButton,
					CustomID: "stop_sound",
				})
				content = append(content, row)
				st, err := s.ChannelMessageSendComplex(i.ChannelID, &discordgo.MessageSend{
					Content:    "➡ Currently Playing by <@" + i.Member.User.ID + ">: " + soundName,
					Components: content,
				})
				if err != nil {
					logger.ErrorLog.Println("Error sending message:", err)
				}
				logger.InfoLog.Printf("User: %s played sound: %s", i.Member.User.GlobalName, soundName)
				// Construct the sound file path
				soundFile := fmt.Sprintf("%s/%s/%s.dca", cfg.SoundsDir, subfolder, soundName)

				// Play the sound
				err = PlaySound(s, &discordgo.MessageCreate{Message: i.Message}, st, g.ID, vs.ChannelID, soundFile, soundName)
				if err != nil {
					logger.ErrorLog.Println("Error playing sound:", err)
				}
				_ = s.ChannelMessageDelete(st.ChannelID, st.ID)
				return
			}

		}

	}

	// If the user is not in a voice channel, send an error message
	logger.InfoLog.Printf("User %s tried to play sound \"%s\" but is not in a voice channel", i.Member.User.GlobalName, soundName)
	_, err = s.ChannelMessageSend(i.ChannelID, "You need to be in a voice channel to play sounds <@"+i.Member.User.ID+">")
	if err != nil {
		logger.ErrorLog.Println("Error sending message:", err)
	}

}

func handleListSoundsInteraction(s *discordgo.Session, i *discordgo.InteractionCreate, customID string) {
	// Extract the category from the custom ID
	category := strings.TrimPrefix(customID, "list_sounds_")
	_, err := s.ChannelMessageSend(i.ChannelID, "➡ Sounds in category - "+category)
	if err != nil {
		logger.ErrorLog.Println("Error sending message:", err)
	}
	// List getSoundsInCategory in the selected category
	sounds, err := getAndSendSoundsInCategory(s, i.ChannelID, category)
	if err != nil {
		logger.ErrorLog.Println("Error listing sounds in category:", err)
		return
	}

	logger.InfoLog.Printf("User: %s listed sounds in category: %s", i.Member.User.GlobalName, category)

	// Split content into multiple messages if it exceeds 5 rows
	for len(sounds) > 0 {
		var messageContent []discordgo.MessageComponent
		if len(sounds) > 5 {
			messageContent, sounds = sounds[:5], sounds[5:]
		} else {
			messageContent, sounds = sounds, nil
		}
		_, err = s.ChannelMessageSendComplex(i.ChannelID, &discordgo.MessageSend{
			Components: messageContent,
		})
		if err != nil {
			logger.ErrorLog.Println("Error sending message:", err)
		}
	}
}

// getSoundsInCategory lists the sounds in the specified category and sends them as buttons
func getAndSendSoundsInCategory(s *discordgo.Session, channelID, category string) ([]discordgo.MessageComponent, error) {
	// Get all sound files in the subfolder
	sounds, err := getSounds(category)
	if err != nil {
		logger.ErrorLog.Println("Error listing sounds in subfolder:", err)
		return nil, err
	}
	if len(sounds) == 0 {
		_, err := s.ChannelMessageSend(channelID, "No sounds found in this category.")
		if err != nil {
			logger.ErrorLog.Println("Error sending message:", err)
		}
		return nil, errors.New("no sounds found in this category")
	}
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

// removeFileExtension removes the file extension from a given file name.
func removeFileExtension(fileName string) string {
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}

// ScanDirectory scans the sound directory and returns a map of folders and files.
func ScanDirectory() (map[string][]string, error) {
	soundsRoot := cfg.SoundsDir
	folderMap := make(map[string][]string)

	err := filepath.WalkDir(soundsRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			// Skip the root folder
			if path == soundsRoot {
				return nil
			}

			// Get relative folder name (e.g., 'folder1/')
			relativeFolder, err := filepath.Rel(soundsRoot, path)
			if err != nil {
				return err
			}

			folderMap[relativeFolder] = []string{} // Initialize an entry for this folder
		} else {
			// Add file to the folder list
			folder := filepath.Dir(path)
			relativeFolder, err := filepath.Rel(soundsRoot, folder)
			if err != nil {
				return err
			}

			// Filter for audio files based on extensions, e.g., ".dca", etc.
			if ext := filepath.Ext(path); ext == ".dca" {
				fileNameWithoutExt := removeFileExtension(filepath.Base(path))
				folderMap[relativeFolder] = append(folderMap[relativeFolder], fileNameWithoutExt)
			}
		}
		return nil
	})

	return folderMap, err
}

// syncDatabaseWithFileSystem will sync the database with the filesystem.
func SyncDatabaseWithFileSystem(folderMap map[string][]string) error {
	existingCategories := fetchCategories() // Get current folders/categories in DB
	existingSounds := fetchSounds()         // Get current files/sounds in DB

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
			soundPath := filepath.Join(cfg.SoundsDir, folder, file+".dca")
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

func fetchSounds() map[int]map[string]string {
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
	alias := removeFileExtension(fileName) // Or any other default value, e.g., ""
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

func addSoundStatistics(soundName string) error {
	_, err := model.Bot.Db.Exec("INSERT OR IGNORE INTO stats_sounds (id) SELECT id FROM sounds WHERE name = ?;", soundName)
	if err != nil {
		return err
	}

	_, err = model.Bot.Db.Exec("UPDATE stats_sounds SET count = count + 1 WHERE id = (SELECT id FROM sounds WHERE alias = ?)", soundName)

	if err != nil {
		return err
	}

	return nil
}

func GetSoundStatistics() (soundStats map[string]int, err error) {
	rows, err := model.Bot.Db.Query("SELECT sounds.alias, count FROM stats_sounds LEFT JOIN sounds ON sounds.id = stats_sounds.id ORDER BY count DESC LIMIT 5")
	if err != nil {
		logger.FatalLog.Fatal(err)
	}
	defer rows.Close()

	soundStats = make(map[string]int)
	for rows.Next() {
		var sound string
		var count int

		err = rows.Scan(&sound, &count)
		if err != nil {
			logger.FatalLog.Fatal(err)
		}
		soundStats[sound] = count
	}

	return soundStats, err
}
