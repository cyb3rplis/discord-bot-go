package view

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/logger"
	"log"
)

func (a *API) PromptInteractionAudio(s *discordgo.Session, i *discordgo.InteractionCreate) {
	//get userID
	if i.Member == nil {
		logger.ErrorLog.Println("error getting member from interaction")
		return
	}
	m := &discordgo.MessageCreate{
		Message: &discordgo.Message{
			ID:        i.ID,
			ChannelID: i.ChannelID,
			Author:    &discordgo.User{ID: i.Member.User.ID},
		},
	}
	if i.Type == discordgo.InteractionApplicationCommand {
		switch i.ApplicationCommandData().Name {
		case "audio":
			arg := i.ApplicationCommandData().Options[0].StringValue()
			err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: "➡ Currently Playing audio by <@" + i.Member.User.Username + "> ",
				},
			})
			if err != nil {
				log.Printf("error executing audio command: %v", err)
			}

			// Check if the user is in the Gulag
			user, err := a.model.GetUserFromUsername(i.Member.User.GlobalName)
			if err != nil {
				logger.ErrorLog.Println("error getting user from username:", err)
			} else {
				if remaining, ok := IsUserInGulag(user); ok {
					user.Remaining = remaining
					message := fmt.Sprintf("<@"+user.ID+"> you are in the Gulag for another %s", user.Remaining)
					_, err = a.SendMessage(message, s, m, true)
					if err != nil {
						logger.ErrorLog.Printf("error sending message: %v", err)
					}
					return
				}
			}
			// Check if the user is in a voice channel
			err = a.VoiceChannelCheck(s, &discordgo.MessageCreate{Message: &discordgo.Message{ID: i.ID, ChannelID: i.ChannelID, Author: i.Member.User}})
			if err != nil {
				logger.ErrorLog.Println("error checking voice channel:", err)
				return
			}
			// Download and convert the audio
			download := Download{URL: arg, Start: "", End: "", Category: "audio", SoundName: "audio"}
			err = a.DownloadAndConvertAudio(download, s, &discordgo.MessageCreate{Message: &discordgo.Message{ID: i.ID, ChannelID: i.ChannelID}})
			if err != nil {
				logger.ErrorLog.Println("error loading audio:", err)
				return
			}
			// Play the custom audio
			err = a.PlayAudio(s, &discordgo.MessageCreate{Message: &discordgo.Message{ID: i.ID, ChannelID: i.ChannelID, Author: i.Member.User}})
			if err != nil {
				logger.ErrorLog.Println("error playing audio:", err)
			}
		}
	}
}
