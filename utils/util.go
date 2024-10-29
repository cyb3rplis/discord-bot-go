package utils

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/config"
	"github.com/cyb3rplis/discord-bot-go/logger"
	"github.com/cyb3rplis/discord-bot-go/model"
)

type SoundInfo struct {
	Name     string `json:"alias"`
	Count    int    `json:"total_plays"`
	Category string `json:"category_name"`
}

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

// AddUserStatistics adds a sound play to the user statistics
func AddUserStatistics(userID int, soundName string) error {
	_, err := model.Bot.Db.Exec("INSERT INTO stats_users (user_id, sound_id, count) VALUES (?, (SELECT id FROM sounds WHERE name = ?), 1) ON CONFLICT(user_id, sound_id) DO UPDATE SET count = count + 1;", userID, soundName)
	if err != nil {
		return err
	}

	return nil
}

// GetSoundStatistics returns the top sounds played
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

// GetUserStatistics returns the top sounds played by a user
func GetUserStatistics(userID string, limit int) (soundStats []SoundInfo, err error) {
	// this can be used to create buttons when the user gets their stats
	rows, err := model.Bot.Db.Query(`
	SELECT s.alias, COALESCE(SUM(su.count), 0) AS total_plays, c.name
	FROM sounds AS s
	LEFT JOIN stats_users AS su ON s.id = su.sound_id AND su.user_id = ?
	JOIN categories AS c ON s.category_id = c.id
	GROUP BY s.id, s.alias
	HAVING total_plays > 0
	ORDER BY total_plays
	DESC LIMIT ?;`, userID, limit)

	if err != nil {
		logger.FatalLog.Fatal(err)
	}
	defer rows.Close()

	soundStats = []SoundInfo{}
	for rows.Next() {
		var sound sql.NullString
		var count sql.NullInt64
		var category sql.NullString

		var stat SoundInfo

		err = rows.Scan(&sound, &count, &category)
		if err != nil {
			logger.FatalLog.Fatal(err)
		}
		if sound.Valid && count.Valid && category.Valid {
			stat.Name = sound.String
			stat.Count = int(count.Int64)
			stat.Category = category.String
		}

		soundStats = append(soundStats, stat)
	}

	return soundStats, err
}

