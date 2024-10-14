package sound

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/config"
)

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

// LoadSound attempts to load an encoded sound file from disk.
func LoadSound(soundName string) error {
	var opusLen int16
	file, err := os.Open(soundName)
	if err != nil {
		log.Println("error opening dca file :", err)
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
			log.Println("error reading from dca file :", err)
			return err
		}

		// Read encoded pcm from dca file.
		InBuf := make([]byte, opusLen)
		err = binary.Read(file, binary.LittleEndian, &InBuf)
		if err != nil {
			log.Println("error reading from dca file :", err)
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
		s.ChannelMessageDelete(lastChannelID, lastMessageID)
		lastChannelID = st.ChannelID
		lastMessageID = st.ID
		stopChannel <- struct{}{}
		time.Sleep(150 * time.Millisecond) // Give some time for the current sound to stop
	}

	// Load the sound file.
	err = LoadSound(soundFile)
	if err != nil {
		log.Printf("error loading sound %s, %v ", soundFile, err)
		_, err = s.ChannelMessageSend(m.ChannelID, "> Sound does not exist\n> Sikerim")
		if err != nil {
			log.Println("error loading sound:", err)
		}
		return
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
	s.ChannelMessageDelete(st.ChannelID, st.ID)
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
		log.Println("Failed to respond to interaction:", err)
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
		_, err := s.ChannelMessageSend(i.ChannelID, "Stop spamming the buttons ➡ "+strings.ToUpper(i.Member.User.GlobalName)+" ⬅ you fucking idiot!!!")
		if err != nil {
			log.Println("error sending message:", err)
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
		log.Println("unknown interaction:", customID)
	}
}

func handleStopSoundInteraction(s *discordgo.Session) {
	// check if the bot is currently speaking, and exit
	if botSpeaking {
		stopChannel <- struct{}{}

		// Delete the last "Now Playing"
		err := s.ChannelMessageDelete(lastChannelID, lastMessageID)
		if err != nil {
			return
		}
		lastChannelID = ""
		lastMessageID = ""
		time.Sleep(150 * time.Millisecond) // Give some time for the current sound to stop
	}
}

func handlePlaySoundInteraction(s *discordgo.Session, i *discordgo.InteractionCreate, customID string) {
	cfg := config.GetConfig()

	// Extract the subfolder and sound name from the custom ID
	parts := strings.SplitN(strings.TrimPrefix(customID, "play_sound_"), "_", 2)
	if len(parts) != 2 {
		log.Println("Invalid custom ID format")
		return
	}
	subfolder := parts[0]
	soundName := parts[1]

	// Find the channel that the interaction came from
	c, err := s.State.Channel(i.ChannelID)
	if err != nil {
		log.Println("error finding channel:", err)
		return
	}

	// Find the guild for that channel
	g, err := s.State.Guild(c.GuildID)
	if err != nil {
		log.Println("error finding guild:", err)
		return
	}

	content := []discordgo.MessageComponent{}
	row := discordgo.ActionsRow{}
	row.Components = append(row.Components, discordgo.Button{
		Label:    "Stop Sound",
		Style:    discordgo.DangerButton,
		CustomID: "stop_sound",
	})
	content = append(content, row)
	st, err := s.ChannelMessageSendComplex(i.ChannelID, &discordgo.MessageSend{
		Content:    "➡ Currently Playing by " + i.Member.User.GlobalName + ": " + soundName,
		Components: content,
	})
	if err != nil {
		log.Println("error sending message:", err)
	}
	log.Printf("User: %s played sound: %s", i.Member.User.GlobalName, soundName)

	// Look for the interaction user in that guild's current voice states
	for _, vs := range g.VoiceStates {
		if vs.UserID == i.Member.User.ID {
			// Construct the sound file path
			soundFile := fmt.Sprintf("%s/%s/%s.dca", cfg.SoundsDir, subfolder, soundName)
			// Play the sound
			err = PlaySound(s, &discordgo.MessageCreate{Message: i.Message}, st, g.ID, vs.ChannelID, soundFile, soundName)
			if err != nil {
				log.Println("error playing sound:", err)
			}

			s.ChannelMessageDelete(st.ChannelID, st.ID)
			return
		}
	}
}

func handleListSoundsInteraction(s *discordgo.Session, i *discordgo.InteractionCreate, customID string) {
	// Extract the category from the custom ID
	category := strings.TrimPrefix(customID, "list_sounds_")
	_, err := s.ChannelMessageSend(i.ChannelID, "➡ Sounds in category - "+category)
	if err != nil {
		log.Println("error sending message:", err)
	}
	// List getSoundsInCategory in the selected category
	sounds, err := getSoundsInCategory(s, i.ChannelID, category)
	if err != nil {
		log.Println("error listing sounds in category:", err)
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
			log.Println("error sending message:", err)
		}
	}
}

// getSoundsInCategory lists the sounds in the specified category and sends them as buttons
func getSoundsInCategory(s *discordgo.Session, channelID, category string) ([]discordgo.MessageComponent, error) {
	// Get all sound files in the subfolder
	sounds, err := WalkSoundFiles(category)
	if err != nil {
		log.Println("error listing sounds in subfolder:", err)
		return nil, err
	}
	if len(sounds) == 0 {
		_, err := s.ChannelMessageSend(channelID, "No sounds found in this category.")
		if err != nil {
			log.Println("error sending message:", err)
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

// WalkSoundFiles returns a list of sound files in a subfolder.
func WalkSoundFiles(subfolder string) ([]string, error) {
	cfg := config.GetConfig()
	baseDir := cfg.SoundsDir
	subfolderPath := filepath.Join(baseDir, subfolder)
	cleanedSubfolderPath := filepath.Clean(subfolderPath)
	// Ensure the cleaned subfolder path is within the base directory
	if !strings.HasPrefix(cleanedSubfolderPath, baseDir) {
		return nil, errors.New("potential path traversal detected")
	}
	files, err := os.ReadDir(cleanedSubfolderPath)
	if err != nil {
		return nil, err
	}
	var soundFiles []string
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".dca" {
			soundFiles = append(soundFiles, file.Name())
		}
	}
	return soundFiles, nil
}

// WalkSoundFolder returns a list of subfolders in the sound folder.
func WalkSoundFolder() ([]string, error) {
	cfg := config.GetConfig()
	soundFolderDir := cfg.SoundsDir
	cleanedSubfolderPath := filepath.Clean(soundFolderDir)
	folders, err := os.ReadDir(cleanedSubfolderPath)
	if err != nil {
		return nil, err
	}

	var subfolders []string
	for _, entry := range folders {
		if entry.IsDir() {
			subfolders = append(subfolders, entry.Name())
		}
	}

	return subfolders, nil
}
