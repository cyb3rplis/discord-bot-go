package message

import (
	"database/sql"
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/logger"
	"github.com/cyb3rplis/discord-bot-go/model"
	"github.com/cyb3rplis/discord-bot-go/sound"
	"github.com/cyb3rplis/discord-bot-go/utils"
)

type Favorite struct {
	ID           string
	UserID       string
	SoundID      string
	SoundName    string
	CategoryID   string
	CategoryName string
}

func HandleFavorite(s *discordgo.Session, m *discordgo.MessageCreate, arg, arg2, command string) error {
	switch arg {
	case "add":
		// check if sound exists
		soundID, _ := sound.GetFavoriteByNameAndUserID(arg2, m.Author.ID)
		if soundID != "" {
			_ = logger.ReactionLogSuccess(s, m, "sound already in favorites", "")
			return nil
		}
		// add sound to favorites
		err := SoundFavoriteAdd(m, arg2)
		if err != nil {
			_ = logger.ReactionLogError(s, m, "error adding sound to favorites", err)
			return err
		}
		_ = logger.ReactionLogSuccess(s, m, "sound added to favorites", "")
		return nil
	case "rm":
		// remove sound from favorites
		err := SoundFavoriteRemove(m, arg2)
		if err != nil {
			_ = logger.ReactionLogError(s, m, "error removing sound from favorites", err)
			return err
		}
		_ = logger.ReactionLogSuccess(s, m, "sound removed from favorites", "")
		return nil
	case "list":
		favorites, err := GetUserFavorites(m.Author.ID)
		if err != nil {
			logger.ErrorLog.Printf("error getting user favorites: %v", err)
		}
		if len(favorites) == 0 {
			_ = logger.ReactionLogError(s, m, "no favorites found", nil)
			return nil
		}
		var soundNames []string
		for _, favorite := range favorites {
			soundNames = append(soundNames, favorite.SoundName)
		}
		// Build buttons for the favorite sounds
		buttons := utils.BuildSoundButtons(soundNames, "favorites", discordgo.SuccessButton)
		// Build messages for the favorite sounds
		initialMessage := &discordgo.MessageSend{
			Content: "Favourites of <@" + m.Author.ID + ">",
		}

		messages := utils.BuildMessages(buttons, initialMessage)

		for i, message := range messages {
			utils.NewComplexMessageRoutine(command+arg+fmt.Sprint(i)+m.Author.ID, m.ChannelID, m.ID, message, s)
		}
		_ = logger.ReactionLogSuccess(s, m, "favorites listed", "")
		return nil
	default:
		message := fmt.Sprintf("🔥  Your favorites helper:\n" +
			"> » **List sounds**\t\t" + model.Bot.Config.Prefix + "list\n" +
			"> » **Add sound**\t\t " + model.Bot.Config.Prefix + "add <sound_name>\n" +
			"> » **Remove sound**\t\t " + model.Bot.Config.Prefix + "rm <sound_name>\n")
		_ = logger.ReactionLogSuccess(s, m, "favorites help message sent", "")
		utils.NewMessageRoutine(command+"help", message, s, m)
	}
	return nil
}

func GetUserFavorites(userID string) ([]Favorite, error) {
	rows, err := model.Bot.Db.Query("SELECT user_favorites.id, user_favorites.user_id, user_favorites.sound_id, sounds.name, categories.id, categories.name "+
		"FROM user_favorites "+
		"LEFT JOIN sounds ON sounds.id = user_favorites.sound_id "+
		"LEFT JOIN categories ON categories.id = sounds.category_id WHERE user_id = ?", userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query favorites: %w", err)
	}
	defer rows.Close()

	var favorites []Favorite
	for rows.Next() {
		fav := Favorite{}
		var id, userID, soundID, soundName, categoryID, categoryName sql.NullString
		err := rows.Scan(&id, &userID, &soundID, &soundName, &categoryID, &categoryName)
		if err != nil {
			return nil, fmt.Errorf("failed to scan favorite: %w", err)
		}
		if id.Valid {
			fav.ID = id.String
		}
		if userID.Valid {
			fav.UserID = userID.String
		}
		if soundID.Valid {
			fav.SoundID = soundID.String
		}
		if soundName.Valid {
			fav.SoundName = soundName.String
		}
		if categoryID.Valid {
			fav.CategoryID = categoryID.String
		}
		if categoryName.Valid {
			fav.CategoryName = categoryName.String
		}
		favorites = append(favorites, fav)
	}
	return favorites, nil
}

func SoundFavoriteAdd(m *discordgo.MessageCreate, arg string) error {
	userID := m.Author.ID
	soundName := arg
	//get soundID by Name
	soundID, err := sound.GetSoundIDByName(soundName)
	if err != nil {
		return fmt.Errorf("failed to get sound by name: %w", err)
	}
	_, err = model.Bot.Db.Exec("INSERT INTO user_favorites (user_id, sound_id) VALUES (?, ?)", userID, soundID)
	if err != nil {
		return fmt.Errorf("failed to insert favorite: %w", err)
	}
	return nil
}

func SoundFavoriteRemove(m *discordgo.MessageCreate, arg string) error {
	userID := m.Author.ID
	soundName := arg
	//get soundID by Name
	soundID, err := sound.GetFavoriteByNameAndUserID(soundName, userID)
	if err != nil {
		return fmt.Errorf("failed to get sound by name: %w", err)
	}
	_, err = model.Bot.Db.Exec("DELETE FROM user_favorites WHERE id = ?", soundID)
	if err != nil {
		return fmt.Errorf("failed to delete favorite: %w", err)
	}
	return nil
}
