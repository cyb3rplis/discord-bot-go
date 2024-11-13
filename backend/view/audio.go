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
			option := i.ApplicationCommandData().Options[0]
			switch option.Name {
			case "last":
				_, err := a.VoiceChannelCheck(s, i)
				if err != nil {
					dlog.ErrorLog.Println("error checking voice channel:", err)
					return
				}
				err = a.SendInteractionRespond("🎶  Playing last played Audio...", s, i)
				if err != nil {
					dlog.ErrorLog.Println("error executing audio command:", err)
				}

				download := Download{SoundName: a.model.Config.AudioTemp}
				err = a.PlayAudio(download.SoundName, s, i)
				if err != nil {
					dlog.ErrorLog.Println("error playing audio:", err)
				}

			case "play":
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
				url := option.Options[0].StringValue()
				dlog.InfoLog.Println("url:", url)

				// Check if the user is in the Gulag
				user, err := a.model.SetUserGulaggedValue(interactionUser)
				if err != nil && err != sql.ErrNoRows {
					dlog.ErrorLog.Println("error getting user from username:", err)
				} else {
					if user, ok := SetUserGulagRemaining(user); ok {
						message := fmt.Sprintf(user.User.Mention()+" you are in the Gulag for another %s", user.Remaining)
						_, err = a.SendMessage(message, s, i, true)
						if err != nil {
							dlog.ErrorLog.Printf("error[audio1] sending message: %v", err)
						}
						return
					}
				}

				err = a.UpdateInteractionResponse("🎶  Preparing Audio, this might take a few seconds...", s, i)
				if err != nil {
					dlog.ErrorLog.Println("error[audio2] sending message:", err)
				}
				// Download and convert the audio
				download := Download{URL: url, Start: "", End: "", Category: "", SoundName: a.model.Config.AudioTemp}
				err = a.DownloadAudio(download)
				if err != nil {
					dlog.ErrorLog.Println("error loading audio:", err)
					return
				}
				err = a.ConvertMP3ToDCA(download.SoundName, "")
				if err != nil {
					dlog.ErrorLog.Println("error converting audio:", err)
					return
				}
				// wait for 8 seconds
				err = a.UpdateInteractionResponse("🎶  Audio is ready, playing now...", s, i)
				if err != nil {
					dlog.ErrorLog.Println("error[audio3] sending message:", err)
				}
				time.Sleep(8 * time.Second)
				// Play the custom audio
				err = a.PlayAudio(download.SoundName, s, i)
				if err != nil {
					dlog.ErrorLog.Println("error playing audio:", err)
				}
			default:
				err := a.SendInteractionRespond("🎶  Something went wrong...", s, i)
				if err != nil {
					dlog.ErrorLog.Println("fallback to default audio handler", err)
				}
			}
		}
	}
}
