package view

import (
	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/logger"
	"github.com/cyb3rplis/discord-bot-go/model"
)

func HandleList(s *discordgo.Session, m *discordgo.MessageCreate, arg, command string) error {
	categories, err := model.GetCategories()
	if err != nil {
		logger.ErrorLog.Println("error getting categories:", err)
	}
	if len(categories) == 0 {
		message := "No sound categories found."
		NewMessageRoutine(command+"nocategories", message, s, m)
		return err
	}

	content := model.BuildListButtons(categories, discordgo.PrimaryButton)
	messages := model.BuildMessages(content, nil)

	for _, message := range messages {
		NewComplexMessageRoutine(command, m.ChannelID, m.ID, message, s)
	}
	return nil
}
