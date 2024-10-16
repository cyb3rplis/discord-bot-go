package utils

import (
	"github.com/cyb3rplis/discord-bot-go/model"
	"os"
	"path/filepath"
	"sort"
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

func SortMapByValue(m map[string]int) map[string]int {
	var keys []string
	var sortedM = make(map[string]int)
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return m[keys[i]] > m[keys[j]]
	})
	for _, k := range keys {
		sortedM[k] = m[k]
	}
	return sortedM
}

func SortMapKeysByValue(m map[string]int) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		return m[keys[i]] > m[keys[j]]
	})
	return keys
}
