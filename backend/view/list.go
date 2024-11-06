package view

import (
	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/logger"
	"github.com/cyb3rplis/discord-bot-go/model"
)

func (a *API) PromptInteractionList(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionApplicationCommand {
		switch i.ApplicationCommandData().Name {
		case "list":
			err := a.SendInteractionRespond("👉 Listing sound categories", i, s, true)
			if err != nil {
				logger.ErrorLog.Println("error executing list command:", err)
			}
			err = a.HandleList(s, &discordgo.MessageCreate{Message: &discordgo.Message{ID: i.ID, ChannelID: i.ChannelID}}, "", ".list")
			if err != nil {
				logger.ErrorLog.Println("error handling list command:", err)
			}
		}
	}
}

func (a *API) HandleList(s *discordgo.Session, m *discordgo.MessageCreate, arg, command string) error {
	categories, err := a.model.GetCategories()
	if err != nil {
		logger.ErrorLog.Println("error getting categories:", err)
	}
	if len(categories) == 0 {
		_, err = a.SendMessage("No sound categories found.", s, m, false)
		if err != nil {
			logger.ErrorLog.Println("error sending message:", err)
		}
		return err
	}
	content := model.BuildListButtons(categories, discordgo.PrimaryButton)
	messages := model.BuildMessages(content, nil)
	for _, message := range messages {
		_, err = a.SendMessageComplex(message, s, m, false)
		if err != nil {
			logger.ErrorLog.Println("error sending message:", err)
		}
	}
	return nil
}
