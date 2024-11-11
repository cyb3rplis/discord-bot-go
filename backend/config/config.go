package config

import (
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/dlog"
)

type Config struct {
	Token        string `json:"token"`
	DataDir      string `json:"data_dir"`
	SoundsDir    string `json:"sounds_dir"`
	DB           string `json:"db"`
	AudioTemp    string `json:"audio_temp"`
	AudioTimeout int    `json:"audio_timeout"`
	AdminRole    string `json:"admin"`
}

type ExtendedUser struct {
	User      *discordgo.User `json:"user"`
	Gulagged  sql.NullTime    `json:"gulagged"`
	Remaining time.Duration   `json:"remaining"`
}

type Sound struct {
	ID         int          `json:"id"`
	Name       string       `json:"name"`
	CreatedAt  sql.NullTime `json:"created_at"`
	CategoryID int          `json:"category_id"`
	Hash       string       `json:"hash"`
}

var (
	configInstance *Config
	configOnce     sync.Once
)

var (
	guildInstance *discordgo.Guild
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

	fmt.Printf("+%s+\n", strings.Repeat("-", 40))
	fmt.Printf("| %-15s | %-20s |\n", "TOKEN", configInstance.Token[0:10]+"...")
	fmt.Printf("|%s|\n", strings.Repeat("-", 40))
	fmt.Printf("| %-15s | %-20s |\n", "SOUNDS_DIR", configInstance.SoundsDir)
	fmt.Printf("|%s|\n", strings.Repeat("-", 40))
	fmt.Printf("| %-15s | %-20s |\n", "DB", configInstance.DB)
	fmt.Printf("|%s|\n", strings.Repeat("-", 40))
	fmt.Printf("| %-15s | %-20s |\n", "AUDIO_TEMP", configInstance.AudioTemp)
	fmt.Printf("|%s|\n", strings.Repeat("-", 40))
	fmt.Printf("| %-15s | %-20d |\n", "AUDIO_TIMEOUT", configInstance.AudioTimeout)
	fmt.Printf("|%s|\n", strings.Repeat("-", 40))
	fmt.Printf("| %-15s | %-20s |\n", "ADMIN_ROLE", configInstance.AdminRole)
	fmt.Printf("+%s+\n", strings.Repeat("-", 40))

	return configInstance
}

// GetConfig provides global access to the configuration instance.
func GetConfig() *Config {
	if configInstance == nil {
		return LoadConfig()
	}
	return configInstance
}

func LoadGuild(guild *discordgo.Guild) *discordgo.Guild {
	guildOnce.Do(func() {
		guildInstance = guild
	})

	return guildInstance
}

func GetGuild() *discordgo.Guild {
	return guildInstance
}
