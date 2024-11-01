package message

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/utils"
)

func HandleUsers(s *discordgo.Session, m *discordgo.MessageCreate, command string) error {
	memberRoles, err := utils.GetMemberRoles(s, m.GuildID, m.Author.ID)
	if err != nil {
		return err
	}

	if utils.IsAdmin(memberRoles) {
		users, err := utils.GetUsers()
		if err != nil {
			return err
		}

		var message string
		if len(users) == 0 {
			message = "🍆  No users\n"
		} else {
			message = "🍆  Users:\n"
			for _, u := range users {
				message = message + fmt.Sprintf("> » ID: %s\tUser: %s\n", u.ID, u.Username)
			}
		}

		utils.NewPrivateMessageRoutine(message, s, m)
		return nil
	}

	return nil
}
