package message

import (
	"fmt"
	"strings"

	"github.com/cyb3rplis/discord-bot-go/model"

	"github.com/cyb3rplis/discord-bot-go/logger"
	"github.com/cyb3rplis/discord-bot-go/utils"

	"github.com/bwmarrin/discordgo"
)

// AudioMessageHandler is created on any channel that the authenticated bot has access to.
func AudioMessageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	isDM := m.GuildID == ""
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
	var arg2 string
	switch {
	case len(args) == 2:
		// Extract the command and arguments
		arg = args[1]
	case len(args) == 3:
		// Extract the command and arguments
		arg = args[1]
		arg2 = args[2]
	case command == fmt.Sprintf("%stts", prefix):
		break
	case len(args) > 3:
		return
	}
	if !isDM {
		// we only want to enable the below commands in Servers
		switch {
		case command == fmt.Sprintf("%shelp", prefix):
			// show help text
			message := fmt.Sprintf("🧐  Help:\n"+
				"> » **Sounds**\t\t\t\t%slist\n"+
				"> » **Youtube Audio**\t%syoutube\n"+
				"> » **Text2Speech**\t%stts\n"+
				"> » **Statistics**\t\t  %sstats\n"+
				"> » **Favorites**\t\t  %sfav\n", prefix, prefix, prefix, prefix, prefix)
			utils.NewMessageRoutine(command, message, s, m)
			s.MessageReactionAdd(m.ChannelID, m.ID, "✅")
			return
		case command == fmt.Sprintf("%scleanup", prefix):
			// Cleanup all messages
			err := HandleCleanUp(s, m, arg, command)
			if err != nil {
				logger.ErrorLog.Println("error handling cleanup:", err)
			}
		case strings.HasPrefix(command, fmt.Sprintf("%syoutube", prefix)):
			// Play the sound of a youtube video
			err := HandleYoutube(s, m, arg, command)
			if err != nil {
				logger.ErrorLog.Println("error handling youtube audio:", err)
				s.MessageReactionAdd(m.ChannelID, m.ID, "❌")
				return
			}
		case strings.HasPrefix(command, fmt.Sprintf("%sstats", prefix)):
			// Handle statistics
			err := HandleStatistics(s, m, arg, command)
			if err != nil {
				logger.ErrorLog.Println("error handling statistics:", err)
				s.MessageReactionAdd(m.ChannelID, m.ID, "❌")
			}
			s.MessageReactionAdd(m.ChannelID, m.ID, "✅")
		case strings.HasPrefix(command, fmt.Sprintf("%sfav", prefix)):
			// Handle favorite sounds
			err := HandleFavorite(s, m, arg, arg2, command)
			if err != nil {
				logger.ErrorLog.Println("error handling favorites:", err)
			}
			return
		case command == fmt.Sprintf("%stts", prefix):
			// Text2Speech
			err := HandleTTS(s, m, command)
			if err != nil {
				logger.ErrorLog.Println("error handling TTS:", err)
			}
		case command == fmt.Sprintf("%slist", prefix):
			// List categories
			err := HandleList(s, m, arg, command)
			if err != nil {
				logger.ErrorLog.Println("error handling list:", err)
				s.MessageReactionAdd(m.ChannelID, m.ID, "❌")
				return
			}
			s.MessageReactionAdd(m.ChannelID, m.ID, "✅")
			return
		case command == fmt.Sprintf("%sdm", prefix):
			// Delete the authors message
			err := s.ChannelMessageDelete(m.ChannelID, m.ID)
			if err != nil {
				logger.ErrorLog.Println("error deleting message:", err)
			}

			// Send a DM to the author
			err = utils.NewPrivateMessageRoutine("Sie haben gerufen?", s, m)
			if err != nil {
				logger.ErrorLog.Println("error sending DM:", err)
			}
		default:
			return
		}
	} else {
		switch {
		case command == fmt.Sprintf("%shelp", prefix):
			// show help text
			message := fmt.Sprintf("🧐  Help:\n"+
				"> » **Statistics**\t\t  %sstats\n"+
				"> » **Favorites**\t\t  %sfav\n", prefix, prefix)

			if utils.IsAdmin(m.Author.ID) {
				message = fmt.Sprintf("🧙🏻‍♂️  Help:\n"+
					"> » **Statistics**\t\t  %sstats\n"+
					"> » **Favorites**\t\t  %sfav\n"+
					"> » **Users**\t\t  %susers\n"+
					"> » **Jail**\t\t  %sjail\n", prefix, prefix, prefix, prefix)
			}

			utils.NewPrivateMessageRoutine(message, s, m)
			return
		case strings.HasPrefix(command, fmt.Sprintf("%sstats", prefix)):
			// Handle statistics
			err := HandleStatistics(s, m, arg, command)
			if err != nil {
				s.MessageReactionAdd(m.ChannelID, m.ID, "❌")
			}
			s.MessageReactionAdd(m.ChannelID, m.ID, "✅")
			return
		case strings.HasPrefix(command, fmt.Sprintf("%sfav", prefix)):
			// Handle favorite sounds
			err := HandleFavorite(s, m, arg, arg2, command)
			if err != nil {
				logger.ErrorLog.Println("error handling favorites:", err)
			}
			return
		case strings.HasPrefix(command, fmt.Sprintf("%susers", prefix)):
			// Handle jail
			err := HandleUsers(s, m, command)
			if err != nil {
				logger.ErrorLog.Println("error handling users:", err)
			}
		case strings.HasPrefix(command, fmt.Sprintf("%sjail", prefix)):
			// Handle jail
			err := HandleJail(s, m, arg, arg2, command)
			if err != nil {
				logger.ErrorLog.Println("error handling jail:", err)
			}
		default:
			return
		}
	}
}
