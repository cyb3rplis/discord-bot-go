package view

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/dlog"
)

func (a *API) PromptInteractionAudio(s *discordgo.Session, i *discordgo.InteractionCreate) {
	interactionUser := i.Member.User

	if i.Type == discordgo.InteractionApplicationCommand {
		switch i.ApplicationCommandData().Name {
		case "audio":
			// Check if the user is in a voice channel
			_, err := a.VoiceChannelCheck(s, i)
			if err != nil {
				dlog.ErrorLog.Println("error checking voice channel:", err)
				return
			}
			err = a.SendInteractionRespond("👉 Loading audio...", s, i)
			if err != nil {
				dlog.ErrorLog.Println("error executing audio command:", err)
			}
			arg := i.ApplicationCommandData().Options[0].StringValue()

			// Check if the user is in the Gulag
			user, err := a.model.SetUserGulaggedValue(interactionUser)
			if err != nil && err != sql.ErrNoRows {
				dlog.ErrorLog.Println("error getting user from username:", err)
			} else {
				if user, ok := SetUserGulagRemaining(user); ok {
					message := fmt.Sprintf(user.User.Mention()+" you are in the Gulag for another %s", user.Remaining)
					_, err = a.SendMessage(message, s, i, true)
					if err != nil {
						dlog.ErrorLog.Printf("error sending message: %v", err)
					}
					return
				}
			}

			err = a.UpdateInteractionResponse("🎶  Preparing Audio, this might take a few seconds...", s, i)
			if err != nil {
				dlog.ErrorLog.Println("error sending message:", err)
			}
			// Download and convert the audio
			download := Download{URL: arg, Start: "", End: "", Category: "", SoundName: a.model.Config.AudioTemp}
			err = a.DownloadAudio(download, s, i)
			if err != nil {
				dlog.ErrorLog.Println("error loading audio:", err)
				return
			}
			err = a.ConvertMP3ToDCA(download.SoundName, "")
			if err != nil {
				dlog.ErrorLog.Println("error converting audio:", err)
				return
			}
			// wait for 5 seconds
			err = a.UpdateInteractionResponse("🎶  Audio is ready, playing now...", s, i)
			if err != nil {
				dlog.ErrorLog.Println("error sending message:", err)
			}
			time.Sleep(5 * time.Second)
			// Play the custom audio
			err = a.PlayAudio(download.SoundName, s, i)
			if err != nil {
				dlog.ErrorLog.Println("error playing audio:", err)
			}
		}
	}
}
