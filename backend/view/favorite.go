package view

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/config"
	"github.com/cyb3rplis/discord-bot-go/dlog"
	"github.com/cyb3rplis/discord-bot-go/model"
)

func (a *API) PromptInteractionFavorite(s *discordgo.Session, i *discordgo.InteractionCreate) {
	interactionUser := config.ExtendedUser{
		User: i.Member.User,
	}
	if i.Type == discordgo.InteractionApplicationCommand {
		switch i.ApplicationCommandData().Name {
		case "favorite":
			option := i.ApplicationCommandData().Options[0]
			switch option.Name {
			case "buttons":
				err := a.SendInteractionRespond("👉 Listing favorites", s, i)
				if err != nil {
					dlog.ErrorLog.Printf("error sending message: %v", err)
				}

				favorites, err := a.model.GetUserFavorites(interactionUser)
				if err != nil {
					dlog.ErrorLog.Printf("error getting user favorites: %v", err)
				}
				if len(favorites) == 0 {
					err = a.UpdateInteractionResponse("No favorites in your list", s, i)
					if err != nil {
						dlog.ErrorLog.Printf("error sending message: %v", err)
					}
					return
				}
				var soundNames []string
				for _, favorite := range favorites {
					soundNames = append(soundNames, favorite.SoundName)
				}
				// Build buttons for the favorite sounds
				buttons := model.BuildSoundButtons(soundNames, "favorites", discordgo.SuccessButton)
				// Build messages for the favorite sounds
				initialMessage := &discordgo.MessageSend{
					Content: "Favourites of " + interactionUser.User.Mention(),
				}
				messages := model.BuildMessages(buttons, initialMessage)
				for _, message := range messages {
					_, err = a.SendMessageComplex(message, s, i, false)
					if err != nil {
						dlog.ErrorLog.Printf("error sending message: %v", err)
					}
				}
			case "add":
				if len(option.Options) > 0 {
					err := a.SendInteractionRespond("👉 Adding sound to favorites", s, i)
					if err != nil {
						dlog.ErrorLog.Printf("error sending message: %v", err)
					}
					arg := option.Options[0].StringValue()
					// check if sound exists
					soundID, _ := a.model.GetFavoriteByNameAndUserID(arg, interactionUser)
					if soundID != "" {
						err := a.UpdateInteractionResponse("sound already in favorites", s, i)
						if err != nil {
							dlog.ErrorLog.Printf("error sending message: %v", err)
						}
						return
					}
					// add sound to favorites
					err = a.model.SoundFavoriteAdd(i, arg)
					if err != nil {
						err := a.UpdateInteractionResponse("error adding sound to favorites", s, i)
						if err != nil {
							dlog.ErrorLog.Printf("error sending message: %v", err)
						}
						return
					}
					err = a.UpdateInteractionResponse(fmt.Sprintf("Sound %s has been added to your favorites", arg), s, i)
					if err != nil {
						dlog.ErrorLog.Printf("error sending message: %v", err)
					}
					return
				}
			case "remove":
				if len(option.Options) > 0 {
					err := a.SendInteractionRespond("👉 Removing sound from favorites", s, i)
					if err != nil {
						dlog.ErrorLog.Printf("error sending message: %v", err)
					}
					arg := option.Options[0].StringValue()
					// remove sound from favorites
					err = a.model.SoundFavoriteRemove(i, arg)
					if err != nil {
						err := a.UpdateInteractionResponse("error removing sound from favorites", s, i)
						if err != nil {
							dlog.ErrorLog.Printf("error sending message: %v", err)
						}
						return
					}
					err = a.UpdateInteractionResponse(fmt.Sprintf("Sound %s has been removed from your favorites", arg), s, i)
					if err != nil {
						dlog.ErrorLog.Printf("error sending message: %v", err)
					}
					return
				}
			default:
				err := a.SendInteractionRespond("👉  Something went wrong...", s, i)
				if err != nil {
					dlog.ErrorLog.Println("fallback to default favorite handler", err)
				}
			}
		}
	}
}
