package view

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cyb3rplis/discord-bot-go/model"

	"github.com/cyb3rplis/discord-bot-go/logger"

	"github.com/bwmarrin/discordgo"
)

// AudioMessageHandler is created on any channel that the authenticated bot has access to.
func (a *API) AudioMessageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	prefix := a.model.Config.Prefix

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
	err = a.model.AddUser(userID, m.Author.GlobalName)
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
	meta := model.Meta
	// check if the message originated from the same Guild ID as the bot's Guild ID when he was started
	if m.GuildID == meta.Guild.ID {
		// we only want to enable the below commands in Servers
		switch {
		case command == fmt.Sprintf("%shelp", prefix):
			// show help text
			message := fmt.Sprintf("🧐  Help:\n"+
				"> » **Sounds**\t\t\t\t%slist\n"+
				"> » **Audio**\t\t\t\t%saudio\n"+
				"> » **Text2Speech**\t%stts\n"+
				"> » **Statistics**\t\t  %sstats\n"+
				"> » **Favorites**\t\t  %sfav\n", prefix, prefix, prefix, prefix, prefix)
			a.NewMessageRoutine(command, message, s, m)
			_ = logger.ReactionLogSuccess(s, m, "help message sent", "")
			return
		case command == fmt.Sprintf("%scleanup", prefix):
			// Cleanup all messages
			err := a.HandleCleanUp(s, m, arg, command)
			if err != nil {
				logger.ErrorLog.Println("error handling cleanup:", err)
			}
		case strings.HasPrefix(command, fmt.Sprintf("%saudio", prefix)):
			// Play the sound of a video streaming platform
			err := a.HandleAudio(s, m, arg, command)
			if err != nil {
				_ = logger.ReactionLogError(s, m, "error handling audio", err)
				return
			}
		case strings.HasPrefix(command, fmt.Sprintf("%sstats", prefix)):
			// Handle statistics
			err := a.HandleStatistics(s, m, arg, command)
			if err != nil {
				_ = logger.ReactionLogError(s, m, "error handling statistics", err)
			}
			_ = logger.ReactionLogSuccess(s, m, "statistics action successful", "")
		case strings.HasPrefix(command, fmt.Sprintf("%sfav", prefix)):
			// Handle favorite sounds
			err := a.HandleFavorite(s, m, arg, arg2, command)
			if err != nil {
				logger.ErrorLog.Println("error handling favorites:", err)
			}
			return
		case command == fmt.Sprintf("%stts", prefix):
			// Text2Speech
			err := a.HandleTTS(s, m, command)
			if err != nil {
				logger.ErrorLog.Println("error handling TTS:", err)
			}
		case command == fmt.Sprintf("%slist", prefix):
			// List categories
			err := a.HandleList(s, m, arg, command)
			if err != nil {
				_ = logger.ReactionLogError(s, m, "error handling list", err)
				return
			}
			_ = logger.ReactionLogSuccess(s, m, "list action successful", "")
			return
		case command == fmt.Sprintf("%sdm", prefix):
			// Delete the authors message
			err := s.ChannelMessageDelete(m.ChannelID, m.ID)
			if err != nil {
				logger.ErrorLog.Println("error deleting message:", err)
			}

			// Send a DM to the author
			err = NewPrivateMessageRoutine("Ready for some action?", s, m)
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

			memberRoles, err := model.GetMemberRoles(s, meta.Guild.ID, m.Author.ID)
			if err != nil {
				logger.ErrorLog.Println("error getting member roles:", err)
			}

			if a.model.IsAdmin(memberRoles) {
				message = fmt.Sprintf("🧙🏻‍♂️  Help:\n"+
					"> » **Statistics**\t\t  %sstats\n"+
					"> » **Favorites**\t\t  %sfav\n"+
					"> » **Users**\t\t  %susers\n"+
					"> » **Gulag**\t\t  %sgulag\n", prefix, prefix, prefix, prefix)
			}

			err = NewPrivateMessageRoutine(message, s, m)
			if err != nil {
				logger.ErrorLog.Println("error sending DM:", err)
			}
			_ = logger.ReactionLogSuccess(s, m, "help message sent", "")
			return
		case strings.HasPrefix(command, fmt.Sprintf("%sstats", prefix)):
			// Handle statistics
			err := a.HandleStatistics(s, m, arg, command)
			if err != nil {
				_ = logger.ReactionLogError(s, m, "error handling statistics", err)
				return
			}
			_ = logger.ReactionLogSuccess(s, m, "statistics action successful", "")
			return
		case strings.HasPrefix(command, fmt.Sprintf("%sfav", prefix)):
			// Handle favorite sounds
			err := a.HandleFavorite(s, m, arg, arg2, command)
			if err != nil {
				logger.ErrorLog.Println("error handling favorites:", err)
			}
			return
		case strings.HasPrefix(command, fmt.Sprintf("%susers", prefix)):
			// Handle gulag
			err := a.HandleUsers(s, m, command)
			if err != nil {
				_ = logger.ReactionLogError(s, m, "error handling users", err)
			} else {
			}
			_ = logger.ReactionLogSuccess(s, m, "users action successful", "")
		case strings.HasPrefix(command, fmt.Sprintf("%sgulag", prefix)):
			// Handle gulag
			err := a.HandleGulag(s, m, arg, arg2, arg3, command)
			if err != nil {
				_ = logger.ReactionLogError(s, m, "error handling gulag", err)
				return
			}
			_ = logger.ReactionLogSuccess(s, m, "gulag action successful", "")
		default:
			return
		}
	}
}

