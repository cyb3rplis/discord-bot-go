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

// playSound plays the current buffer to the provided channel.
func PlaySound(s *discordgo.Session, m *discordgo.MessageCreate, guildID, channelID, soundFile, soundName string) (err error) {

	// check if the bot is currently speaking, and exit early to avoid corrupted sound buffer
	if botSpeaking {
		_, err := s.ChannelMessageSend(m.ChannelID, "🎉🎉 bot is already playing a sound, please try again later 🎉🎉")
		if err != nil {
			fmt.Println("error sending message:", err)
		}
		return nil
	}

	_, err = s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
		Content: fmt.Sprintf("Current sound is playing: %s", soundName),
		Components: []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label:    "Stop Sound",
						Style:    discordgo.PrimaryButton,
						CustomID: "stop_sound", // Unique identifier for the button
					},
				},
			},
		},
	})
	if err != nil {
		fmt.Println("error sending message:", err)
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

func ListSounds() ([]string, error) {
	files, err := ioutil.ReadDir(config.GetValueString("general", "sounds_dir", "-"))
	if err != nil {
		return nil, err
	}
	var soundFiles []string
	baseDir, err := filepath.Abs(config.GetValueString("general", "sounds_dir", "-"))
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		// Only include .dca files
		if filepath.Ext(file.Name()) == ".dca" {
			// resolve the absolute path of each file
			filePath, err := filepath.Abs(filepath.Join(config.GetValueString("general", "sounds_dir", "-"), file.Name()))
			if err != nil {
				return nil, err
			}
			// ensure the file is inside the base directory
			if !strings.HasPrefix(filePath, baseDir) {
				return nil, errors.New("potential path traversal detected")
			}
			// only append if not a symlink
			if !isSymlink(filePath) {
				soundFiles = append(soundFiles, file.Name())
			}
		}
	}
	return soundFiles, nil
}

// InteractionCreate create event handler (for button clicks)
func InteractionCreate(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Check if the interaction is a button click
	if i.Type == discordgo.InteractionMessageComponent {
		switch i.MessageComponentData().CustomID {
		case "stop_sound":
			// Handle the button press, stop the song logic here
			// For example, stop the song in your music player
			// Then, send a response to the interaction to confirm the action
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
