package view

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/logger"
)

func (a *API) HandleYoutube(s *discordgo.Session, m *discordgo.MessageCreate, arg, command string) error {
	prefix := a.model.Config.Prefix
	if len(arg) == 0 {
		message := fmt.Sprintf("🎶  Youtube: Type the URL of the video you want to play\n > » %syoutube https://...\n", prefix)
		a.NewMessageRoutine(command, message, s, m)
		_ = s.MessageReactionAdd(m.ChannelID, m.ID, "✅")
		return nil
	}
	if !strings.Contains(arg, "https://") {
		message := fmt.Sprintf("🎶  Youtube: Invalid URL\n > » %syoutube https://...\n", prefix)
		a.NewMessageRoutine(command, message, s, m)
		return fmt.Errorf("invalid youtube URL: %s", arg)
	}

	_ = s.ChannelMessageDelete(m.ChannelID, m.Message.ID)

	// Check if the user is in the Gulag
	user, err := a.model.GetUserFromUsername(m.Message.Author.GlobalName)
	if err != nil {
		logger.ErrorLog.Println("error getting user from username:", err)
	} else {
		if remaining, ok := IsUserInGulag(user); ok {
			user.Remaining = remaining
			message := fmt.Sprintf("<@"+user.ID+"> you are in the Gulag for another %s", user.Remaining)
			a.NewMessageRoutine(".gulag"+user.ID, message, s, &discordgo.MessageCreate{Message: m.Message})
			return fmt.Errorf("user is in the Gulag: %s", user.ID)
		}
	}

	err = a.VoiceChannelCheck(s, m)
	if err != nil {
		logger.ErrorLog.Println("error checking voice channel:", err)
		return err
	}
	download := Download{URL: arg, Start: "", End: "", Category: "youtube", SoundName: "youtube"}
	err = a.DownloadAndConvertAudio(download, s, m)
	if err != nil {
		logger.ErrorLog.Println("error loading youtube audio:", err)
		return err
	}

	err = a.PlayCustomAudio(s, m, "youtube")
	if err != nil {
		logger.ErrorLog.Println("error playing youtube audio:", err)
	}

	return nil
}
