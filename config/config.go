package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
)

type Config struct {
	Token     string `json:"token"`
	Prefix    string `json:"prefix"`
	SoundsDir string `json:"sounds_dir"`
	DB        string `json:"db"`
	Schema    string `json:"schema"`
}

var (
	configInstance *Config
	once           sync.Once
)

func LoadConfig() *Config {
	confVersion := "local"
	once.Do(func() {
		configFile := "../config/config.local.json"
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			configFile = "../config/config.json" // Fallback to the default config file
			confVersion = "production"
		}

		configInstance = &Config{}
		file, err := os.ReadFile(configFile)
		if err != nil {
			log.Fatalf("Error reading config file: %v", err)
		}

		err = json.Unmarshal(file, configInstance)
		if err != nil {
			log.Fatalf("Error parsing config file: %v", err)
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
