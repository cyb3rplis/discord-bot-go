package model

import (
	"database/sql"
	"github.com/cyb3rplis/discord-bot-go/config"
	"github.com/cyb3rplis/discord-bot-go/logger"
)

func (m *Model) GetUsers() (users []config.User, err error) {
	rows, err := m.Db.Query("SELECT id, username, gulagged FROM users;")

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

func (m *Model) GetUserFromUsername(username string) (user config.User, err error) {
	err = m.Db.QueryRow("SELECT id, username, gulagged FROM users WHERE username = ?;", username).Scan(&user.ID, &user.Username, &user.Gulagged)

	if err != nil {
		if err == sql.ErrNoRows {
			return user, err
		}
		logger.FatalLog.Fatal(err)
	}

	return user, nil
}