func (a *API) NewMessageRoutine(command, message string, s *discordgo.Session, m *discordgo.MessageCreate) (st *discordgo.Message) {
	// send our new message
	st, err := s.ChannelMessageSend(m.ChannelID, message)
	if err != nil {
		logger.ErrorLog.Println("error sending message:", err)
	}

	// get all old messages for this command
	oldMessages, err := a.model.GetAllCommandMessages(command)
	if err != nil {
		logger.ErrorLog.Println("error getting all command messages:", err)
	}

	// iterate over all the old messages and delete them from discord
	for cID, mID := range oldMessages {
		for _, msg := range mID {
			err := s.ChannelMessageDelete(cID, msg)
			if err != nil {
				logger.ErrorLog.Printf("error deleting old message - ID: %s, err: %v", msg, err)
				err = a.model.DeleteMessageID(msg)
				if err != nil {
					logger.ErrorLog.Printf("error deleting message from DB: %v", err)
				}
			}
		}
	}

	// insert the new message id into the database
	err = a.model.InsertMessageID(st.ChannelID, st.ID, command)
	if err != nil {
		logger.ErrorLog.Println("error inserting message id:", err)
	}

	err = a.model.DeleteOldCommandMessages(st.ID, command)
	if err != nil {
		logger.ErrorLog.Println("error deleting old message:", err)
	}

	return st
}

func NewPrivateMessageRoutine(message string, s *discordgo.Session, m *discordgo.MessageCreate) (err error) {
	// Send a DM to the author
	channel, err := s.UserChannelCreate(m.Author.ID)
	if err != nil {
		return fmt.Errorf("error creating private channel with user: %s", m.Author.ID)
	}
	_, err = s.ChannelMessageSend(channel.ID, message)
	if err != nil {
		return fmt.Errorf("error sending private message: %v", err)
	}

	return nil
}

func (a *API) NewComplexMessageRoutine(command, channelID, msgID string, msg *discordgo.MessageSend, s *discordgo.Session) (st *discordgo.Message) {
	// send our new message
	st, err := s.ChannelMessageSendComplex(channelID, msg)
	if err != nil {
		logger.ErrorLog.Println("error sending message:", err)
	}

	// get all old messages for this command
	oldMessages, err := a.model.GetAllCommandMessages(command)
	if err != nil {
		logger.ErrorLog.Println("error getting all command messages:", err)
	}

	// iterate over all the old messages and delete them from discord
	for cID, mID := range oldMessages {
		for _, msg := range mID {
			err := s.ChannelMessageDelete(cID, msg)
			if err != nil {
				logger.ErrorLog.Printf("error deleting old message - ID: %s, err: %v", msg, err)
				err = a.model.DeleteMessageID(msg)
				if err != nil {
					logger.ErrorLog.Printf("error deleting message from DB: %v", err)
				}
			}
		}
	}

	// insert the new message id into the database
	err = a.model.InsertMessageID(st.ChannelID, st.ID, command)
	if err != nil {
		logger.ErrorLog.Println("error inserting message id:", err)
	}

	err = a.model.DeleteOldCommandMessages(st.ID, command)
	if err != nil {
		logger.ErrorLog.Println("error deleting old message:", err)
	}
	return st
}

func (a *API) DeleteMessageRoutine(s *discordgo.Session, command string) {
	oldMessages, err := a.model.GetAllCommandMessages(command)
	if err != nil {
		logger.ErrorLog.Println("error getting all command messages:", err)
	}

	// iterate over all the old messages and delete them from discord
	for cID, mID := range oldMessages {
		for _, msg := range mID {
			err := s.ChannelMessageDelete(cID, msg)
			if err != nil {
				logger.ErrorLog.Printf("error deleting old message - ID: %s - err: %v", msg, err)
			}
		}
	}

	err = a.model.DeleteAllCommandMessages(command)
	if err != nil {
		logger.ErrorLog.Println("error deleting all command messages:", err)
	}
}
