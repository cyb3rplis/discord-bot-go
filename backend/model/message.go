package model

import (
	"database/sql"
	"fmt"
	"github.com/cyb3rplis/discord-bot-go/logger"
)

func (m *Model) GetCategoriesM() (map[string]int, error) {
	rows, err := m.Db.Query("SELECT id, name FROM categories")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := make(map[string]int)
	for rows.Next() {
		var id int
		var name string
		if err := rows.Scan(&id, &name); err != nil {
			return categories, err
		}
		categories[name] = id
	}
	return categories, nil
}

// GetCategories returns a list of sound categories (from DB)
func (m *Model) GetCategories() ([]string, error) {
	rows, err := m.Db.Query("SELECT name FROM categories")
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}
	defer rows.Close()

	var categories []string
	for rows.Next() {
		var category string
		err := rows.Scan(&category)
		if err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, category)
	}
	return categories, nil
}

func (m *Model) GetAllMessages() (messages map[string][]string, err error) {
	rows, err := m.Db.Query("SELECT channel_id, message_id FROM messages;")
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

func (m *Model) GetAllCommandMessages(command string) (messages map[string][]string, err error) {
	rows, err := m.Db.Query("SELECT channel_id, message_id FROM messages WHERE command = ?;", command)
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

func (m *Model) InsertMessageID(channelID, messageID, command string) error {
	_, err := m.Db.Exec("INSERT INTO messages (channel_id, message_id, command) VALUES (?, ?, ?);", channelID, messageID, command)
	if err != nil {
		return err
	}

	return nil
}

func (m *Model) DeleteMessageID(messageID string) error {
	_, err := m.Db.Exec("DELETE FROM messages WHERE message_id = ?;", messageID)
	if err != nil {
		return err
	}

	return nil
}

func (m *Model) DeleteAllCommandMessages(command string) error {
	_, err := m.Db.Exec("DELETE FROM messages WHERE command = ?;", command)
	if err != nil {
		return err
	}

	return nil
}

func (m *Model) DeleteOldCommandMessages(newID, command string) error {
	_, err := m.Db.Exec("DELETE FROM messages WHERE message_id != ? AND command = ?;", newID, command)
	if err != nil {
		return err
	}

	return nil
}

func (m *Model) DeleteAllMessages() error {
	_, err := m.Db.Exec("DELETE FROM messages;")
	if err != nil {
		return err
	}

	return nil
}
