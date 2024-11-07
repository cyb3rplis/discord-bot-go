package model

import (
	"database/sql"

	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/config"
	"github.com/cyb3rplis/discord-bot-go/dlog"
)

func (m *Model) GetUsers() (users []config.ExtendedUser, err error) {
	rows, err := m.Db.Query("SELECT id, username, gulagged FROM users;")
	if err != nil {
		dlog.FatalLog.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var u config.ExtendedUser
		u.User = &discordgo.User{}
		gulagged := sql.NullTime{}

		err = rows.Scan(&u.User.ID, &u.User.GlobalName, &gulagged)
		if err != nil {
			dlog.FatalLog.Fatal(err)
		}

		if gulagged.Valid {
			u.Gulagged = gulagged
		}

		users = append(users, u)
	}

	return users, err
}

func (m *Model) AddUser(user *discordgo.User) error {
	// userID, err := strconv.Atoi(user.ID)
	// if err != nil {
	// 	dlog.ErrorLog.Println("error converting user ID to int:", err)
	// 	return err
	// }

	_, err := m.Db.Exec("INSERT INTO users (id, username) VALUES (?, ?) ON CONFLICT(id) DO UPDATE SET username = excluded.username;", user.ID, user.GlobalName)
	if err != nil {
		return err
	}

	return nil
}

func (m *Model) SetUserGulaggedValue(user *discordgo.User) (config.ExtendedUser, error) {
	extendedUser := config.ExtendedUser{
		User: user,
	}
	err := m.Db.QueryRow("SELECT gulagged FROM users WHERE username = ?;", user.GlobalName).Scan(&extendedUser.Gulagged)

	if err != nil {
		if err == sql.ErrNoRows {
			return extendedUser, err
		}
		return extendedUser, err
	}

	return extendedUser, nil
}

func (m *Model) FetchAndStoreGuildMembers(s *discordgo.Session) {
	if m == nil {
		dlog.ErrorLog.Println("model is nil")
		return
	}
	guildID := Meta.Guild.ID
	if guildID == "" {
		dlog.ErrorLog.Println("guildID is empty")
		return
	}

	after := "" // empty string means starting from the first member
	for {
		// Fetch a batch of up to 1,000 members
		members, err := s.GuildMembers(guildID, after, 1000)
		if err != nil {
			dlog.FatalLog.Printf("Failed to fetch members: %v", err)
		}

		// Exit the loop if no more members are returned
		if len(members) == 0 {
			break
		}

		// Insert members into the database
		for _, member := range members {
			if !member.User.Bot {
				err = m.AddUser(member.User)
				if err != nil {
					dlog.ErrorLog.Printf("Failed to insert member %s: %v", member.User.ID, err)
				}
			}

		}

		// Set the 'after' parameter to the last member's ID for pagination
		after = members[len(members)-1].User.ID
	}
}
