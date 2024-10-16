package model

import (
	"database/sql"
	"github.com/cyb3rplis/discord-bot-go/config"
)

var Bot *Model

type Model struct {
	Db     *sql.DB
	Config *config.Config
}

func New(m *Model) *Model {
	cfg := config.GetConfig()
	return &Model{Db: m.Db, Config: cfg}
}
