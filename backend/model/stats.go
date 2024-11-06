package model

import (
	"database/sql"
	"github.com/cyb3rplis/discord-bot-go/logger"
)

type SoundInfo struct {
	Name     string `json:"alias"`
	Count    int    `json:"total_plays"`
	Category string `json:"category_name"`
}

func (m *Model) GetAllUserStatistics() (soundStats map[string]int, err error) {
	rows, err := m.Db.Query(`
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

// GetUserStatistics returns the top sounds played by a user
func (m *Model) GetUserStatistics(userID string, limit int) (soundStats []SoundInfo, err error) {
	// this can be used to create buttons when the user gets their stats
	rows, err := m.Db.Query(`
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

// GetSoundStatistics returns the top sounds played
func (m *Model) GetSoundStatistics() (soundStats map[string]int, err error) {
	rows, err := m.Db.Query("SELECT s.alias, COALESCE(SUM(su.count), 0) AS total_plays FROM sounds AS s LEFT JOIN stats_users AS su ON s.id = su.sound_id GROUP BY s.id, s.alias HAVING total_plays > 0 ORDER BY total_plays DESC LIMIT 10;")
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

// AddUserStatistics adds a sound play to the user statistics
func (m *Model) AddUserStatistics(userID int, soundName string) error {
	_, err := m.Db.Exec("INSERT INTO stats_users (user_id, sound_id, count) VALUES (?, (SELECT id FROM sounds WHERE name = ?), 1) ON CONFLICT(user_id, sound_id) DO UPDATE SET count = count + 1;", userID, soundName)
	if err != nil {
		return err
	}

	return nil
}
