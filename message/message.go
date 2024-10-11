package message

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/config"
	"github.com/cyb3rplis/discord-bot-go/sound"
)

// AudioMessageHandler is created on any channel that the authenticated bot has access to.
func AudioMessageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	prefix := config.GetValueString("general", "prefix", ".")

	// Get all sound folders to use for later
	soundFolders, err := sound.WalkSoundFolder()
	if err != nil {
		fmt.Println("Error getting sound subfolders")
		return
	}
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}
	if len(m.Content) == 0 { // Ignore empty messages
		fmt.Println("Empty content..")
		return
	}
	// Extract the command and arguments
	args := strings.Split(m.Content, " ")
	command := args[0]

	switch {
	//if the command starts with the prefix and is not a list or stop command
	case command == fmt.Sprintf("%shelp", prefix):
		// Default case: show help message
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("🧐 Usage: \n > » List Categories: <%slist> \n", prefix))
		if err != nil {
			fmt.Println("error sending message:", err)
		}
	// LIST CATEGORIES
	case command == fmt.Sprintf("%slist", prefix):
		if len(soundFolders) == 0 {
			_, err := s.ChannelMessageSend(m.ChannelID, "No sound categories found.")
			if err != nil {
				fmt.Println("error sending message:", err)
			}
			return
		}
		list := "> Available Sound Categories \n"
		for _, folder := range soundFolders {
			list += fmt.Sprintf("> * %s\n", folder)
		}
		_, err = s.ChannelMessageSend(m.ChannelID, list)
		if err != nil {
			fmt.Println("error sending message:", err)
		}

	// LIST SOUNDS IN SUBFOLDER AND CREATE BUTTONS
	case len(args) == 1 && len(soundFolders) >= 1:
		category := strings.TrimPrefix(command, prefix)
		noFolder := true
		for _, folder := range soundFolders {
			if folder == category {
				noFolder = false
				break
			}
		}
		if noFolder {
			return
		}
		// Get all sound files in the subfolder
		sounds, err := sound.WalkSoundFiles(category)
		if err != nil {
			fmt.Println("error listing sounds in subfolder:", err)
			return
		}
		if len(sounds) == 0 {
			_, err := s.ChannelMessageSend(m.ChannelID, "No sounds found in this category.")
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
			_, err = s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
				Components: messageContent,
			})
			if err != nil {
				fmt.Println("error sending message:", err)
			}
		}
	default:
		return
	}
}
