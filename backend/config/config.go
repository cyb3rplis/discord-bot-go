package config

import (
	"database/sql"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/cyb3rplis/discord-bot-go/logger"
)

type Config struct {
	Token      string   `json:"token"`
	Prefix     string   `json:"prefix"`
	SoundsDir  string   `json:"sounds_dir"`
	DB         string   `json:"db"`
	YTDLP      string   `json:"ytdlp"`
	TTS        string   `json:"tts"`
	TTSTemp    string   `json:"tts_temp"`
	TTSOutput  string   `json:"tts_output"`
	YTOutput   string   `json:"yt_output"`
	YTTemp     string   `json:"yt_temp"`
	YTTimeout  int      `json:"yt_timeout"`
	AdminUsers []string `json:"admin_users"`
}

type User struct {
	ID        string        `json:"id"`
	Username  string        `json:"username"`
	Gulagged  sql.NullTime  `json:"gulagged"`
	Remaining time.Duration `json:"remaining"`
}

var (
	configInstance *Config
	once           sync.Once
)

func LoadConfig() *Config {
	once.Do(func() {
		// Hardcode all values except the Token
		configInstance = &Config{
			Token:      os.Getenv("DISCORD_BOT_TOKEN"), // Read the token from .env
			Prefix:     ".",
			SoundsDir:  "./dist//sounds",
			DB:         "./dist/soundbot.db",
			YTDLP:      "/usr/local/bin/yt-dlp",
			TTS:        "./piper",
			TTSTemp:    "./dist/tts.wav",
			TTSOutput:  "./dist/tts.mp3",
			YTOutput:   "./dist/yt.dca",
			YTTemp:     "./dist/yt.mp3",
			YTTimeout:  20,
			AdminUsers: []string{"378670654146478081", "481894532082958346"},
		}

		// Check if Token is actually set
		if configInstance.Token == "" {
			logger.FatalLog.Fatalf("environment variable DISCORD_BOT_TOKEN not set")
		}
	})

	fmt.Println("---------------------------------")
	fmt.Println(" > TOKEN:\t", configInstance.Token[0:10]+"...")
	fmt.Println(" > PREFIX:\t", configInstance.Prefix)
	fmt.Println(" > SOUNDS_DIR:\t", configInstance.SoundsDir)
	fmt.Println("---------------------------------")

	return configInstance
}

// GetConfig provides global access to the configuration instance.
func GetConfig() *Config {
	if configInstance == nil {
		return LoadConfig()
	}
	return configInstance
}