func GetAllUserStatistics() (soundStats map[string]int, err error) {
	rows, err := model.Bot.Db.Query(`
	SELECT u.username, COALESCE(SUM(su.count), 0) AS total_plays
	FROM stats_users AS su
	LEFT JOIN users AS u ON su.user_id = u.id
	GROUP BY u.id
	HAVING total_plays > 0
	ORDER BY total_plays
	DESC LIMIT 10;`)

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

func NewMessageRoutine(command, message string, s *discordgo.Session, m *discordgo.MessageCreate) (st *discordgo.Message) {
	// send our new message
	st, err := s.ChannelMessageSend(m.ChannelID, message)
	if err != nil {
		logger.ErrorLog.Println("error sending message:", err)
	}

	// get all old messages for this command
	oldMessages, err := GetAllCommandMessages(command)
	if err != nil {
		logger.ErrorLog.Println("error getting all command messages:", err)
	}

	// iterate over all the old messages and delete them from discord
	for cID, mID := range oldMessages {
		for _, msg := range mID {
			err := s.ChannelMessageDelete(cID, msg)
			if err != nil {
				logger.ErrorLog.Printf("error deleting old message - ID: %s, err: %v", msg, err)
				err = DeleteMessageID(msg)
				if err != nil {
					logger.ErrorLog.Printf("error deleting message from DB: %v", err)
				}
			}
		}
	}

	// insert the new message id into the database
	err = InsertMessageID(st.ChannelID, st.ID, command)
	if err != nil {
		logger.ErrorLog.Println("error inserting message id:", err)
	}

	err = DeleteOldCommandMessages(st.ID, command)
	if err != nil {
		logger.ErrorLog.Println("error deleting old message:", err)
	}

	return st
}

func NewPrivateMessageRoutine(message string, s *discordgo.Session, m *discordgo.MessageCreate) (err error) {
	// Send a DM to the author
	channel, err := s.UserChannelCreate(m.Author.ID)
	if err != nil {
		return fmt.Errorf("error creating private channel with user: %s", m.Author.ID)
	}
	_, err = s.ChannelMessageSend(channel.ID, message)
	if err != nil {
		return fmt.Errorf("error sending private message: %v", err)
	}

	return nil
}

func NewComplexMessageRoutine(command, channelID, msgID string, msg *discordgo.MessageSend, s *discordgo.Session) (st *discordgo.Message) {
	// send our new message
	st, err := s.ChannelMessageSendComplex(channelID, msg)
	if err != nil {
		logger.ErrorLog.Println("error sending message:", err)
	}

	// get all old messages for this command
	oldMessages, err := GetAllCommandMessages(command)
	if err != nil {
		logger.ErrorLog.Println("error getting all command messages:", err)
	}

	// iterate over all the old messages and delete them from discord
	for cID, mID := range oldMessages {
		for _, msg := range mID {
			err := s.ChannelMessageDelete(cID, msg)
			if err != nil {
				logger.ErrorLog.Printf("error deleting old message - ID: %s, err: %v", msg, err)
				err = DeleteMessageID(msg)
				if err != nil {
					logger.ErrorLog.Printf("error deleting message from DB: %v", err)
				}
			}
		}
	}

	// insert the new message id into the database
	err = InsertMessageID(st.ChannelID, st.ID, command)
	if err != nil {
		logger.ErrorLog.Println("error inserting message id:", err)
	}

	err = DeleteOldCommandMessages(st.ID, command)
	if err != nil {
		logger.ErrorLog.Println("error deleting old message:", err)
	}
	return st
}

func DeleteMessageRoutine(s *discordgo.Session, command string) {
	oldMessages, err := GetAllCommandMessages(command)
	if err != nil {
		logger.ErrorLog.Println("error getting all command messages:", err)
	}

	// iterate over all the old messages and delete them from discord
	for cID, mID := range oldMessages {
		for _, msg := range mID {
			err := s.ChannelMessageDelete(cID, msg)
			if err != nil {
				logger.ErrorLog.Printf("error deleting old message - ID: %s - err: %v", msg, err)
			}
		}
	}

	err = DeleteAllCommandMessages(command)
	if err != nil {
		logger.ErrorLog.Println("error deleting all command messages:", err)
	}
}

func CleanUpSoundFile(module string) {
	if module == "tts" {
		err := os.Remove(model.Bot.Config.TTSTemp)
		if err != nil {
			// Handle error if file deletion fails
			logger.ErrorLog.Printf("error deleting file: %v\n", err)
		}

		err = os.Remove(model.Bot.Config.TTSOutput)
		if err != nil {
			// Handle error if file deletion fails
			logger.ErrorLog.Printf("error deleting file: %v\n", err)
		}

		logger.InfoLog.Println("Deleted temp TTS sound files successfully")
	}
}

func VoiceChannelCheck(s *discordgo.Session, m *discordgo.MessageCreate) error {
	userInVS := false
	c, err := s.State.Channel(m.ChannelID)
	if err != nil {
		logger.ErrorLog.Println("error finding channel:", err)
		return err
	}

	// Find the guild for that channel
	g, err := s.State.Guild(c.GuildID)
	if err != nil {
		logger.ErrorLog.Println("error finding guild:", err)
		return err
	}

	for _, vs := range g.VoiceStates {
		if vs.UserID == m.Author.ID {
			userInVS = true
		}
	}

	if !userInVS {
		// If the user is not in a voice channel, send an error message and avoid processing the youtube audio
		logger.InfoLog.Printf("User %s tried to play youtube sound but is not in a voice channel", m.Author.GlobalName)
		message := "You need to be in a voice channel to play sounds <@" + m.Author.ID + ">"

		NewMessageRoutine(".novc"+m.Author.ID, message, s, m)
		return fmt.Errorf("user not in voice channel, quitting early to avoid delay")
	}

	return nil
}

// BuildSoundButtons creates a list of buttons for the provided category
func BuildSoundButtons(sounds []string, category string, buttonStyle discordgo.ButtonStyle) []discordgo.MessageComponent {
	content := []discordgo.MessageComponent{}
	row := discordgo.ActionsRow{}
	for i, soundName := range sounds {
		soundName = strings.TrimSuffix(soundName, ".dca")
		// only 5 buttons per row - discord does not allow more
		if i > 0 && i%5 == 0 {
			content = append(content, row)
			row = discordgo.ActionsRow{}
		}
		row.Components = append(row.Components, discordgo.Button{
			Label:    soundName,
			Style:    buttonStyle,
			CustomID: fmt.Sprintf("play_sound_%s_%s", category, soundName),
		})
	}
	// Append the last row if it has any components
	if len(row.Components) > 0 {
		content = append(content, row)
	}
	return content
}

// BuildListButtons creates a list of buttons for the provided categories
func BuildListButtons(categories []string, buttonStyle discordgo.ButtonStyle) []discordgo.MessageComponent {
	content := []discordgo.MessageComponent{}
	row := discordgo.ActionsRow{}
	for i, category := range categories {
		// only 5 buttons per row - discord does not allow more
		if i > 0 && i%5 == 0 {
			content = append(content, row)
			row = discordgo.ActionsRow{}
		}
		row.Components = append(row.Components, discordgo.Button{
			Label:    category,
			Style:    buttonStyle,
			CustomID: fmt.Sprintf("list_sounds_%s", category),
		})
	}
	// Append the last row if it has any components
	if len(row.Components) > 0 {
		content = append(content, row)
	}
	return content
}

// BuildMessages creates a list of messages for the provided buttons
func BuildMessages(buttons []discordgo.MessageComponent, initialMessage *discordgo.MessageSend) []*discordgo.MessageSend {
	var messages []*discordgo.MessageSend
	if initialMessage != nil {
		messages = append(messages, initialMessage)
	}

	for len(buttons) > 0 {
		var messageContent []discordgo.MessageComponent
		if len(buttons) > 5 {
			messageContent, buttons = buttons[:5], buttons[5:]
		} else {
			messageContent, buttons = buttons, nil
		}
		message := &discordgo.MessageSend{
			Components: messageContent,
		}
		messages = append(messages, message)
	}
	return messages
}

func GetUsers() (users []config.User, err error) {
	rows, err := model.Bot.Db.Query("SELECT id, username, gulagged FROM users;")

	if err != nil {
		logger.FatalLog.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var u config.User

		err = rows.Scan(&u.ID, &u.Username, &u.Gulagged)
		if err != nil {
			logger.FatalLog.Fatal(err)
		}

		users = append(users, u)
	}

	return users, err
}

func GulagUser(userID string) error {
	_, err := model.Bot.Db.Exec("UPDATE users SET gulagged = CURRENT_TIMESTAMP WHERE id = ?;", userID)
	if err != nil {
		return err
	}

	return nil
}

func ReleaseUser(userID string) error {
	_, err := model.Bot.Db.Exec("UPDATE users SET jailed = NULL WHERE id = ?;", userID)
	if err != nil {
		return err
	}

	return nil
}

func IsAdmin(userID string) bool {
	admins := model.Bot.Config.AdminUsers
	for _, a := range admins {
		if userID == a {
			return true
		}
	}

	return false
}
