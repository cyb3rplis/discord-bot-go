package message

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/config"
	"github.com/cyb3rplis/discord-bot-go/logger"
	"github.com/cyb3rplis/discord-bot-go/model"
	"github.com/cyb3rplis/discord-bot-go/utils"
)

func HandleJail(s *discordgo.Session, m *discordgo.MessageCreate, action, user, command string) error {
	if utils.IsAdmin(m.Author.ID) {
		switch action {
		case "put":
			err := utils.JailUser(user)
			if err != nil {
				message := fmt.Sprintf("🧼  Error jailing user: %s\n", user)
				utils.NewPrivateMessageRoutine(message, s, m)
				return err
			}

			logger.InfoLog.Printf("Admin %s jailed: %s\n", m.Author.GlobalName, user)
			s.MessageReactionAdd(m.ChannelID, m.ID, "✅")

			return nil
		case "release":
			err := utils.ReleaseUser(user)
			if err != nil {
				message := fmt.Sprintf("🧼  Error releasing user: %s\n", user)
				utils.NewPrivateMessageRoutine(message, s, m)
				return err
			}

			logger.InfoLog.Printf("Admin %s released: %s\n", m.Author.GlobalName, user)
			s.MessageReactionAdd(m.ChannelID, m.ID, "✅")

			return nil
		case "list":
			users, err := utils.GetUsers()
			if err != nil {
				return err
			}

			var jailedUsers []config.User

			for _, u := range users {
				if u.Jailed.Valid {
					jailedUsers = append(jailedUsers, u)
				}
			}

			var message string
			if len(jailedUsers) == 0 {
				message = "🧼  No users are jailed\n"
			} else {
				message = "🧼  Jailed Users:\n"
				for _, u := range jailedUsers {
					message = message + fmt.Sprintf("> » ID: %s - User: %s\n", u.ID, u.Username)
				}
			}

			utils.NewPrivateMessageRoutine(message, s, m)
			return nil
		default:
			message := fmt.Sprintf("🧼  Your jail helper:\n" +
				"> » **Jail User**\t\t" + model.Bot.Config.Prefix + "jail put <user_id>\n" +
				"> » **Release User**\t\t " + model.Bot.Config.Prefix + "jail release <user_id>\n" +
				"> » **List Users**\t\t " + model.Bot.Config.Prefix + "jail list\n")
			utils.NewPrivateMessageRoutine(message, s, m)
		}
	}

	return nil
}
