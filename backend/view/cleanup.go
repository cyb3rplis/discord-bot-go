package view

import (
	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/logger"
	"github.com/cyb3rplis/discord-bot-go/model"
)

func HandleCleanUp(s *discordgo.Session, m *discordgo.MessageCreate, arg, command string) error {
	meta := model.Meta
	memberRoles, err := model.GetMemberRoles(s, meta.Guild.ID, m.Author.ID)
	if err != nil {
		return err
	}

	if model.IsAdmin(memberRoles) {
		logger.InfoLog.Println("Cleanup initiated: ", m.Author)
		err := s.ChannelMessageDelete(m.ChannelID, m.ID)
		if err != nil {
			logger.ErrorLog.Println("error deleting message:", err)
		}
		messages, err := model.GetAllMessages()
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
		err = model.DeleteAllMessages()
		if err != nil {
			logger.ErrorLog.Println("error deleting all messages:", err)
		}
		return nil
	}

	err = s.ChannelMessageDelete(m.ChannelID, m.ID)
	if err != nil {
		logger.ErrorLog.Println("error deleting message:", err)
	}

	return nil
}
