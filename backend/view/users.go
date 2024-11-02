package view

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/model"
)

func HandleUsers(s *discordgo.Session, m *discordgo.MessageCreate, command string) error {
	meta := model.Meta
	memberRoles, err := model.GetMemberRoles(s, meta.Guild.ID, m.Author.ID)
	if err != nil {
		return err
	}

	if model.IsAdmin(memberRoles) {
		users, err := model.GetUsers()
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

		err = NewPrivateMessageRoutine(message, s, m)
		if err != nil {
			return err
		}
		return nil
	}

	return nil
}
