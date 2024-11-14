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

func (m *Model) LeaveVoiceChannel(s *discordgo.Session) {
	// Get the voice connection for the guild
	vc := s.VoiceConnections[Meta.Guild.ID]
	if vc == nil {
		// Bot is not in a voice channel
		return
	}

	// Check if the bot is currently connected to a voice channel in the guild
	guild, err := s.State.Guild(Meta.Guild.ID)
	if err != nil {
		dlog.ErrorLog.Println("error finding guild:", err)
		return
	}

	botInVC := false
	for _, vs := range guild.VoiceStates {
		if vs.UserID == s.State.User.ID {
			botInVC = true
			break
		}
	}

	// If the bot is in a voice channel, attempt to disconnect
	if botInVC {
		err := vc.Disconnect()
		if err != nil {
			dlog.ErrorLog.Println("error disconnecting from voice channel:", err)
			return
		}
		dlog.InfoLog.Println("Bot successfully left the voice channel after inactivity.")
	}
}

// getChannelByName returns a channel by its name
func getChannelByName(s *discordgo.Session, channelName string) (*discordgo.Channel, error) {
	guild, err := s.State.Guild(Meta.Guild.ID)
	if err != nil {
		dlog.FatalLog.Fatalf("Failed to get guild: %v", err)
	}

	channels, err := s.GuildChannels(guild.ID)
	if err != nil {
		return nil, err
	}

	// Loop through channels to find one that matches the name
	for _, channel := range channels {
		if channel.Name == channelName {
			return channel, nil
		}
	}

	// Return an error if no channel was found with the specified name
	return nil, fmt.Errorf("channel with name %s not found", channelName)
}

// PinNewSoundButtons pins the new sound buttons to the bot channel
func (m *Model) PinNewSoundButtons(s *discordgo.Session) {
	channel, err := getChannelByName(s, m.Config.BotChannel)
	if err != nil {
		dlog.ErrorLog.Printf("Failed to get channel by name: %v", err)
		return
	}

	pinnedMessages, err := s.ChannelMessagesPinned(channel.ID)
	if err != nil {
		dlog.ErrorLog.Printf("Failed to get pinned messages: %v", err)
		return
	}

	// we only want to have messages from the bot
	botMessages := []*discordgo.Message{}
	for _, message := range pinnedMessages {
		if message.Author.ID == s.State.User.ID {
			botMessages = append(botMessages, message)
		}
	}

	newSounds, err := m.GetNewSounds()
	if err != nil {
		dlog.ErrorLog.Printf("Failed to get new sounds: %v", err)
		return
	}

	// if there are no new sounds, we also want to delete the old pinned messages
	// that means there are no new sounds within the last 24 hours, so they are considered old and dont need a pinned message
	if len(newSounds) == 0 {
		for _, message := range botMessages {
			err = s.ChannelMessageUnpin(channel.ID, message.ID)
			if err != nil {
				dlog.ErrorLog.Printf("Failed to unpin message: %v", err)
			}
		}
		return
	}

	if ok := m.CompareNewSoundsWithPinnedSounds(newSounds, botMessages); !ok {
		// sounds didnt change, just return
		return
	}

	for _, message := range botMessages {
		err = s.ChannelMessageUnpin(channel.ID, message.ID)
		if err != nil {
			dlog.ErrorLog.Printf("Failed to unpin message: %v", err)
			return
		}
	}

	message := fmt.Sprintf("Here are the %d newest sounds:", len(newSounds))
	_, err = s.ChannelMessageSend(channel.ID, message)
	if err != nil {
		dlog.ErrorLog.Printf("error sending new sounds message: %v", err)
		return
	}

	buttons := BuildSoundButtons(newSounds, "new", discordgo.SuccessButton)
	buttonMessage := &discordgo.MessageSend{
		Components: buttons,
	}
	st, err := s.ChannelMessageSendComplex(channel.ID, buttonMessage)
	if err != nil {
		dlog.ErrorLog.Printf("error sending new sounds message: %v", err)
		return
	}

	err = s.ChannelMessagePin(channel.ID, st.ID)
	if err != nil {
		dlog.ErrorLog.Printf("Failed to pin message: %v", err)
		return
	}
}

// CompareNewSoundsWithPinnedSounds compares the new sounds with the pinned sounds and returns false if they are equal
func (m *Model) CompareNewSoundsWithPinnedSounds(sounds []string, pinnedMessages []*discordgo.Message) bool {
	var pinnedSounds []string

	for _, message := range pinnedMessages {
		for _, component := range message.Components {
			actionRow, ok := component.(*discordgo.ActionsRow)
			if !ok {
				continue
			}

			for _, item := range actionRow.Components {
				button, ok := item.(*discordgo.Button)
				if !ok {
					continue
				}

				pinnedSounds = append(pinnedSounds, button.Label)
			}
		}
	}

	sort.Strings(pinnedSounds)
	sort.Strings(sounds)

	// compare if both slices are equal
	if len(pinnedSounds) == len(sounds) {
		equal := true
		for i := range pinnedSounds {
			if pinnedSounds[i] != sounds[i] {
				equal = false
				break
			}
		}
		if equal {
			return false
		}
	}

	return true
}
