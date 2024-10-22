package message

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cyb3rplis/discord-bot-go/model"

	"github.com/cyb3rplis/discord-bot-go/logger"
	"github.com/cyb3rplis/discord-bot-go/sound"
	"github.com/cyb3rplis/discord-bot-go/utils"

	"github.com/bwmarrin/discordgo"
)

// AudioMessageHandler is created on any channel that the authenticated bot has access to.
func AudioMessageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	prefix := model.Bot.Config.Prefix

	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}
	if len(m.Content) == 0 { // Ignore empty messages
		logger.InfoLog.Println("Empty content in command, ignore")
		return
	}

	args := strings.Split(m.Content, " ")
	command := args[0]
	var arg string

	switch {
	case len(args) == 2:
		// Extract the command and arguments
		arg = args[1]
	case command == fmt.Sprintf("%stts", prefix):
		break
	case len(args) > 2:
		return
	}

	switch {
	case command == fmt.Sprintf("%shelp", prefix):
		// show help text
		message := fmt.Sprintf("🧐  Help:\n> » **Sounds**\t\t\t\t%slist\n> » **Youtube Audio**\t%syoutube\n> » **Text2Speech**\t%stts\n> » **Statistics**\t\t  %sstats\n", prefix, prefix, prefix, prefix)

		utils.NewMessageRoutine(command, message, s, m, true)
		return
	case command == fmt.Sprintf("%scleanup", prefix):
		if m.Author.ID != "378670654146478081" && m.Author.ID != "481894532082958346" {
			s.ChannelMessageDelete(m.ChannelID, m.ID)
			return
		}

		logger.InfoLog.Println("Cleanup initiated: ", m.Author)

		s.ChannelMessageDelete(m.ChannelID, m.ID)
		messages, err := utils.GetAllMessages()
		if err != nil {
			logger.ErrorLog.Println("Error getting all messages:", err)
		}

		for cID, mID := range messages {
			for _, m := range mID {
				err := s.ChannelMessageDelete(cID, m)
				if err != nil {
					logger.ErrorLog.Printf("Error deleting old message - ID: %s - err: %v", m, err)
				}
			}
		}

		utils.DeleteAllMessages()
		return
	case strings.HasPrefix(command, fmt.Sprintf("%syoutube", prefix)):
		// Play the sound of a youtube video
		if len(arg) == 0 {
			message := fmt.Sprintf("🎶  Youtube: Type the URL of the video you want to play\n > » %syoutube https://...\n", prefix)
			utils.NewMessageRoutine(command, message, s, m, true)
			return
		}

		if !strings.Contains(arg, "https://") {
			message := fmt.Sprintf("🎶  Youtube: Invalid URL\n > » %syoutube https://...\n", prefix)
			utils.NewMessageRoutine(command, message, s, m, true)
			return
		}

		err := sound.LoadYouTubeAudio(arg, s, m)
		if err != nil {
			logger.ErrorLog.Println("Error loading youtube audio:", err)
			return
		}

		err = sound.PlayCustomAudio(s, m, "youtube")
		if err != nil {
			logger.ErrorLog.Println("Error playing youtube audio:", err)
		}

		utils.CleanUpSoundFile("youtube")

		return
	case strings.HasPrefix(command, fmt.Sprintf("%sstats", prefix)):
		if arg == "sounds" {
			soundStats, err := utils.GetSoundStatistics()
			if err != nil {
				logger.ErrorLog.Printf("Error getting sound statistics: %v", err)
			}
			sortedKeys := utils.SortMapKeysByValue(soundStats)

			message := "🔥  Top 10 played sounds: \n\n"
			for _, c := range sortedKeys {
				message = message + fmt.Sprintf("> %dx:\t%s\n", soundStats[c], c)
			}

			utils.NewMessageRoutine(command+arg, message, s, m, true)
			return
		} else if arg == "users" {
			userStats, err := utils.GetAllUserStatistics()
			if err != nil {
				logger.ErrorLog.Printf("Error getting all users statistics: %v", err)
			}
			sortedKeys := utils.SortMapKeysByValue(userStats)

			// send table instead of loose lines -> formatting
			message := "🤡  Top 10 Users: \n\n"
			for i, c := range sortedKeys {
				i += 1
				message = message + fmt.Sprintf("> %d.\t%s\t\tplayed: %d\n", i, c, userStats[c])
			}

			utils.NewMessageRoutine(command+arg, message, s, m, true)

			return
		} else if arg == "me" {
			userStats, err := utils.GetUserStatistics(m.Author.ID)
			if err != nil {
				logger.ErrorLog.Printf("Error getting user statistics: %v", err)
			}
			sortedKeys := utils.SortMapKeysByValue(userStats)

			message := "🔥  <@" + m.Author.ID + ">'s top 10 played sounds: \n\n"
			for _, c := range sortedKeys {
				message = message + fmt.Sprintf("> %dx:\t%s\n", userStats[c], c)
			}

			utils.NewMessageRoutine(command+m.Author.ID, message, s, m, true)

			return
		} else {
			message := fmt.Sprintf("🔥  Stats:\n> » **Global Sounds**\t\t%sstats sounds\n> » **Global Users**\t\t%sstats users\n> » **Your Sounds**\t\t\t%sstats me\n", prefix, prefix, prefix)

			utils.NewMessageRoutine(command+"help", message, s, m, false)
			return
		}

	case command == fmt.Sprintf("%stts", prefix):
		// Text2Speech
		if m.Content == fmt.Sprintf("%stts", prefix) {
			message := fmt.Sprintf("📢 TTS: Type text which will be played via Text to Speech in your Voice Channel\n > » %stts \"This is Text to Speech\"\n", prefix)

			utils.NewMessageRoutine(command+"help", message, s, m, true)
			return
		}

		ttsText := m.Content[5:len(m.Content)]
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
				err = sound.PlayCustomAudio(s, m, "tts")
				if err != nil {
					logger.ErrorLog.Println("Error playing youtube audio:", err)
				}

				utils.CleanUpSoundFile("tts")

			} else {
				logger.InfoLog.Println("TTS Text does not match regex pattern: ", ttsText)
				return
			}
			return
		}

		message := fmt.Sprintf("Text has to be in Quotes\n > » %stts \"This is Text to Speech\"\n", prefix)

		utils.NewMessageRoutine(command+"quote", message, s, m, true)
		return
	case command == fmt.Sprintf("%slist", prefix):
		// List categories
		categories, err := sound.GetCategories()
		if err != nil {
			logger.ErrorLog.Println("Error getting categories:", err)
		}
		if len(categories) == 0 {
			message := "No sound categories found."
			utils.NewMessageRoutine(command+"nocategories", message, s, m, true)

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

		message := &discordgo.MessageSend{
			Content:    "➡ Click on a category [blue button]",
			Components: content,
		}

		utils.NewComplexMessageRoutine(".listinit", m.ChannelID, m.ID, message, s, true)

		return
	case strings.Contains(strings.ToLower(m.Content), "mutter"):
		_, err := s.ChannelMessageSend(m.ChannelID, MutterWitz())
		if err != nil {
			logger.ErrorLog.Println("Error sending message:", err)
		}
	default:
		return
	}
}
