package model

import (
	"fmt"
	"strconv"

	"github.com/cyb3rplis/discord-bot-go/config"
)

func (m *Model) GulagUser(user config.ExtendedUser, minutes int) error {
	timeout := "+" + strconv.Itoa(minutes) + " minutes"
	res, err := m.Db.Exec("UPDATE users SET gulagged = DATETIME(CURRENT_TIMESTAMP, ?) WHERE username = ?;", timeout, user.User.GlobalName)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (m *Model) ReleaseUser(user config.ExtendedUser) error {
	_, err := m.Db.Exec("UPDATE users SET gulagged = NULL WHERE username = ?;", user.User.GlobalName)
	if err != nil {
		return err
	}

	return nil
}
