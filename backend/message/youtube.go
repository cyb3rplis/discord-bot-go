package message

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/logger"
	"github.com/cyb3rplis/discord-bot-go/model"
	"github.com/cyb3rplis/discord-bot-go/sound"
	"github.com/cyb3rplis/discord-bot-go/utils"
	"strings"
)

func HandleYoutube(s *discordgo.Session, m *discordgo.MessageCreate, arg, command string) error {
	prefix := model.Bot.Config.Prefix
	if len(arg) == 0 {
		message := fmt.Sprintf("🎶  Youtube: Type the URL of the video you want to play\n > » %syoutube https://...\n", prefix)
		utils.NewMessageRoutine(command, message, s, m)
		_ = s.MessageReactionAdd(m.ChannelID, m.ID, "✅")
		return nil
	}
	if !strings.Contains(arg, "https://") {
		message := fmt.Sprintf("🎶  Youtube: Invalid URL\n > » %syoutube https://...\n", prefix)
		utils.NewMessageRoutine(command, message, s, m)
		return fmt.Errorf("invalid youtube URL: %s", arg)
	}

	_ = s.ChannelMessageDelete(m.ChannelID, m.Message.ID)

	// Check if the user is in the Gulag
	user, err := utils.GetUserFromUsername(m.Message.Author.GlobalName)
	if err != nil {
		logger.ErrorLog.Println("error getting user from username:", err)
	} else {
		if remaining, ok := utils.IsUserInGulag(user); ok {
			user.Remaining = remaining
			message := fmt.Sprintf("<@"+user.ID+"> you are in the Gulag for another %s", user.Remaining)
			utils.NewMessageRoutine(".gulag"+user.ID, message, s, &discordgo.MessageCreate{Message: m.Message})
			return fmt.Errorf("user is in the Gulag: %s", user.ID)
		}
	}

	err = utils.VoiceChannelCheck(s, m)
	if err != nil {
		logger.ErrorLog.Println("error checking voice channel:", err)
		return err
	}
	download := Download{URL: arg, Start: "00:00:10", End: "00:00:20", Category: "youtube", SoundName: "youtube"}
	err = DownloadAndConvertAudio(download, s, m)
	if err != nil {
		logger.ErrorLog.Println("error loading youtube audio:", err)
		return err
	}

	err = sound.PlayCustomAudio(s, m, "youtube")
	if err != nil {
		logger.ErrorLog.Println("error playing youtube audio:", err)
	}

	return nil
}
