package model

import (
	"database/sql"

	"github.com/cyb3rplis/discord-bot-go/config"
)

var Bot *Model
var Meta *Info

type Model struct {
	Db     *sql.DB
	Config *config.Config
}

type Info struct {
	Guild *config.Guild
}

func NewBot(m *Model) *Model {
	cfg := config.GetConfig()
	return &Model{Db: m.Db, Config: cfg}
}

func NewInfo() *Info {
	guild := config.GetGuild()
	return &Info{Guild: guild}
}
