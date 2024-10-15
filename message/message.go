package message

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/cyb3rplis/discord-bot-go/logger"
	"github.com/cyb3rplis/discord-bot-go/sound"

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
		logger.InfoLog.Println("Empty content in command, ignore")
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
			logger.ErrorLog.Println("error sending message:", err)
		}
	// LIST CATEGORIES
	case command == fmt.Sprintf("%slist", prefix):
		// Get all sound folders to use for later
		//get categories from database
		categories, err := sound.GetCategories()
		if err != nil {
			logger.ErrorLog.Println("error getting categories:", err)
		}
		if len(categories) == 0 {
			_, err := s.ChannelMessageSend(m.ChannelID, "No sound categories found.")
			if err != nil {
				logger.ErrorLog.Println("error sending message:", err)
			}
			return
		}
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
				Style:    discordgo.PrimaryButton,
				CustomID: fmt.Sprintf("list_sounds_%s", category),
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
			logger.ErrorLog.Println("error sending message:", err)
		}
	case strings.Contains(strings.ToLower(m.Content), "mutter"):
		_, err := s.ChannelMessageSend(m.ChannelID, mutterWitze[rand.Intn(len(mutterWitze))])
		if err != nil {
			logger.ErrorLog.Println("error sending message:", err)
		}
	default:
		return
	}
}
