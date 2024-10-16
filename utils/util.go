package utils

import (
	"github.com/cyb3rplis/discord-bot-go/model"
	"os"
	"path/filepath"
	"strings"
)

// ScanDirectory scans the sound directory and returns a map of folders and files.
func ScanDirectory() (map[string][]string, error) {
	soundsRoot := model.Bot.Config.SoundsDir
	folderMap := make(map[string][]string)

	err := filepath.WalkDir(soundsRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			// Skip the root folder
			if path == soundsRoot {
				return nil
			}

			// Get relative folder name (e.g., 'folder1/')
			relativeFolder, err := filepath.Rel(soundsRoot, path)
			if err != nil {
				return err
			}

			folderMap[relativeFolder] = []string{} // Initialize an entry for this folder
		} else {
			// Add file to the folder list
			folder := filepath.Dir(path)
			relativeFolder, err := filepath.Rel(soundsRoot, folder)
			if err != nil {
				return err
			}

			// Filter for audio files based on extensions, e.g., ".dca", etc.
			if ext := filepath.Ext(path); ext == ".dca" {
				fileNameWithoutExt := RemoveFileExtension(filepath.Base(path))
				folderMap[relativeFolder] = append(folderMap[relativeFolder], fileNameWithoutExt)
			}
		}
		return nil
	})

	return folderMap, err
}

// RemoveFileExtension removes the file extension from a given file name.
func RemoveFileExtension(fileName string) string {
	return strings.TrimSuffix(fileName, filepath.Ext(fileName))
}
