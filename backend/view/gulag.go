package view

import (
	"fmt"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/config"
	"github.com/cyb3rplis/discord-bot-go/logger"
	"github.com/cyb3rplis/discord-bot-go/model"
)

func HandleGulag(s *discordgo.Session, m *discordgo.MessageCreate, action, user, timeOut, command string) error {
	meta := model.Meta
	memberRoles, err := model.GetMemberRoles(s, meta.Guild.ID, m.Author.ID)
	if err != nil {
		return err
	}

	if model.IsAdmin(memberRoles) {
		switch action {
		case "add":
			if timeOut == "" {
				err := model.GulagUser(user, 3)
				if err != nil {
					message := fmt.Sprintf("🧱  Error gulagging user: %s\n", user)
					err = NewPrivateMessageRoutine(message, s, m)

					return err
				}

				logger.InfoLog.Printf("Admin %s gulagged user: %s for 3 Minutes\n", m.Author.GlobalName, user)
				return nil
			} else {
				minutes, err := strconv.Atoi(timeOut)
				if err != nil {
					message := fmt.Sprintf("🧱  Invalid time out value: %s\n", timeOut)
					err = NewPrivateMessageRoutine(message, s, m)
					return err
				}

				err = model.GulagUser(user, minutes)
				if err != nil {
					message := fmt.Sprintf("🧱  Error gulagging user: %s\n", user)
					err = NewPrivateMessageRoutine(message, s, m)
					return err
				}

				logger.InfoLog.Printf("Admin %s gulagged user: %s for %d Minutes\n", m.Author.GlobalName, user, minutes)
				return nil
			}
		case "rm":
			err := model.ReleaseUser(user)
			if err != nil {
				message := fmt.Sprintf("🧱  Error releasing user %s from gulag\n", user)
				err = NewPrivateMessageRoutine(message, s, m)
				return err
			}

			logger.InfoLog.Printf("Admin %s released: %s from gulag\n", m.Author.GlobalName, user)
			return nil
		case "list":
			users, err := model.GetUsers()
			if err != nil {
				return err
			}

			var gulaggedUsers []config.User
			for _, u := range users {
				if remaining, ok := IsUserInGulag(u); ok {
					u.Remaining = remaining
					gulaggedUsers = append(gulaggedUsers, u)
				}
			}

			var message string
			if len(gulaggedUsers) == 0 {
				message = "🧱  No users are gulagged\n"
			} else {
				message = "🧱  Gulagged Users:\n"
				for _, u := range gulaggedUsers {
					message = message + fmt.Sprintf("> » User: %s - Remaining: %s\n", u.Username, u.Remaining)
				}
			}
			err = NewPrivateMessageRoutine(message, s, m)
			if err != nil {
				return err
			}
			return nil
		default:
			message := fmt.Sprintf("🧱  Your gulag helper:\n" +
				"> » **Gulag User**\t\t" + model.Bot.Config.Prefix + "gulag add <username> <optional: minutes, default 3>\n" +
				"> » **Release User**\t\t " + model.Bot.Config.Prefix + "gulag rm <username>\n" +
				"> » **List Users**\t\t " + model.Bot.Config.Prefix + "gulag list\n")
			err = NewPrivateMessageRoutine(message, s, m)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func IsUserInGulag(user config.User) (time.Duration, bool) {
	var remaining time.Duration
	now := time.Now()

	if user.Gulagged.Valid {
		rem := user.Gulagged.Time.Sub(now)
		if rem > 0 {
			remaining = rem.Truncate(time.Second)
			return remaining, true
		}
	}
	return remaining, false
}
