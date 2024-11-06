package config

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/dlog"
)

type Config struct {
	Token     string `json:"token"`
	Prefix    string `json:"prefix"`
	SoundsDir string `json:"sounds_dir"`
	DB        string `json:"db"`
	YTOutput  string `json:"yt_output"`
	YTTemp    string `json:"yt_temp"`
	YTTimeout int    `json:"yt_timeout"`
}

type User struct {
	ID        string        `json:"id"`
	Username  string        `json:"username"`
	Gulagged  sql.NullTime  `json:"gulagged"`
	Remaining time.Duration `json:"remaining"`
}

type Guild struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var (
	configInstance *Config
	configOnce     sync.Once
)

var (
	guildInstance *Guild
	guildOnce     sync.Once
)

var AppPath = func() string {
	if path := os.Getenv("APP_PATH"); path != "" { //use for local development
		return path
	}
	return "./"
}

func LoadConfig() *Config {
	configOnce.Do(func() {

		configInstance = &Config{
			Token:     os.Getenv("TOKEN"), // Read the token from .env
			Prefix:    ".",
			SoundsDir: filepath.Join(AppPath(), "data", "sounds"),
			DB:        filepath.Join(AppPath(), "data", "soundbot.db"),
			YTOutput:  filepath.Join(AppPath(), "data", "yt.dca"),
			YTTemp:    filepath.Join(AppPath(), "data", "yt.mp3"),
			YTTimeout: 20,
		}
		// Check if Token is actually set
		if configInstance.Token == "" {
			dlog.FatalLog.Fatalf("environment variable TOKEN not set")
		}

		// check if necessary binaries are on the system
		binaries := []string{"dca", "ffmpeg", "yt-dlp"}
		for _, bin := range binaries {
			_, err := exec.LookPath(bin)
			if err != nil {
				dlog.FatalLog.Fatalf("%s not in PATH", bin)
			}
		}
	})

	fmt.Println("---------------------------------------------------")
	fmt.Println(" > TOKEN:\t", configInstance.Token[0:10]+"...")
	fmt.Println(" > PREFIX:\t", configInstance.Prefix)
	fmt.Println(" > SOUNDS_DIR:\t", configInstance.SoundsDir)
	fmt.Println(" > DB:\t\t", configInstance.DB)
	fmt.Println(" > YT_OUTPUT:\t", configInstance.YTOutput)
	fmt.Println(" > YT_TEMP:\t", configInstance.YTTemp)
	fmt.Println(" > YTTimeout:\t", configInstance.YTTimeout)
	fmt.Println("---------------------------------------------------")

	return configInstance
}

// GetConfig provides global access to the configuration instance.
func GetConfig() *Config {
	if configInstance == nil {
		return LoadConfig()
	}
	return configInstance
}

func LoadGuild(event *discordgo.GuildCreate) *Guild {
	guildOnce.Do(func() {
		guildInstance = &Guild{
			ID:   event.Guild.ID,
			Name: event.Guild.Name,
		}
	})

	return guildInstance
}

func GetGuild() *Guild {
	return guildInstance
}
