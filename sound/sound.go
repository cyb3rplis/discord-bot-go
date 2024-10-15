package sound

import (
	"database/sql"
	"encoding/binary"
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
	"github.com/cyb3rplis/discord-bot-go/util"

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
		handlePlaySoundInteraction(s, i, customID)
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

func handlePlaySoundInteraction(s *discordgo.Session, i *discordgo.InteractionCreate, customID string) {
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

// InsertCategoriesAndSounds inserts sound categories and sounds into the database
// TODO: remove files from database if no longer exists in folder
func InsertCategoriesAndSounds() error {
	soundFolders, err := util.WalkSoundFolder()
	if err != nil {
		return fmt.Errorf("failed to walk sound folder: %w", err)
	}

	for _, folder := range soundFolders {
		// Insert category
		categoryID, categoryName, err := checkExistingCategory(folder)
		if err != nil {
			return fmt.Errorf("error checking existing category: %w", err)
		}
		// If the category does not exist, insert it
		if categoryName == "" {
			categoryID, err = insertCategory(folder)
			if err != nil {
				return fmt.Errorf("failed to insert category: %w", err)
			}
		}

		// Insert sounds
		soundFiles, err := util.WalkSoundFiles(folder)
		if err != nil {
			return fmt.Errorf("failed to walk sound files: %w", err)
		}
		// iterate sound files and insert the file as a blob into the database
		for _, soundFile := range soundFiles {
			soundPath := filepath.Join(cfg.SoundsDir, folder, soundFile)
			fileData, err := os.ReadFile(soundPath)
			if err != nil {
				return fmt.Errorf("failed to read sound file: %w", err)
			}
			// Check if file already exists
			existingSound, err := checkExistingSound(soundFile)
			if err != nil {
				return fmt.Errorf("error checking existing sound: %w", err)
			}
			soundName := strings.TrimSuffix(soundFile, ".dca")
			// If the sound does not exist, insert it
			if existingSound == "" {
				err = insertSound(soundFile, soundName, categoryID, fileData)
				if err != nil {
					return fmt.Errorf("error inserting sound: %w", err)
				}
			}
		}
	}
	return nil
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

// checkExistingCategory checks if a category already exists in the database
func checkExistingCategory(folder string) (int64, string, error) {
	var categoryID int64
	var categoryName string
	err := model.Bot.Db.QueryRow(`SELECT id, name FROM categories WHERE name = ?`, folder).Scan(&categoryID, &categoryName)
	if err != nil && err != sql.ErrNoRows {
		return 0, "", err
	}
	return categoryID, categoryName, nil
}

// checkExistingSound checks if a sound already exists in the database
func checkExistingSound(soundFile string) (string, error) {
	var fileName string
	err := model.Bot.Db.QueryRow("SELECT name FROM sounds WHERE name = ?", soundFile).Scan(&fileName)
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}
	return fileName, nil
}

// insertCategory inserts a category into the database
func insertCategory(folder string) (int64, error) {
	res, err := model.Bot.Db.Exec("INSERT INTO categories (name) VALUES (?)", folder)
	if err != nil {
		return 0, err
	}
	categoryID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}
	return categoryID, nil
}

// deleteCategory removes a category from the database
func deleteCategory(name string) error {
	_, err := model.Bot.Db.Exec("DELETE FROM categories WHERE name = ?", name)
	return err
}

// insertSound inserts a sound into the database
func insertSound(soundFile, soundName string, categoryID int64, fileData []byte) error {
	_, err := model.Bot.Db.Exec("INSERT INTO sounds (name, alias, category_id, file) VALUES (?, ?, ?, ?)", soundFile, soundName, categoryID, fileData)
	if err != nil {
		return err
	}
	return nil
}

// deleteSound removes a sound from the database
func deleteSound(id int) error {
	_, err := model.Bot.Db.Exec("DELETE FROM sounds WHERE id = ?", id)
	return err
}
