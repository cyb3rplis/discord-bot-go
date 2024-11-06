package view

import (
	"fmt"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/config"
	"github.com/cyb3rplis/discord-bot-go/dlog"
)

func (a *API) PromptInteractionGulag(s *discordgo.Session, i *discordgo.InteractionCreate) {
	//get userID
	if i.Member == nil {
		dlog.ErrorLog.Println("error getting member from interaction")
		return
	}
	m := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			ID:        i.ID,
			ChannelID: i.ChannelID,
			Author:    &discordgo.User{ID: i.Member.User.ID},
		},
	}
	if i.Type == discordgo.InteractionApplicationCommand {
		switch i.ApplicationCommandData().Name {
		case "gulag":
			option := i.ApplicationCommandData().Options[0]
			switch option.Name {
			case "list":
				users, err := a.model.GetUsers()
				if err != nil {
					return
				}
				var gulaggedUsers []config.User
				for _, u := range users {
					if remaining, ok := IsUserInGulag(u); ok {
						u.Remaining = remaining
						gulaggedUsers = append(gulaggedUsers, u)
					}
				}
				message := "🧱  Gulagged Users:\n"
				for _, u := range gulaggedUsers {
					message = message + fmt.Sprintf("> » User: %s - Remaining: %s\n", u.Username, u.Remaining)
				}
				if len(gulaggedUsers) == 0 {
					message = "🧱  No users are gulagged\n"
				} else {

				}
				err = a.SendInteractionRespond(message, s, i, true)
				if err != nil {
					return
				}
				return
			case "add":
				if len(option.Options) > 0 {
					user := option.Options[0].StringValue()
					timeout := option.Options[1].StringValue()
					minutes, err := strconv.Atoi(timeout)
					if err != nil {
						_ = a.SendInteractionRespond(fmt.Sprintf("🧱  Invalid time out value: %s\n", timeout), s, i, true)
						return
					}
					err = a.model.GulagUser(user, minutes)
					if err != nil {
						_ = a.SendInteractionRespond(fmt.Sprintf("🧱  error putting user into gulag: %s --> %v \n", user, err), s, i, true)
						return
					}
					dlog.InfoLog.Printf("Admin %s put user: %s in the gulag for %d Minutes\n", m.Author.GlobalName, user, minutes)
					err = a.SendInteractionRespond(fmt.Sprintf("🧱  User %s has been put into the gulag for %d minutes\n", user, minutes), s, i, true)
					if err != nil {
						dlog.ErrorLog.Println("error sending hidden message:", err)
						return
					}
					return
				}
			case "remove":
				if len(option.Options) > 0 {
					user := option.Options[0].StringValue()
					err := a.model.ReleaseUser(user)
					if err != nil {
						dlog.ErrorLog.Println("error releasing user from gulag:", err)
						_ = a.SendInteractionRespond(fmt.Sprintf("🧱  error releasing user %s from gulag --> %v \n", user, err), s, i, true)
						return
					}
					dlog.InfoLog.Printf("Admin %s released: %s from gulag\n", m.Author.GlobalName, user)
					err = a.SendInteractionRespond(fmt.Sprintf("🧱  User %s has been released from the gulag\n", user), s, i, true)
					if err != nil {
						dlog.ErrorLog.Println("error sending hidden message:", err)
						return
					}

					return
				}
			}
		}
	}
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
