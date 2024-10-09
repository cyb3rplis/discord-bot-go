package sound

import (
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/config"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"
)

var buffer = make([][]byte, 0)

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
func PlaySound(s *discordgo.Session, guildID, channelID, soundName string) (err error) {

	// Load the sound file.
	err = LoadSound(soundName)
	if err != nil {
		fmt.Printf("error loading sound %s, %v ", soundName, err)
		return
	}

	// Join the provided voice channel.
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return err
	}

	// Sleep for a specified amount of time before playing the sound
	time.Sleep(250 * time.Millisecond)

	// Start speaking.
	vc.Speaking(true)

	// Send the buffer data.
	for _, buff := range buffer {
		vc.OpusSend <- buff
	}

	// Stop speaking
	vc.Speaking(false)

	// Sleep for a specificed amount of time before ending.
	time.Sleep(250 * time.Millisecond)

	// Disconnect from the provided voice channel.
	//vc.Disconnect()

	return nil
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

// helper
func isSymlink(path string) bool {
	fileInfo, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeSymlink) != 0
}
