package model

import (
	"database/sql"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/config"
)

var (
	Meta   *Info
	metaMu sync.Mutex
)

type Info struct {
	Guild       *discordgo.Guild
	BotActivity *time.Time
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
	botActivity := time.Now()
	Meta = &Info{Guild: guild, BotActivity: &botActivity}
	return Meta
}

func UpdateBotActivity() {
	metaMu.Lock()
	defer metaMu.Unlock()
	if Meta != nil && Meta.BotActivity != nil {
		currentTime := time.Now()
		Meta.BotActivity = &currentTime
	}
}
