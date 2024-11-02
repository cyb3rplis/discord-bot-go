package logger

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"log"
	"os"
)

var (
	WarningLog *log.Logger
	InfoLog    *log.Logger
	ErrorLog   *log.Logger
	FatalLog   *log.Logger
)

func init() {
	InfoLog = log.New(os.Stdout, "INFO ", log.Ldate|log.Ltime|log.Lmsgprefix)
	WarningLog = log.New(os.Stdout, "WARNING ", log.Ldate|log.Ltime|log.Lmsgprefix)
	ErrorLog = log.New(os.Stderr, "ERROR ", log.Ldate|log.Ltime|log.Lmsgprefix)
	FatalLog = log.New(os.Stderr, "FATAL ", log.Ldate|log.Ltime|log.Lmsgprefix|log.Lshortfile)
}

func ReactionLogError(s *discordgo.Session, m *discordgo.MessageCreate, errorMessage string, err error) error {
	ErrorLog.Println(errorMessage, err)
	err = s.MessageReactionAdd(m.ChannelID, m.ID, "☹️")
	//send message with error to the user
	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("> %s", errorMessage))
	if err != nil {
		return err
	}
	return nil
}

func ReactionLogSuccess(s *discordgo.Session, m *discordgo.MessageCreate, message string, icon string) error {
	reactionIcon := "👍"
	if icon != "" {
		reactionIcon = icon
	}
	InfoLog.Println(message)
	err := s.MessageReactionAdd(m.ChannelID, m.ID, reactionIcon)
	if err != nil {
		return err
	}
	return nil
}

func ReactionLogSuccessWithFeedback(s *discordgo.Session, m *discordgo.MessageCreate, message string, icon string) error {
	reactionIcon := "👍"
	if icon != "" {
		reactionIcon = icon
	}
	InfoLog.Println(message)
	err := s.MessageReactionAdd(m.ChannelID, m.ID, reactionIcon)
	if err != nil {
		return err
	}
	//send message with feedback to the user
	_, err = s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("> %s", message))
	if err != nil {
		return err
	}
	return nil
}
