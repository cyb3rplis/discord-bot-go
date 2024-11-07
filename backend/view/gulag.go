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
				var gulaggedUsers []config.ExtendedUser
				for _, u := range users {
					if u, ok := SetUserGulagRemaining(u); ok {
						gulaggedUsers = append(gulaggedUsers, u)
					}
				}
				message := "🧱  Gulagged Users:\n"
				for _, u := range gulaggedUsers {
					message = message + fmt.Sprintf("> » User: %s - Remaining: %s\n", u.User.GlobalName, u.Remaining)
				}
				if len(gulaggedUsers) == 0 {
					message = "🧱  No users are gulagged\n"
				} else {

				}
				err = a.SendInteractionRespond(message, s, i)
				if err != nil {
					return
				}
				return
			case "add":
				if len(option.Options) > 0 {
					userGlobalName := option.Options[0].StringValue()
					user := config.ExtendedUser{
						User: &discordgo.User{
							GlobalName: userGlobalName,
						},
					}
					timeout := option.Options[1].StringValue()
					minutes, err := strconv.Atoi(timeout)
					if err != nil {
						_ = a.SendInteractionRespond(fmt.Sprintf("🧱  Invalid time out value: %s\n", timeout), s, i)
						dlog.ErrorLog.Println("error converting timeout to int:", err)
						return
					}
					err = a.model.GulagUser(user, minutes)
					if err != nil {
						_ = a.UpdateInteractionResponse(fmt.Sprintf("🧱  error putting user into gulag: %s --> %v \n", user.User.GlobalName, err), s, i)
						dlog.ErrorLog.Println("error putting user into gulag:", err)
						return
					}
					dlog.InfoLog.Printf("Admin %s put user: %s in the gulag for %d Minutes\n", m.Author.GlobalName, user.User.GlobalName, minutes)
					err = a.SendInteractionRespond(fmt.Sprintf("🧱  User %s has been put into the gulag for %d minutes\n", user.User.GlobalName, minutes), s, i)
					if err != nil {
						dlog.ErrorLog.Println("error sending hidden message:", err)
						return
					}
					return
				}
			case "remove":
				if len(option.Options) > 0 {
					user := config.ExtendedUser{
						User: &discordgo.User{
							GlobalName: option.Options[0].StringValue(),
						},
					}
					err := a.model.ReleaseUser(user)
					if err != nil {
						dlog.ErrorLog.Println("error releasing user from gulag:", err)
						_ = a.SendInteractionRespond(fmt.Sprintf("🧱  error releasing user %s from gulag --> %v \n", user.User.GlobalName, err), s, i)
						return
					}
					dlog.InfoLog.Printf("Admin %s released: %s from gulag\n", m.Author.GlobalName, user.User.GlobalName)
					err = a.SendInteractionRespond(fmt.Sprintf("🧱  User %s has been released from the gulag\n", user.User.GlobalName), s, i)
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

func SetUserGulagRemaining(user config.ExtendedUser) (config.ExtendedUser, bool) {
	var remaining time.Duration
	now := time.Now()

	if user.Gulagged.Valid {
		rem := user.Gulagged.Time.Sub(now)
		if rem > 0 {
			remaining = rem.Truncate(time.Second)
			user.Remaining = remaining
			return user, true
		}
	}
	return user, false
}
