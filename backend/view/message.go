package view

import (
	"fmt"
	"time"

	"github.com/cyb3rplis/discord-bot-go/logger"

	"github.com/bwmarrin/discordgo"
)

func (a *API) SendMessageComplex(msg *discordgo.MessageSend, s *discordgo.Session, mc *discordgo.MessageCreate, delete bool) (*discordgo.Message, error) {
	// send complex message
	message, err := s.ChannelMessageSendComplex(mc.ChannelID, msg)
	if err != nil {
		logger.ErrorLog.Println("error sending message:", err)
		return nil, err
	}
	if delete {
		// delete message again to keep the channel clean
		time.Sleep(5 * time.Second)
		err = s.ChannelMessageDelete(message.ChannelID, message.ID)
		if err != nil {
			logger.ErrorLog.Println("error deleting message:", err)
		}
	}
	return message, nil
}

func (a *API) SendMessage(msg string, s *discordgo.Session, mc *discordgo.MessageCreate, delete bool) (*discordgo.Message, error) {
	// send message
	message, err := s.ChannelMessageSend(mc.ChannelID, msg)
	if err != nil {
		logger.ErrorLog.Println("error sending message:", err)
		return nil, err
	}
	if delete {
		// delete message again to keep the channel clean
		time.Sleep(5 * time.Second)
		err = s.ChannelMessageDelete(message.ChannelID, message.ID)
		if err != nil {
			logger.ErrorLog.Println("error deleting message:", err)
		}
	}
	return message, nil
}

func (a *API) SendHiddenMessage(msg string, i *discordgo.InteractionCreate, s *discordgo.Session, delete bool) error {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		return fmt.Errorf("error responding to interaction: %v", err)
	}
	if delete {
		go func() {
			time.Sleep(10 * time.Second)
			err := s.InteractionResponseDelete(i.Interaction)
			if err != nil {
				logger.ErrorLog.Println("error deleting hidden message:", err)
			}
		}()
	}
	return nil
}
