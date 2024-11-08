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
	interactionUser := i.Member.User
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
					gulagUser := config.ExtendedUser{
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
					err = a.model.GulagUser(gulagUser, minutes)
					if err != nil {
						_ = a.UpdateInteractionResponse(fmt.Sprintf("🧱  error putting user into gulag: %s --> %v \n", gulagUser.User.GlobalName, err), s, i)
						dlog.ErrorLog.Println("error putting user into gulag:", err)
						return
					}
					dlog.InfoLog.Printf("Admin %s put user: %s in the gulag for %d Minutes\n", interactionUser.GlobalName, gulagUser.User.GlobalName, minutes)
					err = a.SendInteractionRespond(fmt.Sprintf("🧱  User %s has been put into the gulag for %d minutes\n", gulagUser.User.GlobalName, minutes), s, i)
					if err != nil {
						dlog.ErrorLog.Println("error sending hidden message:", err)
						return
					}
					return
				}
			case "remove":
				if len(option.Options) > 0 {
					gulagUser := config.ExtendedUser{
						User: &discordgo.User{
							GlobalName: option.Options[0].StringValue(),
						},
					}
					err := a.model.ReleaseUser(gulagUser)
					if err != nil {
						dlog.ErrorLog.Println("error releasing user from gulag:", err)
						_ = a.SendInteractionRespond(fmt.Sprintf("🧱  error releasing user %s from gulag --> %v \n", gulagUser.User.GlobalName, err), s, i)
						return
					}
					dlog.InfoLog.Printf("Admin %s released: %s from gulag\n", interactionUser.GlobalName, gulagUser.User.GlobalName)
					err = a.SendInteractionRespond(fmt.Sprintf("🧱  User %s has been released from the gulag\n", interactionUser.GlobalName), s, i)
					if err != nil {
						dlog.ErrorLog.Println("error sending hidden message:", err)
						return
					}

					return
				}
			default:
				err := a.SendInteractionRespond("🧱  Something went wrong...", s, i)
				if err != nil {
					dlog.ErrorLog.Println("fallback to default gulag handler", err)
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
