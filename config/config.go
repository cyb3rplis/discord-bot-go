package config

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/cyb3rplis/discord-bot-go/logger"
)

type Config struct {
	Token      string   `json:"token"`
	Prefix     string   `json:"prefix"`
	SoundsDir  string   `json:"sounds_dir"`
	DB         string   `json:"db"`
	TTS        string   `json:"tts"`
	TTSInput   string   `json:"tts_input"`
	TTSOutput  string   `json:"tts_output"`
	YTOutput   string   `json:"yt_output"`
	YTTemp     string   `json:"yt_temp"`
	YTTimeout  int      `json:"yt_timeout"`
	AdminUsers []string `json:"admin_users"`
}

var (
	configInstance *Config
	once           sync.Once
)

func LoadConfig() *Config {
	confVersion := "local"
	once.Do(func() {
		configFile := "./config/config.local.json"
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			configFile = "./config/config.json" // Fallback to the default config file
			confVersion = "production"
		}

		configInstance = &Config{}
		file, err := os.ReadFile(configFile)
		if err != nil {
			logger.FatalLog.Fatalf("error reading config file: %v", err)
		}

		err = json.Unmarshal(file, configInstance)
		if err != nil {
			logger.FatalLog.Fatalf("error parsing config file: %v", err)
			os.Exit(1)
		}
	})

	fmt.Println("---------------------------------")
	fmt.Println(" > CONFIG:\t", confVersion)
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
