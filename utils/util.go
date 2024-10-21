package utils

import (
	"database/sql"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/logger"
	"github.com/cyb3rplis/discord-bot-go/model"
)

// ScanDirectory scans the sound directory and returns a map of folders and files.
func ScanDirectory() (map[string][]string, error) {
	soundsRoot := model.Bot.Config.SoundsDir
	folderMap := make(map[string][]string)

	err := filepath.WalkDir(soundsRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			// Skip the root folder
			if path == soundsRoot {
				return nil
			}

			// Get relative folder name (e.g., 'folder1/')
			relativeFolder, err := filepath.Rel(soundsRoot, path)
			if err != nil {
				return err
			}

			folderMap[relativeFolder] = []string{} // Initialize an entry for this folder
		} else {
			// Add file to the folder list
			folder := filepath.Dir(path)
			relativeFolder, err := filepath.Rel(soundsRoot, folder)
			if err != nil {
				return err
			}

			// Filter for audio files based on extensions, e.g., ".dca", etc.
			if ext := filepath.Ext(path); ext == ".dca" {
				fileNameWithoutExt := RemoveFileExtension(filepath.Base(path))
				folderMap[relativeFolder] = append(folderMap[relativeFolder], fileNameWithoutExt)
			}
		}
		return nil
	})

	return folderMap, err
}

// RemoveFileExtension removes the file extension from a given file name.
func RemoveFileExtension(fileName string) string {
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}

func SortMapByValue(m map[string]int) map[string]int {
	var keys []string
	var sortedM = make(map[string]int)
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return m[keys[i]] > m[keys[j]]
	})
	for _, k := range keys {
		sortedM[k] = m[k]
	}
	return sortedM
}

func SortMapKeysByValue(m map[string]int) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return m[keys[i]] > m[keys[j]]
	})
	return keys
}

func AddUser(userID int, userName string) error {
	_, err := model.Bot.Db.Exec("INSERT INTO users (id, username) VALUES (?, ?) ON CONFLICT(id) DO UPDATE SET username = excluded.username;", userID, userName)
	if err != nil {
		return err
	}

	return nil
}

func AddUserStatistics(userID int, soundName string) error {
	_, err := model.Bot.Db.Exec("INSERT INTO stats_users (user_id, sound_id, count) VALUES (?, (SELECT id FROM sounds WHERE name = ?), 1) ON CONFLICT(user_id, sound_id) DO UPDATE SET count = count + 1;", userID, soundName)
	if err != nil {
		return err
	}

	return nil
}

func GetSoundStatistics() (soundStats map[string]int, err error) {
	rows, err := model.Bot.Db.Query("SELECT s.alias, COALESCE(SUM(su.count), 0) AS total_plays FROM sounds AS s LEFT JOIN stats_users AS su ON s.id = su.sound_id GROUP BY s.id, s.alias HAVING total_plays > 0 ORDER BY total_plays DESC LIMIT 10;")
	if err != nil {
		logger.FatalLog.Fatal(err)
	}
	defer rows.Close()

	soundStats = make(map[string]int)
	for rows.Next() {
		var sound sql.NullString
		var count sql.NullInt64

		err = rows.Scan(&sound, &count)
		if err != nil {
			logger.FatalLog.Fatal(err)
		}
		if sound.Valid && count.Valid {
			soundStats[sound.String] = int(count.Int64)
		}
	}
	//sort map by value
	soundStats = SortMapByValue(soundStats)
	return soundStats, err
}

func GetUserStatistics(userID string) (soundStats map[string]int, err error) {
	rows, err := model.Bot.Db.Query("SELECT s.alias, COALESCE(SUM(su.count), 0) AS total_plays FROM sounds AS s LEFT JOIN stats_users AS su ON s.id = su.sound_id WHERE su.user_id = ? GROUP BY s.id, s.alias HAVING total_plays > 0 ORDER BY total_plays DESC LIMIT 10;", userID)
	if err != nil {
		logger.FatalLog.Fatal(err)
	}
	defer rows.Close()

	soundStats = make(map[string]int)
	for rows.Next() {
		var sound sql.NullString
		var count sql.NullInt64

		err = rows.Scan(&sound, &count)
		if err != nil {
			logger.FatalLog.Fatal(err)
		}
		if sound.Valid && count.Valid {
			soundStats[sound.String] = int(count.Int64)
		}
	}
	//sort map by value
	soundStats = SortMapByValue(soundStats)
	return soundStats, err
}

func GetAllUserStatistics() (soundStats map[string]int, err error) {
	rows, err := model.Bot.Db.Query("SELECT u.username, COALESCE(SUM(su.count), 0) AS total_plays FROM stats_users AS su LEFT JOIN users AS u ON su.user_id = u.id GROUP BY u.id HAVING total_plays > 0 ORDER BY total_plays DESC LIMIT 10;")
	if err != nil {
		logger.FatalLog.Fatal(err)
	}
	defer rows.Close()

	soundStats = make(map[string]int)
	for rows.Next() {
		var sound sql.NullString
		var count sql.NullInt64

		err = rows.Scan(&sound, &count)
		if err != nil {
			logger.FatalLog.Fatal(err)
		}
		if sound.Valid && count.Valid {
			soundStats[sound.String] = int(count.Int64)
		}
	}
	//sort map by value
	soundStats = SortMapByValue(soundStats)
	return soundStats, err
}

func InsertMessageID(channelID, messageID, command string) error {
	_, err := model.Bot.Db.Exec("INSERT INTO messages (channel_id, message_id, command) VALUES (?, ?, ?);", channelID, messageID, command)
	if err != nil {
		return err
	}

	return nil
}

