package message

import (
	"fmt"
	"math/rand"
	"regexp"
	"strings"
	"time"

	"github.com/cyb3rplis/discord-bot-go/logger"
	"github.com/cyb3rplis/discord-bot-go/sound"
	"github.com/cyb3rplis/discord-bot-go/utils"

	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/config"
)

// AudioMessageHandler is created on any channel that the authenticated bot has access to.
func AudioMessageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	cfg := config.GetConfig()
	prefix := cfg.Prefix

	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}
	if len(m.Content) == 0 { // Ignore empty messages
		logger.InfoLog.Println("Empty content in command, ignore")
		return
	}
	// Extract the command and arguments
	args := strings.Split(m.Content, " ")
	command := args[0]

	switch {
	case command == fmt.Sprintf("%shelp", prefix):
		// show help text
		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("🧐  Usage:\n> » **Sounds**\t\t\t\t%slist\n> » **Text2Speech**\t%stts\n> » **Statistics**\t\t\t%sstats\n", prefix, prefix, prefix))
		if err != nil {
			logger.ErrorLog.Println("error sending message:", err)
		}
	case command == fmt.Sprintf("%sstats", prefix):
		soundStats, err := sound.GetSoundStatistics()
		if err != nil {
			logger.ErrorLog.Printf("Error getting sound statistics: %v", err)
		}

		message := "🔥  Top 10 played sounds: \n\n"
		for s, c := range soundStats {
			message = message + fmt.Sprintf("> » %s - %d\n", s, c)
		}

		_, err = s.ChannelMessageSend(m.ChannelID, message)
		if err != nil {
			logger.ErrorLog.Println("error sending message:", err)
		}

		return
	case command == fmt.Sprintf("%stts", prefix):
		// Text2Speech
		ttsText := m.Content[5:len(m.Content)]

		if m.Content == fmt.Sprintf("%stts", prefix) {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("📢 Type text which will be played via Text to Speech in your Voice Channel\n > » %stts \"This is Text to Speech\"\n", prefix))
			if err != nil {
				logger.ErrorLog.Println("error sending message:", err)
				return
			}

			return
		}

		if strings.HasPrefix(ttsText, "\"") && strings.HasSuffix(ttsText, "\"") {
			pattern := `^\"[öäüÖÄÜa-zA-Z0-9\.!:,? ]+\"$`

			re, err := regexp.Compile(pattern)
			if err != nil {
				logger.ErrorLog.Println("Error compiling regex:", err)
				return
			}

			if re.MatchString(ttsText) {
				err := utils.TextToSpeech(ttsText)
				if err != nil {
					logger.ErrorLog.Println("Error converting text to speech:", err)
					return
				}

				err = utils.WAVtoDCA()
				if err != nil {
					logger.ErrorLog.Println("Error converting wav to dca:", err)
					return
				}

				// play sound and clean up files
				//sound.HandlePlaySoundInteraction(s, i, "play_sound_temp_tts")
				time.Sleep(150 * time.Millisecond)
				//utils.CleanUpSoundFile()

			} else {
				logger.InfoLog.Println("TTS Text does not match regex pattern: ", ttsText)
				return
			}
			return
		}

		_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Text has to be in Quotes\n > » %stts \"This is Text to Speech\"\n", prefix))
		if err != nil {
			logger.ErrorLog.Println("error sending message:", err)
			return
		}
	case command == fmt.Sprintf("%slist", prefix):
		// List categories
		// Get all sound folders to use for later
		//get categories from database
		categories, err := sound.GetCategories()
		if err != nil {
			logger.ErrorLog.Println("error getting categories:", err)
		}
		if len(categories) == 0 {
			_, err := s.ChannelMessageSend(m.ChannelID, "No sound categories found.")
			if err != nil {
				logger.ErrorLog.Println("error sending message:", err)
			}
			return
		}
		content := []discordgo.MessageComponent{}
		row := discordgo.ActionsRow{}
		for i, category := range categories {
			// only 5 buttons per row - discord does not allow more
			if i > 0 && i%5 == 0 {
				content = append(content, row)
				row = discordgo.ActionsRow{}
			}
			row.Components = append(row.Components, discordgo.Button{
				Label:    category,
				Style:    discordgo.PrimaryButton,
				CustomID: fmt.Sprintf("list_sounds_%s", category),
			})
		}
		// Append the last row if it has any components
		if len(row.Components) > 0 {
			content = append(content, row)
		}
		_, err = s.ChannelMessageSendComplex(m.ChannelID, &discordgo.MessageSend{
			Content:    "➡ Click on a category [blue button]",
			Components: content,
		})
		if err != nil {
			logger.ErrorLog.Println("error sending message:", err)
		}
	case strings.Contains(strings.ToLower(m.Content), "mutter"):
		_, err := s.ChannelMessageSend(m.ChannelID, mutterWitze[rand.Intn(len(mutterWitze))])
		if err != nil {
			logger.ErrorLog.Println("error sending message:", err)
		}
	default:
		return
	}
}
