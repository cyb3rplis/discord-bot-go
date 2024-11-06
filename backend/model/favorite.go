package model

import (
	"database/sql"
	"fmt"
	"github.com/bwmarrin/discordgo"
)

type Favorite struct {
	ID           string
	UserID       string
	SoundID      string
	SoundName    string
	CategoryID   string
	CategoryName string
}

func (m *Model) GetUserFavorites(userID string) ([]Favorite, error) {
	rows, err := m.Db.Query("SELECT user_favorites.id, user_favorites.user_id, user_favorites.sound_id, sounds.name, categories.id, categories.name "+
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

func (m *Model) SoundFavoriteAdd(i *discordgo.InteractionCreate, arg string) error {
	userID := i.Member.User.ID
	soundName := arg
	//get soundID by Name
	soundID, err := m.GetSoundIDByName(soundName)
	if err != nil {
		return fmt.Errorf("failed to get sound by name: %w", err)
	}
	_, err = m.Db.Exec("INSERT INTO user_favorites (user_id, sound_id) VALUES (?, ?)", userID, soundID)
	if err != nil {
		return fmt.Errorf("failed to insert favorite: %w", err)
	}
	return nil
}

func (m *Model) SoundFavoriteRemove(i *discordgo.InteractionCreate, arg string) error {
	userID := i.Member.User.ID
	soundName := arg
	//get soundID by Name
	soundID, err := m.GetFavoriteByNameAndUserID(soundName, userID)
	if err != nil {
		return fmt.Errorf("failed to get sound by name: %w", err)
	}
	_, err = m.Db.Exec("DELETE FROM user_favorites WHERE id = ?", soundID)
	if err != nil {
		return fmt.Errorf("failed to delete favorite: %w", err)
	}
	return nil
}

func (m *Model) GetFavoriteByNameAndUserID(name, userID string) (soundName string, err error) {
	err = m.Db.QueryRow("SELECT user_favorites.id FROM user_favorites LEFT JOIN sounds s on user_favorites.sound_id = s.id WHERE name = ? AND user_id = ?", name, userID).Scan(&soundName)
	if err != nil {
		return "", fmt.Errorf("failed to query favorite by name and user ID: %w", err)
	}
	return soundName, nil
}

func (m *Model) GetSoundIDByName(name string) (soundName string, err error) {
	err = m.Db.QueryRow("SELECT id FROM sounds WHERE name = ?", name).Scan(&soundName)
	if err != nil {
		return "", fmt.Errorf("failed to query sound by alias: %w", err)
	}
	return soundName, nil
}
