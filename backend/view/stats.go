package view

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/dlog"
	"github.com/cyb3rplis/discord-bot-go/model"
)

func (a *API) PromptInteractionStats(s *discordgo.Session, i *discordgo.InteractionCreate) {
	//get userID
	interactionUser := i.Member.User

	if i.Type == discordgo.InteractionApplicationCommand {
		switch i.ApplicationCommandData().Name {
		case "stats":
			option := i.ApplicationCommandData().Options[0]
			switch option.Name {
			case "sounds":
				err := a.SendInteractionRespond("👉 Getting sound statistics", s, i)
				if err != nil {
					dlog.ErrorLog.Println("error executing stats command:", err)
				}
				stats, err := a.getStatsSounds()
				if err != nil {
					dlog.ErrorLog.Println("error executing stats command:", err)
				}
				err = a.UpdateInteractionResponse(stats, s, i)
				if err != nil {
					dlog.ErrorLog.Println("error executing stats command:", err)
				}
			case "users":
				err := a.SendInteractionRespond("👉 Getting user statistics", s, i)
				if err != nil {
					dlog.ErrorLog.Println("error executing stats command:", err)
				}
				stats, err := a.getStatsUsers()
				if err != nil {
					dlog.ErrorLog.Println("error executing stats command:", err)
				}
				err = a.UpdateInteractionResponse(stats, s, i)
				if err != nil {
					dlog.ErrorLog.Println("error executing stats command:", err)
				}
			case "me":
				err := a.SendInteractionRespond("👉 Getting your statistics", s, i)
				if err != nil {
					dlog.ErrorLog.Println("error executing stats command:", err)
				}
				stats, err := a.getStatsMe(interactionUser)
				if err != nil {
					dlog.ErrorLog.Println("error executing stats command:", err)
				}
				err = a.UpdateInteractionResponse(stats, s, i)
				if err != nil {
					dlog.ErrorLog.Println("error executing stats command:", err)
				}
			}
		default:
			err := a.SendInteractionRespond("👉  Something went wrong...", s, i)
			if err != nil {
				dlog.ErrorLog.Println("fallback to default stats handler", err)
			}
		}
	}
}

func (a *API) getStatsSounds() (string, error) {
	soundStats, err := a.model.GetSoundStatistics()
	if err != nil {
		dlog.ErrorLog.Printf("error getting sound statistics: %v", err)
	}
	sortedKeys := model.SortMapKeysByValue(soundStats)
	content := strings.Builder{}
	content.Write([]byte("🔥  Top 10 sounds: \n\n"))
	for _, c := range sortedKeys {
		content.Write([]byte(fmt.Sprintf("> %dx:\t%s\n", soundStats[c], c)))
	}
	return content.String(), nil
}

func (a *API) getStatsUsers() (string, error) {
	userStats, err := a.model.GetAllUserStatistics()
	if err != nil {
		dlog.ErrorLog.Printf("error getting all users statistics: %v", err)
	}
	sortedKeys := model.SortMapKeysByValue(userStats)

	content := strings.Builder{}
	content.Write([]byte("🔥  Top 10 users: \n\n"))
	for _, c := range sortedKeys {
		content.Write([]byte(fmt.Sprintf("> %dx:\t%s\n", userStats[c], c)))
	}
	return content.String(), nil
}

func (a *API) getStatsMe(user *discordgo.User) (string, error) {
	userStats, err := a.model.GetUserStatistics(user, 10)
	if err != nil {
		dlog.ErrorLog.Printf("error getting user statistics: %v", err)
	}
	content := strings.Builder{}
	content.Write([]byte("🔥  " + user.Mention() + "'s top 10 played sounds: \n\n"))
	for _, s := range userStats {
		content.Write([]byte(fmt.Sprintf("> %dx:\t%s\n", s.Count, s.Name)))
	}
	return content.String(), nil
}
