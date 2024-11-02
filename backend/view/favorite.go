package view

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/logger"
	"github.com/cyb3rplis/discord-bot-go/model"
)

func HandleFavorite(s *discordgo.Session, m *discordgo.MessageCreate, arg, arg2, command string) error {
	switch arg {
	case "add":
		// check if sound exists
		soundID, _ := model.GetFavoriteByNameAndUserID(arg2, m.Author.ID)
		if soundID != "" {
			_ = logger.ReactionLogSuccess(s, m, "sound already in favorites", "")
			return nil
		}
		// add sound to favorites
		err := model.SoundFavoriteAdd(m, arg2)
		if err != nil {
			_ = logger.ReactionLogError(s, m, "error adding sound to favorites", err)
			return err
		}
		_ = logger.ReactionLogSuccess(s, m, "sound added to favorites", "")
		return nil
	case "rm":
		// remove sound from favorites
		err := model.SoundFavoriteRemove(m, arg2)
		if err != nil {
			_ = logger.ReactionLogError(s, m, "error removing sound from favorites", err)
			return err
		}
		_ = logger.ReactionLogSuccess(s, m, "sound removed from favorites", "")
		return nil
	case "list":
		favorites, err := model.GetUserFavorites(m.Author.ID)
		if err != nil {
			logger.ErrorLog.Printf("error getting user favorites: %v", err)
		}
		if len(favorites) == 0 {
			_ = logger.ReactionLogSuccessWithFeedback(s, m, "No favorites in your list", "")
			return nil
		}
		var soundNames []string
		for _, favorite := range favorites {
			soundNames = append(soundNames, favorite.SoundName)
		}
		// Build buttons for the favorite sounds
		buttons := model.BuildSoundButtons(soundNames, "favorites", discordgo.SuccessButton)
		// Build messages for the favorite sounds
		initialMessage := &discordgo.MessageSend{
			Content: "Favourites of <@" + m.Author.ID + ">",
		}

		messages := model.BuildMessages(buttons, initialMessage)

		for i, message := range messages {
			NewComplexMessageRoutine(command+arg+fmt.Sprint(i)+m.Author.ID, m.ChannelID, m.ID, message, s)
		}
		_ = logger.ReactionLogSuccess(s, m, "favorites listed", "")
		return nil
	default:
		message := fmt.Sprintf("🔥  Your favorites helper:\n" +
			"> » **List sounds**\t\t" + model.Bot.Config.Prefix + "list\n" +
			"> » **Add sound**\t\t " + model.Bot.Config.Prefix + "add <sound_name>\n" +
			"> » **Remove sound**\t\t " + model.Bot.Config.Prefix + "rm <sound_name>\n")
		_ = logger.ReactionLogSuccess(s, m, "favorites help message sent", "")
		NewMessageRoutine(command+"help", message, s, m)
	}
	return nil
}
