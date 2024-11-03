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

func (a *API) HandleGulag(s *discordgo.Session, mc *discordgo.MessageCreate, action, user, timeOut, command string) error {
	meta := model.Meta
	memberRoles, err := model.GetMemberRoles(s, meta.Guild.ID, mc.Author.ID)
	if err != nil {
		return err
	}

	if a.model.IsAdmin(memberRoles) {
		switch action {
		case "add":
			if timeOut == "" {
				err := a.model.GulagUser(user, 3)
				if err != nil {
					message := fmt.Sprintf("🧱  Error gulagging user: %s\n", user)
					err = NewPrivateMessageRoutine(message, s, mc)

					return err
				}

				logger.InfoLog.Printf("Admin %s gulagged user: %s for 3 Minutes\n", mc.Author.GlobalName, user)
				return nil
			} else {
				minutes, err := strconv.Atoi(timeOut)
				if err != nil {
					message := fmt.Sprintf("🧱  Invalid time out value: %s\n", timeOut)
					err = NewPrivateMessageRoutine(message, s, mc)
					return err
				}

				err = a.model.GulagUser(user, minutes)
				if err != nil {
					message := fmt.Sprintf("🧱  Error gulagging user: %s\n", user)
					err = NewPrivateMessageRoutine(message, s, mc)
					return err
				}

				logger.InfoLog.Printf("Admin %s gulagged user: %s for %d Minutes\n", mc.Author.GlobalName, user, minutes)
				return nil
			}
		case "rm":
			err := a.model.ReleaseUser(user)
			if err != nil {
				message := fmt.Sprintf("🧱  Error releasing user %s from gulag\n", user)
				err = NewPrivateMessageRoutine(message, s, mc)
				return err
			}

			logger.InfoLog.Printf("Admin %s released: %s from gulag\n", mc.Author.GlobalName, user)
			return nil
		case "list":
			users, err := a.model.GetUsers()
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
			err = NewPrivateMessageRoutine(message, s, mc)
			if err != nil {
				return err
			}
			return nil
		default:
			message := fmt.Sprintf("🧱  Your gulag helper:\n" +
				"> » **Gulag User**\t\t" + a.model.Config.Prefix + "gulag add <username> <optional: minutes, default 3>\n" +
				"> » **Release User**\t\t " + a.model.Config.Prefix + "gulag rm <username>\n" +
				"> » **List Users**\t\t " + a.model.Config.Prefix + "gulag list\n")
			err = NewPrivateMessageRoutine(message, s, mc)
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
