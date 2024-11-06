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
	Token        string `json:"token"`
	Prefix       string `json:"prefix"`
	DataDir      string `json:"data_dir"`
	SoundsDir    string `json:"sounds_dir"`
	DB           string `json:"db"`
	AudioTemp    string `json:"audio_temp"`
	AudioTimeout int    `json:"audio_timeout"`
	AdminRole    string `json:"admin"`
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
			Token:        os.Getenv("TOKEN"), // Read the token from .env
			Prefix:       ".",
			DataDir:      filepath.Join(AppPath(), "data"),
			SoundsDir:    filepath.Join(AppPath(), "data", "sounds"),
			DB:           filepath.Join(AppPath(), "data", "soundbot.db"),
			AudioTemp:    "temp",
			AudioTimeout: 20,
			AdminRole:    os.Getenv("ADMIN_ROLE"),
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

	fmt.Println(" > AUDIO_TEMP:\t", configInstance.AudioTemp)
	fmt.Println(" > AUDIO_TIMEOUT:\t", configInstance.AudioTimeout)
	fmt.Println(" > ADMIN_ROLE:\t", configInstance.AdminRole)

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
