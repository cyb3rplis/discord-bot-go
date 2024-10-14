package util

import (
	"errors"
	"github.com/cyb3rplis/discord-bot-go/config"
	"os"
	"path/filepath"
	"strings"
)

var cfg = config.GetConfig()

// WalkSoundFiles returns a list of sound files in a subfolder.
func WalkSoundFiles(subfolder string) ([]string, error) {
	baseDir := cfg.SoundsDir
	subfolderPath := filepath.Join(baseDir, subfolder)
	cleanedSubfolderPath := filepath.Clean(subfolderPath)
	// Ensure the cleaned subfolder path is within the base directory
	if !strings.HasPrefix(cleanedSubfolderPath, baseDir) {
		return nil, errors.New("potential path traversal detected")
	}
	files, err := os.ReadDir(cleanedSubfolderPath)
	if err != nil {
		return nil, err
	}
	var soundFiles []string
	for _, file := range files {
		if filepath.Ext(file.Name()) == ".dca" {
			soundFiles = append(soundFiles, file.Name())
		}
	}
	return soundFiles, nil
}

// WalkSoundFolder returns a list of subfolders in the sound folder.
func WalkSoundFolder() ([]string, error) {
	cleanedSubfolderPath := filepath.Clean(cfg.SoundsDir)
	folders, err := os.ReadDir(cleanedSubfolderPath)
	if err != nil {
		return nil, err
	}

	var subfolders []string
	for _, entry := range folders {
		if entry.IsDir() {
			subfolders = append(subfolders, entry.Name())
		}
	}
	return subfolders, nil
}
