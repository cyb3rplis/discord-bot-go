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
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/config"
)

var buffer = make([][]byte, 0)
var botSpeaking = false
var stopChannel = make(chan struct{})

// LoadSound attempts to load an encoded sound file from disk.
func LoadSound(soundName string) error {
	var opusLen int16
	file, err := os.Open(soundName)
	if err != nil {
		fmt.Println("error opening dca file :", err)
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
			fmt.Println("error reading from dca file :", err)
			return err
		}

		// Read encoded pcm from dca file.
		InBuf := make([]byte, opusLen)
		err = binary.Read(file, binary.LittleEndian, &InBuf)
		if err != nil {
			fmt.Println("error reading from dca file :", err)
			return err
		}

		// Append encoded pcm data to the buffer.
		buffer = append(buffer, InBuf)
	}
}

// PlaySound plays the current buffer to the provided channel.
func PlaySound(s *discordgo.Session, m *discordgo.MessageCreate, guildID, channelID, soundFile, soundName string) (err error) {

	// check if the bot is currently speaking, and exit early to avoid corrupted sound buffer
	if botSpeaking {
		stopChannel <- struct{}{}
		time.Sleep(250 * time.Millisecond) // Give some time for the current sound to stop
	}

	// Load the sound file.
	err = LoadSound(soundFile)
	if err != nil {
		fmt.Printf("error loading sound %s, %v ", soundFile, err)
		_, err = s.ChannelMessageSend(m.ChannelID, "> Sound does not exist\n> Sikerim")
		if err != nil {
			fmt.Println("error loading sound:", err)
		}
		return
	}

	fmt.Println("> playing sound file: ", soundName)

	// Join the provided voice channel.
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return err
	}
	botSpeaking = true

	// Sleep for a specified amount of time before playing the sound
	time.Sleep(250 * time.Millisecond)

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

	return nil
}

// InteractionHandler create event handler (for button clicks)
func InteractionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Check if the interaction is a button click
	if i.Type == discordgo.InteractionMessageComponent {
		switch {
		case i.MessageComponentData().CustomID == "stop_sound":
			// Handle the stop sound button press
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseUpdateMessage,
				Data: &discordgo.InteractionResponseData{
					Content:    "The sound has been stopped.",
					Components: []discordgo.MessageComponent{}, // Remove the button after it's pressed
				},
			})
			if err != nil {
				log.Println("Failed to respond to interaction:", err)
			}
			// Send stop signal to stopChannel
			stopChannel <- struct{}{}

		default:
			if strings.HasPrefix(i.MessageComponentData().CustomID, "play_sound_") {

				// Acknowledge the interaction (without this the interaction will be marked as failed)
				err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseDeferredMessageUpdate,
				})
				if err != nil {
					log.Println("Failed to respond to interaction:", err)
					return
				}

				// extract the subfolder and sound name from the custom ID
				parts := strings.SplitN(strings.TrimPrefix(i.MessageComponentData().CustomID, "play_sound_"), "_", 2)
				if len(parts) != 2 {
					fmt.Println("Invalid custom ID format")
					return
				}
				// Get the subfolder and sound name from the custom ID
				subfolder := parts[0]
				soundName := parts[1]

				// Find the channel that the interaction came from
				c, err := s.State.Channel(i.ChannelID)
				if err != nil {
					fmt.Println("error finding channel:", err)
					return
				}

				// Find the guild for that channel
				g, err := s.State.Guild(c.GuildID)
				if err != nil {
					fmt.Println("error finding guild:", err)
					return
				}

				// Look for the interaction user in that guild's current voice states
				for _, vs := range g.VoiceStates {
					if vs.UserID == i.Member.User.ID {
						// Construct the sound file path
						soundFile := fmt.Sprintf("%s/%s/%s.dca", config.GetValueString("general", "sounds_dir", "-"), subfolder, soundName)
						// Play the sound
						err = PlaySound(s, &discordgo.MessageCreate{Message: i.Message}, g.ID, vs.ChannelID, soundFile, soundName)
						if err != nil {
							fmt.Println("error playing sound:", err)
						}
						return
					}
				}
			}
		}
	}
	// Check if the interaction is a button click
	if i.Type == discordgo.InteractionMessageComponent {
		switch {
		case strings.HasPrefix(i.MessageComponentData().CustomID, "list_sounds_"):
			// Acknowledge the interaction
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseDeferredMessageUpdate,
			})
			if err != nil {
				log.Println("Failed to respond to interaction:", err)
				return
			}

			// Extract the category from the custom ID
			category := strings.TrimPrefix(i.MessageComponentData().CustomID, "list_sounds_")

			// List sounds in the selected category
			listSoundsInCategory(s, i.ChannelID, category)
		}
	}
}

// listSoundsInCategory lists the sounds in the specified category and sends them as buttons
func listSoundsInCategory(s *discordgo.Session, channelID, category string) {
	// Get all sound files in the subfolder
	sounds, err := WalkSoundFiles(category)
	if err != nil {
		fmt.Println("error listing sounds in subfolder:", err)
		return
	}
	if len(sounds) == 0 {
		_, err := s.ChannelMessageSend(channelID, "No sounds found in this category.")
		if err != nil {
			fmt.Println("error sending message:", err)
		}
		return
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

	// Split content into multiple messages if it exceeds 5 rows
	for len(content) > 0 {
		var messageContent []discordgo.MessageComponent
		if len(content) > 5 {
			messageContent, content = content[:5], content[5:]
		} else {
			messageContent, content = content, nil
		}
		_, err = s.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
			Components: messageContent,
		})
		if err != nil {
			fmt.Println("error sending message:", err)
		}
	}
}

// WalkSoundFiles returns a list of sound files in a subfolder.
func WalkSoundFiles(subfolder string) ([]string, error) {
	baseDir := config.GetValueString("general", "sounds_dir", "-")
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
	soundFolderDir := config.GetValueString("general", "sounds_dir", "-")
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
