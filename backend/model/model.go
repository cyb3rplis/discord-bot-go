package model

import (
	"database/sql"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/config"
)

var Meta *Info

type Info struct {
	Guild *discordgo.Guild
}
type Model struct {
	Db     *sql.DB
	Config *config.Config
	Mu     *sync.Mutex
}

// New returns a new Model struct
func New(m *Model) *Model {
	cfg := config.GetConfig()
	return &Model{Db: m.Db, Config: cfg}
}

// NewInfo returns a new Info struct
func NewInfo() *Info {
	guild := config.GetGuild()
	Meta = &Info{Guild: guild}
	return Meta
}
