package sound

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
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

// loadSound attempts to load an encoded sound file from disk.
func LoadSound(soundName string) error {
	var opusLen int16
	file, err := os.Open(soundName)
	if err != nil {
		fmt.Println("error opening dca file :", err)
		return err
	}
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

		// Should not be any end of file errors
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
		_, err := s.ChannelMessageSend(m.ChannelID, "🎉🎉 bot is already playing a sound, please try again later 🎉🎉")
		if err != nil {
			fmt.Println("error sending message:", err)
		}
		return nil
	}

	// Load the sound file.
	err = LoadSound(soundFile)
	if err != nil {
		fmt.Printf("error loading sound %s, %v ", soundFile, err)
		_, err = s.ChannelMessageSend(m.ChannelID, "> Sound does not exist\n> Use .list to show all categories")
		if err != nil {
			fmt.Println("error loading sound:", err)
		}
		return
	}

	fmt.Println("playing sound file: ", soundName)

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

func ListSoundsCategories() ([]string, error) {
	files, err := ioutil.ReadDir(config.GetValueString("general", "sounds_dir", "-"))
	if err != nil {
		return nil, err
	}
	var subfolders []string
	for _, file := range files {
		if file.IsDir() {
			subfolders = append(subfolders, file.Name())
		}
	}
	return subfolders, nil
}

func ListSoundsInSubfolder(subfolder string) ([]string, error) {
	baseDir := config.GetValueString("general", "sounds_dir", "-")
	subfolderPath := filepath.Join(baseDir, subfolder)
	cleanedSubfolderPath := filepath.Clean(subfolderPath)
	// Ensure the cleaned subfolder path is within the base directory
	if !strings.HasPrefix(cleanedSubfolderPath, baseDir) {
		return nil, errors.New("potential path traversal detected")
	}
	files, err := ioutil.ReadDir(cleanedSubfolderPath)
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

// Interaction create event handler (for button clicks)
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

				// Acknowledge the interaction
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
				subfolder := parts[0]
				soundName := parts[1]

				// Find the channel that the interaction came from
				c, err := s.State.Channel(i.ChannelID)
				if err != nil {
					// could not find channel
					return
				}

				// Find the guild for that channel
				g, err := s.State.Guild(c.GuildID)
				if err != nil {
					// could not find guild
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
}

// helper
func isSymlink(path string) bool {
	fileInfo, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeSymlink) != 0
}
