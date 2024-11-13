package model

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/dlog"
)

// ScanDirectory scans the sound directory and returns a map of folders and files.
func (m *Model) ScanDirectory() (map[string][]string, error) {
	soundsRoot := m.Config.SoundsDir
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

			// Filter for audio files based on extensions, e.g., ".mp3", etc.
			if ext := filepath.Ext(path); ext == ".mp3" {
				fileNameWithoutExt := RemoveFileExtension(filepath.Base(path))
				folderMap[relativeFolder] = append(folderMap[relativeFolder], fileNameWithoutExt)
			}
		}
		return nil
	})

	return folderMap, err
}

// RemoveFileExtension removes the file extension from a given file name.
func RemoveFileExtension(fileName string) string {
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}

func SortMapByValue(m map[string]int) map[string]int {
	var keys []string
	var sortedM = make(map[string]int)
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return m[keys[i]] > m[keys[j]]
	})
	for _, k := range keys {
		sortedM[k] = m[k]
	}
	return sortedM
}

// SortMapKeysByValue sorts a map by its values and returns the keys in descending order
func SortMapKeysByValue(m map[string]int) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return m[keys[i]] > m[keys[j]]
	})
	return keys
}

// BuildSoundButtons creates a list of buttons for the provided category
func BuildSoundButtons(sounds []string, category string, buttonStyle discordgo.ButtonStyle) []discordgo.MessageComponent {
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
			Style:    buttonStyle,
			CustomID: fmt.Sprintf("play_sound_%s_%s", category, soundName),
		})
	}
	// Append the last row if it has any components
	if len(row.Components) > 0 {
		content = append(content, row)
	}
	return content
}

// BuildSingleSoundButton creates a single button
func BuildSingleSoundButton(soundName, category string, buttonStyle discordgo.ButtonStyle) []discordgo.MessageComponent {
	content := []discordgo.MessageComponent{}
	button := discordgo.Button{
		Label:    soundName,
		Style:    buttonStyle,
		CustomID: fmt.Sprintf("play_sound_%s_%s", category, soundName),
	}

	actionRow := discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{button},
	}

	content = append(content, actionRow)

	return content
}

// BuildListButtons creates a list of buttons for the provided categories
func BuildListButtons(categories []string, buttonStyle discordgo.ButtonStyle) []discordgo.MessageComponent {
	content := []discordgo.MessageComponent{}
	row := discordgo.ActionsRow{}
	for i, category := range categories {
		// only 5 buttons per row - discord does not allow more
		if i > 0 && i%5 == 0 {
			content = append(content, row)
			row = discordgo.ActionsRow{}
		}
		row.Components = append(row.Components, discordgo.Button{
			Label:    category,
			Style:    buttonStyle,
			CustomID: fmt.Sprintf("list_sounds_%s", category),
		})
	}
	// Append the last row if it has any components
	if len(row.Components) > 0 {
		content = append(content, row)
	}
	return content
}

// BuildMessages creates a list of messages for the provided buttons
func BuildMessages(buttons []discordgo.MessageComponent, initialMessage *discordgo.MessageSend) []*discordgo.MessageSend {
	var messages []*discordgo.MessageSend
	if initialMessage != nil {
		messages = append(messages, initialMessage)
	}
	for len(buttons) > 0 {
		var messageContent []discordgo.MessageComponent
		if len(buttons) > 5 {
			messageContent, buttons = buttons[:5], buttons[5:]
		} else {
			messageContent, buttons = buttons, nil
		}
		message := &discordgo.MessageSend{
			Components: messageContent,
		}
		messages = append(messages, message)
	}
	return messages
}

// ComputeFileHash computes the SHA-256 hash of a given file
func ComputeFileHash(filePath string) (string, error) {
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

func FileExistsInDB(existingSounds map[int]map[string]string, categoryID int, fileName string, fileHash string) bool {
	if soundsInCategory, exists := existingSounds[categoryID]; exists {
		if dbHash, fileExists := soundsInCategory[fileName]; fileExists {
			// Check if the hash matches
			return dbHash == fileHash
		}
	}
	return false
}

func FileExistsInFS(fsFiles []string, fileName string) bool {
	for _, fsFile := range fsFiles {
		if fsFile == fileName {
			return true
		}
	}
	return false
}

func (m *Model) InactiveLeaveVoiceChannel(s *discordgo.Session) error {
	botInVC := false
	vc := s.VoiceConnections[Meta.Guild.ID]

	// Check if the bot is in a voice channel
	if vc != nil {
		botInVC = true
	}

	// if the bot is in a voice channel and the bot is inactive, leave the voice channel
	if botInVC {
		err := vc.Disconnect()
		if err != nil {
			dlog.ErrorLog.Println("error closing voice connection:", err)
			return err
		}

		dlog.InfoLog.Printf("Bot is inactive since a prolonged period, leaving voice channel")
	}

	return nil
}
