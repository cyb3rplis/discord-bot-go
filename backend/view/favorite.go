package view

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/dlog"
	"github.com/cyb3rplis/discord-bot-go/model"
)

func (a *API) PromptInteractionFavorite(s *discordgo.Session, i *discordgo.InteractionCreate) {
	//get userID
	if i.Member == nil {
		dlog.ErrorLog.Println("error getting member from interaction")
		return
	}
	if i.Type == discordgo.InteractionApplicationCommand {
		switch i.ApplicationCommandData().Name {
		case "favorite":
			option := i.ApplicationCommandData().Options[0]
			switch option.Name {
			case "list":
				err := a.SendInteractionRespond("👉 Listing favorites", s, i)
				if err != nil {
					dlog.ErrorLog.Printf("error sending message: %v", err)
				}
				// Check if the user is in the Gulag
				user, err := a.model.GetUserFromUsername(i.Member.User.GlobalName)
				if err != nil {
					dlog.ErrorLog.Println("error getting user from username:", err)
				}
				if remaining, ok := IsUserInGulag(user); ok {
					user.Remaining = remaining
					message := fmt.Sprintf("<@"+user.ID+"> you are in the Gulag for another %s", user.Remaining)
					_, err = a.SendMessage(message, s, i, false)
					if err != nil {
						dlog.ErrorLog.Printf("error sending message: %v", err)
					}
					return
				}
				favorites, err := a.model.GetUserFavorites(i.Member.User.ID)
				if err != nil {
					dlog.ErrorLog.Printf("error getting user favorites: %v", err)
				}
				if len(favorites) == 0 {
					dlog.ErrorLog.Println("No favorites in your list")
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
					Content: "Favourites of <@" + i.Member.User.ID + ">",
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
					soundID, _ := a.model.GetFavoriteByNameAndUserID(arg, i.Member.User.ID)
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
					_, err = a.SendMessage(fmt.Sprintf("Sound %s has been removed from your favorites", arg), s, i, false)
					if err != nil {
						dlog.ErrorLog.Printf("error sending message: %v", err)
					}
					return
				}
			}
		}
	}
}
