package view

import (
	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/dlog"
	"github.com/cyb3rplis/discord-bot-go/model"
)

func (a *API) PromptInteractionButtons(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionApplicationCommand {
		switch i.ApplicationCommandData().Name {
		case "buttons":
			//case list, create:
			if len(i.ApplicationCommandData().Options) == 0 {
				err := a.handleList(s, i)
				if err != nil {
					dlog.ErrorLog.Println("error handling list:", err)
				}
			} else {
				option := i.ApplicationCommandData().Options[0]
				switch option.Name {
				case "list":
					err := a.SendInteractionRespond("👉 Listing sound category buttons", s, i)
					if err != nil {
						dlog.ErrorLog.Println("error executing buttons command:", err)
					}
					err = a.handleList(s, i)
					if err != nil {
						dlog.ErrorLog.Println("error handling buttons command:", err)
					}

				}
			}
		}
	}

}

func (a *API) handleList(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	categories, err := a.model.GetCategories()
	if err != nil {
		dlog.ErrorLog.Println("error getting categories:", err)
	}
	if len(categories) == 0 {
		_, err = a.SendMessage("No sound categories found.", s, i, false)
		if err != nil {
			dlog.ErrorLog.Println("error sending message:", err)
		}
		return err
	}
	content := model.BuildListButtons(categories, discordgo.PrimaryButton)
	messages := model.BuildMessages(content, nil)
	for _, message := range messages {
		_, err = a.SendMessageComplex(message, s, i, false)
		if err != nil {
			dlog.ErrorLog.Println("error sending message:", err)
		}
	}
	return nil
}
