package message

import (
	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/logger"
	"github.com/cyb3rplis/discord-bot-go/sound"
	"github.com/cyb3rplis/discord-bot-go/utils"
)

func HandleList(s *discordgo.Session, m *discordgo.MessageCreate, arg, command string) error {
	categories, err := sound.GetCategories()
	if err != nil {
		logger.ErrorLog.Println("error getting categories:", err)
	}
	if len(categories) == 0 {
		message := "No sound categories found."
		utils.NewMessageRoutine(command+"nocategories", message, s, m, true)
		return err
	}

	content := utils.BuildListButtons(categories, discordgo.PrimaryButton)
	messages := utils.BuildMessages(content)

	for _, message := range messages {
		utils.NewComplexMessageRoutine(command, m.ChannelID, m.ID, message, s, true)
	}
	return nil
}
