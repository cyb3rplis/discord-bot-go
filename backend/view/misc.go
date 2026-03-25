package view

import (
	"github.com/bwmarrin/discordgo"

	log "github.com/cyb3rplis/discord-bot-go/logger"
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
					log.ErrorLog.Println("leave interaction", err)
				}

				// Get the guild ID
				guildID := i.GuildID
				if guildID == "" {
					err := a.UpdateInteractionResponse("👋  Something went wrong...", s, i)

					if err != nil {
						log.ErrorLog.Println("leave interaction", err)
					}
					return
				}

				voiceConnection, ok := s.VoiceConnections[guildID]
				if !ok {
					log.ErrorLog.Println("Bot is not connected to a voice channel in this guild")
					return
				}

				// Leave the voice channel
				err = voiceConnection.Disconnect()
				if err != nil {
					log.ErrorLog.Printf("Error disconnecting from the voice channel: %v\n", err)
				} else {
					err := a.UpdateInteractionResponse("👋  Bye Bye", s, i)

					if err != nil {
						log.ErrorLog.Println("leave interaction", err)
					}
				}

			default:
				err := a.SendInteractionRespond("👋  Something went wrong...", s, i)
				if err != nil {
					log.ErrorLog.Println("fallback to default misc handler", err)
				}
			}
		}
	}
}
