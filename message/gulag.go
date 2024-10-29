package message

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/config"
	"github.com/cyb3rplis/discord-bot-go/logger"
	"github.com/cyb3rplis/discord-bot-go/model"
	"github.com/cyb3rplis/discord-bot-go/utils"
)

func HandleGulag(s *discordgo.Session, m *discordgo.MessageCreate, action, user, command string) error {
	if utils.IsAdmin(m.Author.ID) {
		switch action {
		case "put":
			err := utils.GulagUser(user)
			if err != nil {
				message := fmt.Sprintf("🧼  Error gulagging user: %s\n", user)
				utils.NewPrivateMessageRoutine(message, s, m)
				return err
			}

			logger.InfoLog.Printf("Admin %s gulagged: %s\n", m.Author.GlobalName, user)
			_ = s.MessageReactionAdd(m.ChannelID, m.ID, "✅")

			return nil
		case "release":
			err := utils.ReleaseUser(user)
			if err != nil {
				message := fmt.Sprintf("🧼  Error releasing user: %s\n", user)
				utils.NewPrivateMessageRoutine(message, s, m)
				return err
			}

			logger.InfoLog.Printf("Admin %s released: %s\n", m.Author.GlobalName, user)
			_ = s.MessageReactionAdd(m.ChannelID, m.ID, "✅")

			return nil
		case "list":
			users, err := utils.GetUsers()
			if err != nil {
				return err
			}

			var gulaggedUsers []config.User

			for _, u := range users {
				if u.Gulagged.Valid {
					gulaggedUsers = append(gulaggedUsers, u)
				}
			}

			var message string
			if len(gulaggedUsers) == 0 {
				message = "🧼  No users are gulagged\n"
			} else {
				message = "🧼  Gulagged Users:\n"
				for _, u := range gulaggedUsers {
					message = message + fmt.Sprintf("> » ID: %s - User: %s\n", u.ID, u.Username)
				}
			}

			utils.NewPrivateMessageRoutine(message, s, m)
			return nil
		default:
			message := fmt.Sprintf("🧼  Your gulag helper:\n" +
				"> » **Gulag User**\t\t" + model.Bot.Config.Prefix + "gulag put <user_id>\n" +
				"> » **Release User**\t\t " + model.Bot.Config.Prefix + "gulag release <user_id>\n" +
				"> » **List Users**\t\t " + model.Bot.Config.Prefix + "gulag list\n")
			utils.NewPrivateMessageRoutine(message, s, m)
			_ = s.MessageReactionAdd(m.ChannelID, m.ID, "✅")
		}
	}

	return nil
}
