package message

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cyb3rplis/discord-bot-go/model"

	"github.com/cyb3rplis/discord-bot-go/logger"
	"github.com/cyb3rplis/discord-bot-go/utils"

	"github.com/bwmarrin/discordgo"
)

// AudioMessageHandler is created on any channel that the authenticated bot has access to.
func AudioMessageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	meta := model.Meta
	prefix := model.Bot.Config.Prefix

	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}
	if len(m.Content) == 0 { // Ignore empty messages
		logger.InfoLog.Println("Empty content in command, ignore")
		return
	}

	// add user to DB
	userID, err := strconv.Atoi(m.Author.ID)
	if err != nil {
		logger.ErrorLog.Println("error converting user ID to int:", err)
	}
	err = utils.AddUser(userID, m.Author.GlobalName)
	if err != nil {
		logger.ErrorLog.Println("error adding user to DB:", err)
	}

	args := strings.Split(m.Content, " ")
	command := args[0]
	var arg string
	var arg2 string
	var arg3 string
	switch {
	case len(args) == 2:
		// Extract the command and arguments
		arg = args[1]
	case len(args) == 3:
		// Extract the command and arguments
		arg = args[1]
		arg2 = args[2]
	case len(args) == 4:
		// Extract the command and arguments
		arg = args[1]
		arg2 = args[2]
		arg3 = args[3]
	case command == fmt.Sprintf("%stts", prefix):
		break
	case len(args) > 4:
		return
	}

	// check if the message originated from the same Guild ID as the bot's Guild ID when he was started
	if m.GuildID == meta.Guild.ID {
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
			_ = s.MessageReactionAdd(m.ChannelID, m.ID, "✅")
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
				_ = s.MessageReactionAdd(m.ChannelID, m.ID, "❌")
				return
			}
		case strings.HasPrefix(command, fmt.Sprintf("%sstats", prefix)):
			// Handle statistics
			err := HandleStatistics(s, m, arg, command)
			if err != nil {
				logger.ErrorLog.Println("error handling statistics:", err)
				_ = s.MessageReactionAdd(m.ChannelID, m.ID, "❌")
			}
			_ = s.MessageReactionAdd(m.ChannelID, m.ID, "✅")
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
				_ = s.MessageReactionAdd(m.ChannelID, m.ID, "❌")
				return
			}
			_ = s.MessageReactionAdd(m.ChannelID, m.ID, "✅")
			return
		case command == fmt.Sprintf("%sdm", prefix):
			// Delete the authors message
			err := s.ChannelMessageDelete(m.ChannelID, m.ID)
			if err != nil {
				logger.ErrorLog.Println("error deleting message:", err)
			}

			// Send a DM to the author
			err = utils.NewPrivateMessageRoutine("Ready for some action?", s, m)
			if err != nil {
				logger.ErrorLog.Println("error sending DM:", err)
			}
		default:
			return
		}
	} else {
		// This else block is only being used when the bot received a message from outside the Guild ID he initially joined
		// This will only happen if a user with the admin role DM'ed the bot
		switch {
		case command == fmt.Sprintf("%shelp", prefix):
			// show help text
			message := fmt.Sprintf("🧐  Help:\n"+
				"> » **Statistics**\t\t  %sstats\n"+
				"> » **Favorites**\t\t  %sfav\n", prefix, prefix)

			memberRoles, err := utils.GetMemberRoles(s, meta.Guild.ID, m.Author.ID)
			if err != nil {
				logger.ErrorLog.Println("error getting member roles:", err)
			}

			if utils.IsAdmin(memberRoles) {
				message = fmt.Sprintf("🧙🏻‍♂️  Help:\n"+
					"> » **Statistics**\t\t  %sstats\n"+
					"> » **Favorites**\t\t  %sfav\n"+
					"> » **Users**\t\t  %susers\n"+
					"> » **Gulag**\t\t  %sgulag\n", prefix, prefix, prefix, prefix)
			}

			utils.NewPrivateMessageRoutine(message, s, m)
			_ = s.MessageReactionAdd(m.ChannelID, m.ID, "✅")
			return
		case strings.HasPrefix(command, fmt.Sprintf("%sstats", prefix)):
			// Handle statistics
			err := HandleStatistics(s, m, arg, command)
			if err != nil {
				_ = s.MessageReactionAdd(m.ChannelID, m.ID, "❌")
			}
			_ = s.MessageReactionAdd(m.ChannelID, m.ID, "✅")
			return
		case strings.HasPrefix(command, fmt.Sprintf("%sfav", prefix)):
			// Handle favorite sounds
			err := HandleFavorite(s, m, arg, arg2, command)
			if err != nil {
				logger.ErrorLog.Println("error handling favorites:", err)
			}
			return
		case strings.HasPrefix(command, fmt.Sprintf("%susers", prefix)):
			// Handle gulag
			err := HandleUsers(s, m, command)
			if err != nil {
				logger.ErrorLog.Println("error handling users:", err)
				_ = s.MessageReactionAdd(m.ChannelID, m.ID, "❌")
			} else {
				_ = s.MessageReactionAdd(m.ChannelID, m.ID, "✅")
			}
		case strings.HasPrefix(command, fmt.Sprintf("%sgulag", prefix)):
			// Handle gulag
			err := HandleGulag(s, m, arg, arg2, arg3, command)
			if err != nil {
				logger.ErrorLog.Println("error handling gulag:", err)
				_ = s.MessageReactionAdd(m.ChannelID, m.ID, "❌")
				return
			} else {
				_ = s.MessageReactionAdd(m.ChannelID, m.ID, "✅")
			}
		default:
			return
		}
	}
}
