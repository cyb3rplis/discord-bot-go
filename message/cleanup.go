package message

import (
	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/logger"
	"github.com/cyb3rplis/discord-bot-go/model"
	"github.com/cyb3rplis/discord-bot-go/utils"
)

func HandleCleanUp(s *discordgo.Session, m *discordgo.MessageCreate, arg, command string) error {
	admins := model.Bot.Config.AdminUsers
	for _, a := range admins {
		if m.Author.ID == a {
			logger.InfoLog.Println("Cleanup initiated: ", m.Author)
			err := s.ChannelMessageDelete(m.ChannelID, m.ID)
			if err != nil {
				logger.ErrorLog.Println("error deleting message:", err)
			}
			messages, err := utils.GetAllMessages()
			if err != nil {
				logger.ErrorLog.Println("error getting all messages:", err)
			}
			for cID, mID := range messages {
				for _, m := range mID {
					err := s.ChannelMessageDelete(cID, m)
					if err != nil {
						logger.ErrorLog.Printf("error deleting old message - ID: %s - err: %v", m, err)
					}
				}
			}
			err = utils.DeleteAllMessages()
			if err != nil {
				logger.ErrorLog.Println("error deleting all messages:", err)
			}
			return nil
		}
	}
	err := s.ChannelMessageDelete(m.ChannelID, m.ID)
	if err != nil {
		logger.ErrorLog.Println("error deleting message:", err)
	}
	return nil
}
