package model

import (
	"database/sql"

	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/config"
	"github.com/cyb3rplis/discord-bot-go/dlog"
)

// GetUsers returns all users from the database
func (m *Model) GetUsers() (users []config.ExtendedUser, err error) {
	rows, err := m.Db.Query("SELECT id, username, gulagged FROM users;")
	if err != nil {
		dlog.FatalLog.Fatal(err)
	}
	defer rows.Close()

	for rows.Next() {
		var u config.ExtendedUser
		u.User = &discordgo.User{}
		err = rows.Scan(&u.User.ID, &u.User.GlobalName, &u.Gulagged)
		if err != nil {
			dlog.FatalLog.Fatal(err)
		}

		users = append(users, u)
	}

	return users, err
}

// AddUser adds a user to the database
func (m *Model) AddUser(user *discordgo.User) error {
	_, err := m.Db.Exec("INSERT INTO users (id, username) VALUES (?, ?) ON CONFLICT(id) DO UPDATE SET username = excluded.username;", user.ID, user.GlobalName)
	if err != nil {
		return err
	}

	return nil
}

// SetUserGulaggedValue sets the gulagged value of a user
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

// FetchAndStoreGuildMembers fetches all members of the guild and stores them in the database
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
		members, err := s.GuildMembers(guildID, after, 25)
		if err != nil {
			dlog.FatalLog.Printf("Failed to fetch members: %v", err)
		}

		// Exit the loop if no more members are returned
		if len(members) == 0 {
			break
		}

		// Insert members into the database
		for _, member := range members {
			// add only non-bot users with a global name, they also have to be NOT offline
			presence, err := s.State.Presence(guildID, member.User.ID)
			if err != nil {
				continue
			}

			if !member.User.Bot && member.User.GlobalName != "" && presence.Status != discordgo.StatusOffline {
				{
					err = m.AddUser(member.User)
					if err != nil {
						dlog.ErrorLog.Printf("Failed to insert member %s: %v", member.User.ID, err)
					}
				}
			}
		}

		// Set the 'after' parameter to the last member's ID for pagination
		after = members[len(members)-1].User.ID
	}
}
