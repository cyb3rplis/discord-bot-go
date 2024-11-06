package view

import (
	"fmt"
	"time"

	"github.com/cyb3rplis/discord-bot-go/dlog"

	"github.com/bwmarrin/discordgo"
)

func (a *API) SendMessageComplex(msg *discordgo.MessageSend, s *discordgo.Session, i *discordgo.InteractionCreate, delete bool) (*discordgo.Message, error) {
	// send complex message
	message, err := s.ChannelMessageSendComplex(i.ChannelID, msg)
	if err != nil {
		dlog.ErrorLog.Println("error sending message:", err)
		return nil, err
	}
	if delete {
		// delete message again to keep the channel clean
		time.Sleep(5 * time.Second)
		err = s.ChannelMessageDelete(message.ChannelID, message.ID)
		if err != nil {
			dlog.ErrorLog.Println("error deleting message:", err)
		}
	}
	return message, nil
}

func (a *API) SendMessage(msg string, s *discordgo.Session, i *discordgo.InteractionCreate, delete bool) (*discordgo.Message, error) {
	// send message
	message, err := s.ChannelMessageSend(i.ChannelID, msg)
	if err != nil {
		dlog.ErrorLog.Println("error sending message:", err)
		return nil, err
	}
	if delete {
		// delete message again to keep the channel clean
		time.Sleep(5 * time.Second)
		err = s.ChannelMessageDelete(message.ChannelID, message.ID)
		if err != nil {
			dlog.ErrorLog.Println("error deleting message:", err)
		}
	}
	return message, nil
}

func (a *API) SendInteractionRespond(msg string, s *discordgo.Session, i *discordgo.InteractionCreate, hidden bool) error {
	var err error
	if hidden {

		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: msg,
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
	} else {
		err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: msg,
			}})
	}
	if err != nil {
		return fmt.Errorf("error responding to interaction: %v", err)
	}

	return nil
}