func DeleteMessageID(messageID string) error {
	_, err := model.Bot.Db.Exec("DELETE FROM messages WHERE message_id = ?;", messageID)
	if err != nil {
		return err
	}

	return nil
}

func DeleteAllCommandMessages(command string) error {
	_, err := model.Bot.Db.Exec("DELETE FROM messages WHERE command = ?;", command)
	if err != nil {
		return err
	}

	return nil
}

func DeleteOldCommandMessages(newID, command string) error {
	_, err := model.Bot.Db.Exec("DELETE FROM messages WHERE message_id != ? AND command = ?;", newID, command)
	if err != nil {
		return err
	}

	return nil
}

func DeleteAllMessages() error {
	_, err := model.Bot.Db.Exec("DELETE FROM messages;")
	if err != nil {
		return err
	}

	return nil
}

func GetAllCommandMessages(command string) (messages map[string][]string, err error) {
	rows, err := model.Bot.Db.Query("SELECT channel_id, message_id FROM messages WHERE command = ?;", command)
	if err != nil {
		logger.FatalLog.Fatal(err)
	}
	defer rows.Close()

	messages = make(map[string][]string)

	for rows.Next() {
		var cID sql.NullString
		var mID sql.NullString

		err = rows.Scan(&cID, &mID)
		if err != nil {
			logger.FatalLog.Fatal(err)
		}
		if cID.Valid && mID.Valid {
			messages[cID.String] = append(messages[cID.String], mID.String)
		}
	}

	return messages, err
}

func GetAllMessages() (messages map[string][]string, err error) {
	rows, err := model.Bot.Db.Query("SELECT channel_id, message_id FROM messages;")
	if err != nil {
		logger.FatalLog.Fatal(err)
	}
	defer rows.Close()

	messages = make(map[string][]string)

	for rows.Next() {
		var cID sql.NullString
		var mID sql.NullString

		err = rows.Scan(&cID, &mID)
		if err != nil {
			logger.FatalLog.Fatal(err)
		}
		if cID.Valid && mID.Valid {
			messages[cID.String] = append(messages[cID.String], mID.String)
		}
	}

	return messages, err
}

func NewMessageRoutine(command, message string, s *discordgo.Session, m *discordgo.MessageCreate, deleteInitial bool) {
	// delete the message from the user to keep the chat clean
	if deleteInitial {
		err := s.ChannelMessageDelete(m.ChannelID, m.ID)
		if err != nil {
			logger.ErrorLog.Println("Error deleting initial user message: ", err)
		}
	}

	// get all old messages for this command
	oldMessages, err := GetAllCommandMessages(command)
	if err != nil {
		logger.ErrorLog.Println("Error getting all command messages:", err)
	}

	// iterate over all the old messages and delete them from discord
	for cID, mID := range oldMessages {
		for _, m := range mID {
			err := s.ChannelMessageDelete(cID, m)
			if err != nil {
				logger.ErrorLog.Printf("Error deleting old message - ID: %s, err: %v", m, err)
				DeleteMessageID(m)
			}
		}
	}
	// send our new message
	st, err := s.ChannelMessageSend(m.ChannelID, message)
	if err != nil {
		logger.ErrorLog.Println("Error sending message:", err)
	}

	// insert the new message id into the database
	err = InsertMessageID(st.ChannelID, st.ID, command)
	if err != nil {
		logger.ErrorLog.Println("Error inserting message id:", err)
	}

	DeleteOldCommandMessages(st.ID, command)
}

func NewComplexMessageRoutine(command, channelID, msgID string, msg *discordgo.MessageSend, s *discordgo.Session, deleteInitial bool) (st *discordgo.Message) {
	// delete the message from the user to keep the chat clean
	if deleteInitial {
		err := s.ChannelMessageDelete(channelID, msgID)
		if err != nil {
			logger.ErrorLog.Println("Error deleting initial complex user message: ", err)
		}
	}

	// get all old messages for this command
	oldMessages, err := GetAllCommandMessages(command)
	if err != nil {
		logger.ErrorLog.Println("Error getting all command messages:", err)
	}

	// iterate over all the old messages and delete them from discord
	for cID, mID := range oldMessages {
		for _, m := range mID {
			err := s.ChannelMessageDelete(cID, m)
			if err != nil {
				logger.ErrorLog.Printf("Error deleting old message - ID: %s, err: %v", m, err)
				DeleteMessageID(m)
			}
		}
	}

	// send our new message
	st, err = s.ChannelMessageSendComplex(channelID, msg)
	if err != nil {
		logger.ErrorLog.Println("Error sending message:", err)
	}

	// insert the new message id into the database
	err = InsertMessageID(st.ChannelID, st.ID, command)
	if err != nil {
		logger.ErrorLog.Println("Error inserting message id:", err)
	}

	DeleteOldCommandMessages(st.ID, command)
	return st
}

func StopButtonRoutine(command string, s *discordgo.Session) {

	oldMessages, err := GetAllCommandMessages(command)
	if err != nil {
		logger.ErrorLog.Println("Error getting all command messages:", err)
	}

	// iterate over all the old messages and delete them from discord
	for cID, mID := range oldMessages {
		for _, m := range mID {
			err := s.ChannelMessageDelete(cID, m)
			if err != nil {
				logger.ErrorLog.Printf("Error deleting old message - ID: %s - err: %v", m, err)
			}
		}
	}

	DeleteAllCommandMessages(command)
}
