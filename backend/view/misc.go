package view

import (
	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/dlog"
)

func (a *API) PromptInteractionMisc(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionApplicationCommand {
		switch i.ApplicationCommandData().Name {
		case "misc":
			option := i.ApplicationCommandData().Options[0]
			switch option.Name {
			case "leave":
				err := a.SendInteractionRespond("👋  Leaving Voice Channel", s, i)
				if err != nil {
					dlog.ErrorLog.Println("leave interaction", err)
				}

				// Get the guild ID
				guildID := i.GuildID
				if guildID == "" {
					err := a.UpdateInteractionResponse("👋  Something went wrong...", s, i)

					if err != nil {
						dlog.ErrorLog.Println("leave interaction", err)
					}
					return
				}

				voiceConnection, ok := s.VoiceConnections[guildID]
				if !ok {
					dlog.ErrorLog.Println("Bot is not connected to a voice channel in this guild")
					return
				}

				// Leave the voice channel
				err = voiceConnection.Disconnect()
				if err != nil {
					dlog.ErrorLog.Printf("Error disconnecting from the voice channel: %v\n", err)
				} else {
					err := a.UpdateInteractionResponse("👋  Bye Bye", s, i)

					if err != nil {
						dlog.ErrorLog.Println("leave interaction", err)
					}
				}

			default:
				err := a.SendInteractionRespond("👋  Something went wrong...", s, i)
				if err != nil {
					dlog.ErrorLog.Println("fallback to default misc handler", err)
				}
			}
		}
	}
}
