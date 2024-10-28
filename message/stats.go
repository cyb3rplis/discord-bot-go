package message

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/logger"
	"github.com/cyb3rplis/discord-bot-go/model"
	"github.com/cyb3rplis/discord-bot-go/utils"
)

func HandleStatistics(s *discordgo.Session, m *discordgo.MessageCreate, arg, command string) error {
	prefix := model.Bot.Config.Prefix
	switch arg {
	case "sounds":
		err := soundStats(s, m, arg, command)
		if err != nil {
			logger.ErrorLog.Println("error listing sounds:", err)
		}
	case "users":
		err := userStats(s, m, arg, command)
		if err != nil {
			logger.ErrorLog.Println("error listing users:", err)
		}
	case "me":
		err := meStats(s, m, arg, command)
		if err != nil {
			logger.ErrorLog.Println("error listing user sounds:", err)
		}
	default:
		message := fmt.Sprintf("🔥  Stats:\n> » **Global Sounds**\t\t%sstats sounds\n> » **Global Users**\t\t%sstats users\n> » **Your Sounds**\t\t\t%sstats me\n", prefix, prefix, prefix)
		utils.NewMessageRoutine(command+"help", message, s, m, false)
		return nil
	}
	return nil
}

func soundStats(s *discordgo.Session, m *discordgo.MessageCreate, arg, command string) error {
	soundStats, err := utils.GetSoundStatistics()
	if err != nil {
		logger.ErrorLog.Printf("error getting sound statistics: %v", err)
	}
	sortedKeys := utils.SortMapKeysByValue(soundStats)

	message := "🔥  Top 10 played sounds: \n\n"
	for _, c := range sortedKeys {
		message = message + fmt.Sprintf("> %dx:\t%s\n", soundStats[c], c)
	}
	utils.NewMessageRoutine(command+arg, message, s, m, true)
	return nil
}

func userStats(s *discordgo.Session, m *discordgo.MessageCreate, arg, command string) error {
	userStats, err := utils.GetAllUserStatistics()
	if err != nil {
		logger.ErrorLog.Printf("error getting all users statistics: %v", err)
	}
	sortedKeys := utils.SortMapKeysByValue(userStats)

	// send table instead of loose lines -> formatting
	message := "🤡  Top 10 Users: \n\n"
	for i, c := range sortedKeys {
		i += 1
		message = message + fmt.Sprintf("> %d.\t%s\t\tplayed: %d\n", i, c, userStats[c])
	}
	utils.NewMessageRoutine(command+arg, message, s, m, true)
	return nil
}

func meStats(s *discordgo.Session, m *discordgo.MessageCreate, arg, command string) error {
	userStats, err := utils.GetUserStatistics(m.Author.ID, 10)
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
	utils.NewComplexMessageRoutine(command+arg+m.Author.ID, m.ChannelID, m.ID, message2, s, true)
	return nil
}
