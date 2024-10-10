package message

import (
	"fmt"
	"strings"

	"github.com/cyb3rplis/discord-bot-go/config"

	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/sound"
)

// message is created on any channel that the autenticated bot has access to.
func AudioMessageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	prefix := config.GetValueString("general", "prefix", ".")

	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	if len(m.Content) == 0 {
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
		_, err := s.ChannelMessageSend(m.ChannelID, "🧐 Usage: \n > » List Categories: <.list> \n > » List Sounds: <.category_name> <.sound_name> \n > » Play Sound: <.category> <sound_name>")
		if err != nil {
			fmt.Println("error sending message:", err)
		}
	// LIST CATEGORIES
	case command == fmt.Sprintf("%slist", prefix):
		soundCategories, err := sound.ListSoundsCategories()
		if err != nil {
			fmt.Println("error listing sound categories:", err)
			return
		}
		if len(soundCategories) == 0 {
			_, err := s.ChannelMessageSend(m.ChannelID, "No sound categories found.")
			if err != nil {
				fmt.Println("error sending message:", err)
			}
			return
		}
		list := fmt.Sprintf("> Available Sound Categories \n")
		for _, folder := range soundCategories {
			list += fmt.Sprintf("> * %s\n", folder)
		}
		_, err = s.ChannelMessageSend(m.ChannelID, list)
		if err != nil {
			fmt.Println("error sending message:", err)
		}
	// LIST SOUNDS IN SUBFOLDER
	case len(args) == 1 && strings.HasPrefix(command, prefix):
		subfolder := strings.TrimPrefix(command, prefix)
		sounds, err := sound.ListSoundsInSubfolder(subfolder)
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
			// only 5 buttons per row
			if i > 0 && i%5 == 0 {
				content = append(content, row)
				row = discordgo.ActionsRow{}
			}
			row.Components = append(row.Components, discordgo.Button{
				Label:    soundName,
				Style:    discordgo.SecondaryButton,
				CustomID: fmt.Sprintf("play_sound_%s_%s", subfolder, soundName),
			})
		}
		// Append the last row if it has any components
		if len(row.Components) > 0 {
			content = append(content, row)
		}

		_, err = s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
			Components: content,
		})
		if err != nil {
			fmt.Println("error sending message:", err)
		}
	default:
		return
	}
}
