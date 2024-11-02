package model

import (
	"fmt"
	"strconv"
)

func GulagUser(userID string, minutes int) error {
	timeout := "+" + strconv.Itoa(minutes) + " minutes"
	res, err := Bot.Db.Exec("UPDATE users SET gulagged = DATETIME(CURRENT_TIMESTAMP, ?) WHERE username = ?;", timeout, userID)
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

func ReleaseUser(userID string) error {
	_, err := Bot.Db.Exec("UPDATE users SET gulagged = NULL WHERE username = ?;", userID)
	if err != nil {
		return err
	}

	return nil
}
