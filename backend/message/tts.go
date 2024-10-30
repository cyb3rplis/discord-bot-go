package message

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/logger"
	"github.com/cyb3rplis/discord-bot-go/model"
	"github.com/cyb3rplis/discord-bot-go/sound"
	"github.com/cyb3rplis/discord-bot-go/utils"
)

func HandleTTS(s *discordgo.Session, m *discordgo.MessageCreate, command string) error {
	prefix := model.Bot.Config.Prefix
	if m.Content == fmt.Sprintf("%stts", prefix) {
		message := fmt.Sprintf("📢  TTS: Type text which will be played via Text to Speech in your Voice Channel\n > » %stts \"This is Text to Speech\"\n", prefix)

		utils.NewMessageRoutine(command+"help", message, s, m)
		_ = s.MessageReactionAdd(m.ChannelID, m.ID, "✅")
		return nil
	}

	_ = s.ChannelMessageDelete(m.ChannelID, m.ID)

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

	ttsText := m.Content[5:len(m.Content)]
	if strings.HasPrefix(ttsText, "\"") && strings.HasSuffix(ttsText, "\"") {
		pattern := `^\"[öäüÖÄÜa-zA-Z0-9\.!:,? ]+\"$`
		re, err := regexp.Compile(pattern)
		if err != nil {
			logger.ErrorLog.Println("error compiling regex:", err)
			return err
		}
		if re.MatchString(ttsText) {
			err := utils.VoiceChannelCheck(s, m)
			if err != nil {
				logger.ErrorLog.Println("error checking voice channel:", err)
				return err
			}

			utils.CleanUpSoundFile("tts")

			err = utils.TextToSpeech(ttsText)
			if err != nil {
				logger.ErrorLog.Println("error converting text to speech:", err)
				return err
			}

			err = utils.WAVtoDCA()
			if err != nil {
				logger.ErrorLog.Println("error converting wav to dca:", err)
				return err
			}

			// play sound and clean up files
			err = sound.PlayCustomAudio(s, m, "tts")
			if err != nil {
				logger.ErrorLog.Println("error playing youtube audio:", err)
			}
		} else {
			logger.InfoLog.Println("TTS Text does not match regex pattern: ", ttsText)
			return err
		}
		return nil
	}

	message := fmt.Sprintf("Text has to be in Quotes\n > » %stts \"This is Text to Speech\"\n", prefix)
	utils.NewMessageRoutine(command+"quote", message, s, m)
	return nil
}
