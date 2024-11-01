package config

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
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

var AppPath = func() string {
	if path := os.Getenv("APP_PATH"); path != "" { //use for local development
		return path
	}
	return "./"
}

func LoadConfig() *Config {
	once.Do(func() {

		configInstance = &Config{
			Token:      os.Getenv("TOKEN"), // Read the token from .env
			Prefix:     ".",
			SoundsDir:  filepath.Join(AppPath(), "data", "sounds"),
			DB:         filepath.Join(AppPath(), "data", "soundbot.db"),
			YTDLP:      "/usr/local/bin/yt-dlp",
			TTS:        filepath.Join(AppPath(), "piper"),
			TTSTemp:    filepath.Join(AppPath(), "data", "tts.wav"),
			TTSOutput:  filepath.Join(AppPath(), "data", "tts.mp3"),
			YTOutput:   filepath.Join(AppPath(), "data", "yt.dca"),
			YTTemp:     filepath.Join(AppPath(), "data", "yt.mp3"),
			YTTimeout:  20,
			AdminUsers: []string{"378670654146478081", "481894532082958346"},
		}
		// Check if Token is actually set
		if configInstance.Token == "" {
			logger.FatalLog.Fatalf("environment variable TOKEN not set")
		}
	})

	fmt.Println("---------------------------------------------------")
	fmt.Println(" > TOKEN:\t", configInstance.Token[0:10]+"...")
	fmt.Println(" > PREFIX:\t", configInstance.Prefix)
	fmt.Println(" > SOUNDS_DIR:\t", configInstance.SoundsDir)
	fmt.Println(" > DB:\t\t", configInstance.DB)
	fmt.Println(" > YTDLP:\t", configInstance.YTDLP)
	fmt.Println(" > TTS:\t\t", configInstance.TTS)
	fmt.Println(" > TTS_TEMP:\t", configInstance.TTSTemp)
	fmt.Println(" > TTS_OUTPUT:\t", configInstance.TTSOutput)
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
