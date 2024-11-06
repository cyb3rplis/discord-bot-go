package view

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/logger"
	"github.com/cyb3rplis/discord-bot-go/model"
)

func (a *API) PromptInteractionStats(s *discordgo.Session, i *discordgo.InteractionCreate) {
	//get userID
	if i.Member == nil {
		logger.ErrorLog.Println("error getting member from interaction")
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
		case "stats":
			option := i.ApplicationCommandData().Options[0]
			switch option.Name {
			case "sounds":
				err := a.SendInteractionRespond("👉 Getting sound statistics", i, s, true)
				if err != nil {
					logger.ErrorLog.Println("error executing stats command:", err)
				}
				err = a.soundStats(s, m)
				if err != nil {
					logger.ErrorLog.Println("error executing stats command:", err)
				}
			case "users":
				err := a.SendInteractionRespond("👉 Getting user statistics", i, s, true)
				if err != nil {
					logger.ErrorLog.Println("error executing stats command:", err)
				}
				err = a.userStats(s, m)
				if err != nil {
					logger.ErrorLog.Println("error executing stats command:", err)
				}
			case "me":
				err := a.SendInteractionRespond("👉 Getting your statistics", i, s, true)
				if err != nil {
					logger.ErrorLog.Println("error executing stats command:", err)
				}
				err = a.meStats(s, m)
				if err != nil {
					logger.ErrorLog.Println("error executing stats command:", err)
				}
			}
		}
	}
}

func (a *API) soundStats(s *discordgo.Session, m *discordgo.MessageCreate) error {
	soundStats, err := a.model.GetSoundStatistics()
	if err != nil {
		logger.ErrorLog.Printf("error getting sound statistics: %v", err)
	}
	sortedKeys := model.SortMapKeysByValue(soundStats)
	message := "🔥  Top 10 played sounds: \n\n"
	for _, c := range sortedKeys {
		message = message + fmt.Sprintf("> %dx:\t%s\n", soundStats[c], c)
	}
	_, err = a.SendMessage(message, s, m, false)
	if err != nil {
		logger.ErrorLog.Println("error sending message:", err)
	}
	return nil
}

func (a *API) userStats(s *discordgo.Session, m *discordgo.MessageCreate) error {
	userStats, err := a.model.GetAllUserStatistics()
	if err != nil {
		logger.ErrorLog.Printf("error getting all users statistics: %v", err)
	}
	sortedKeys := model.SortMapKeysByValue(userStats)

	// send table instead of loose lines -> formatting
	message := "🤡  Top 10 Users: \n\n"
	for i, c := range sortedKeys {
		i += 1
		message = message + fmt.Sprintf("> %d.\t%s\t\tplayed: %d\n", i, c, userStats[c])
	}

	_, err = a.SendMessage(message, s, m, false)
	if err != nil {
		logger.ErrorLog.Println("error sending message:", err)
	}
	return nil
}

func (a *API) meStats(s *discordgo.Session, m *discordgo.MessageCreate) error {
	userStats, err := a.model.GetUserStatistics(m.Author.ID, 10)
	if err != nil {
		logger.ErrorLog.Printf("error getting user statistics: %v", err)
	}
	content := []discordgo.MessageComponent{}
	row := discordgo.ActionsRow{}
	for i, s := range userStats {
		// only 5 buttons per row - discord does not allow more
		if i > 0 && i%5 == 0 {
			content = append(content, row)
			row = discordgo.ActionsRow{}
		}
		row.Components = append(row.Components, discordgo.Button{
			Label:    fmt.Sprintf("%dx: %s", s.Count, s.Name),
			Style:    discordgo.SuccessButton,
			CustomID: fmt.Sprintf("play_sound_%s_%s", s.Category, s.Name),
		})
	}
	// Append the last row if it has any components
	if len(row.Components) > 0 {
		content = append(content, row)
	}
	message2 := &discordgo.MessageSend{
		Content:    "🔥  <@" + m.Author.ID + ">'s top 10 played sounds: \n\n",
		Components: content,
	}

	_, err = a.SendMessageComplex(message2, s, m, false)
	if err != nil {
		logger.ErrorLog.Println("error sending message:", err)
	}
	return nil
}
