package view

import (
	"fmt"
	"strings"
	"time"

	"github.com/cyb3rplis/discord-bot-go/dlog"

	"github.com/bwmarrin/discordgo"
)

func (a *API) SendMessageComplex(msg *discordgo.MessageSend, s *discordgo.Session, i *discordgo.InteractionCreate, delete bool) (*discordgo.Message, error) {
	// send complex message
	message, err := s.ChannelMessageSendComplex(i.ChannelID, msg)
	if err != nil {
		dlog.ErrorLog.Println("error[msg1] sending message:", err)
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
		dlog.ErrorLog.Println("error[msg2] sending message:", err)
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

// SendInteractionRespond sends the initial response to an interaction
func (a *API) SendInteractionRespond(msg string, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: msg,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
	if err != nil {
		return fmt.Errorf("error SendInteractionRespond: %v", err)
	}
	return nil
}

// UpdateInteractionResponse updates the interaction response with a new message
func (a *API) UpdateInteractionResponse(msg string, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content: &msg,
	})
	if err != nil {
		return fmt.Errorf("error UpdateInteractionResponse:%v", err)
	}
	return nil
}

// UpdateInteractionResponseWithButton updates the interaction response with a new message and a buttons
func (a *API) UpdateInteractionResponseWithButton(msg string, component []discordgo.MessageComponent, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	_, err := s.InteractionResponseEdit(i.Interaction, &discordgo.WebhookEdit{
		Content:    &msg,
		Components: &component,
	})
	if err != nil {
		return fmt.Errorf("error UpdateInteractionResponseWithButton: %v", err)
	}
	return nil
}

// DeleteOldStopSoundButtons deletes all old stop sound buttons before the last one
func (a *API) DeleteOldStopSoundButtons(s *discordgo.Session, st *discordgo.Message) error {
	// Fetch messages in the channel (limit to the most recent 100)
	messages, err := s.ChannelMessages(st.ChannelID, 10, st.ID, "", "")
	if err != nil {
		dlog.ErrorLog.Printf("Error fetching messages: %v", err)
		return err
	}

	// Get the bot's user ID
	botUserID := s.State.User.ID
	var bulkDelete []string

	// Loop through each message and check for buttons from the bot with the specified custom_id
	for _, message := range messages {
		// Only process messages sent by the bot
		if message.Author.ID == botUserID && strings.HasPrefix(message.Content, "➡ Currently Playing ") {
			bulkDelete = append(bulkDelete, message.ID)
		}
	}

	if len(bulkDelete) > 0 {
		err = s.ChannelMessagesBulkDelete(st.ChannelID, bulkDelete)
		if err != nil {
			dlog.ErrorLog.Printf("Error deleting messages in bulk: %v", err)
		}
	}

	return nil
}
