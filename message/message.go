package message

import (
	"fmt"
	"github.com/cyb3rplis/discord-bot-go/util"
	"log"
	"math/rand"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/config"
)

// AudioMessageHandler is created on any channel that the authenticated bot has access to.
func AudioMessageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	cfg := config.GetConfig()
	prefix := cfg.Prefix

	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}
	if len(m.Content) == 0 { // Ignore empty messages
		log.Println("Empty content..")
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
			log.Println("error sending message:", err)
		}
	// LIST CATEGORIES
	case command == fmt.Sprintf("%slist", prefix):
		// Get all sound folders to use for later
		soundFolders, err := util.WalkSoundFolder()
		if err != nil {
			log.Println("Error getting sound subfolders")
			return
		}
		if len(soundFolders) == 0 {
			_, err := s.ChannelMessageSend(m.ChannelID, "No sound categories found.")
			if err != nil {
				log.Println("error sending message:", err)
			}
			return
		}
		content := []discordgo.MessageComponent{}
		row := discordgo.ActionsRow{}
		for i, folder := range soundFolders {
			// only 5 buttons per row - discord does not allow more
			if i > 0 && i%5 == 0 {
				content = append(content, row)
				row = discordgo.ActionsRow{}
			}
			row.Components = append(row.Components, discordgo.Button{
				Label:    folder,
				Style:    discordgo.PrimaryButton,
				CustomID: fmt.Sprintf("list_sounds_%s", folder),
			})
		}
		// Append the last row if it has any components
		if len(row.Components) > 0 {
			content = append(content, row)
		}
		_, err = s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
			Content:    "➡ Click on a category [blue button]",
			Components: content,
		})
		if err != nil {
			log.Println("error sending message:", err)
		}
	case strings.Contains(strings.ToLower(m.Content), "mutter"):
		_, err := s.ChannelMessageSend(m.ChannelID, mutterWitze[rand.Intn(len(mutterWitze))])
		if err != nil {
			log.Println("error sending message:", err)
		}
	default:
		return
	}
}
