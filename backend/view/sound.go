package view

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cyb3rplis/discord-bot-go/dlog"
	"github.com/cyb3rplis/discord-bot-go/model"

	"github.com/bwmarrin/discordgo"
)

var botSpeaking = false
var stopChannel = make(chan struct{})

var userLastInteraction = make(map[string]time.Time)
var userInteractionCount = make(map[string]int)
var mu sync.Mutex

const maxInteractions = 15             // Maximum allowed interactions before timeout
const resetDuration = 15 * time.Second // Duration to reset the interaction count

type Entry struct {
	ID   int
	Name string
}

func (a *API) PromptInteractionPlaySound(s *discordgo.Session, i *discordgo.InteractionCreate) {
	//get userID
	if i.Member == nil {
		dlog.ErrorLog.Println("error getting member from interaction")
		return
	}

	if i.Type == discordgo.InteractionApplicationCommand {
		switch i.ApplicationCommandData().Name {
		case "play":
			_ = a.SendInteractionRespond("➡ Playing sound", s, i)
			// Find the channel that the interaction came from
			c, err := s.State.Channel(i.ChannelID)
			if err != nil {
				dlog.ErrorLog.Println("error finding channel:", err)
				return
			}

			// Find the guild for that channel
			g, err := s.State.Guild(c.GuildID)
			if err != nil {
				dlog.ErrorLog.Println("error finding guild:", err)
				return
			}

			// Look for the interaction user in that guild's current voice states
			for _, vs := range g.VoiceStates {
				if vs.UserID == i.Member.User.ID {
					soundName := i.ApplicationCommandData().Options[0].StringValue()
					err := a.UpdateInteractionResponse("➡ Playing sound", s, i)
					if err != nil {
						log.Printf("error executing play command: %v", err)
					}

					// Check if the user is in the Gulag
					user, err := a.model.GetUserFromUsername(i.Member.User.GlobalName)
					if err != nil {
						dlog.ErrorLog.Println("error getting user from username:", err)
					} else {
						if remaining, ok := IsUserInGulag(user); ok {
							user.Remaining = remaining
							message := fmt.Sprintf("<@"+user.ID+"> you are in the Gulag for another %s", user.Remaining)
							_, err = a.SendMessage(message, s, i, true)
							if err != nil {
								dlog.ErrorLog.Printf("error sending message: %v", err)
							}
							return
						}
					}
					// Check if the user is in a voice channel
					err = a.VoiceChannelCheck(s, i)
					if err != nil {
						dlog.ErrorLog.Println("error checking voice channel:", err)
						return
					}

					content := []discordgo.MessageComponent{}
					row := discordgo.ActionsRow{}
					row.Components = append(row.Components, discordgo.Button{
						Label:    "Stop Sound",
						Style:    discordgo.DangerButton,
						CustomID: "stop_sound",
					})
					content = append(content, row)

					msg := &discordgo.MessageSend{
						Content:    "➡ Currently Playing by <@" + i.Member.User.Username + ">: " + soundName,
						Components: content,
					}

					// Send the message (+stop button)
					_, err = a.SendMessageComplex(msg, s, i, false)
					if err != nil {
						dlog.ErrorLog.Println("error sending message:", err)
						return
					}

					// Play the custom sound
					err = a.PlaySound(s, i, g.ID, vs.ChannelID, soundName)
					if err != nil {
						dlog.ErrorLog.Println("error playing sound:", err)
					}
				}

			}

		}
	}
}

// PlaySound plays the current buffer to the provided channel.
func (a *API) PlaySound(s *discordgo.Session, i *discordgo.InteractionCreate, guildID, channelID, soundName string) error {
	// check if the bot is currently speaking, and exit early to avoid corrupted sound buffer
	if botSpeaking {
		stopChannel <- struct{}{}
		time.Sleep(150 * time.Millisecond) // Give some time for the current sound to stop
	}

	// Load the sound file.
	var err error
	if soundName == a.model.Config.AudioTemp {
		err = a.model.LoadSoundFS(filepath.Join(a.model.Config.DataDir, soundName+".dca")) //play file from system if function (audio) is played
	} else {
		err = a.model.LoadSound(soundName)
	}

	if err != nil {
		dlog.ErrorLog.Printf("error loading sound %s, %v ", soundName, err)
		_, err := a.SendMessage("Sound does not exist\n> Sikerim", s, i, true)
		if err != nil {
			dlog.ErrorLog.Println("error sending message:", err)
		}
		return err
	}

	// Join the provided voice channel.
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return err
	}
	botSpeaking = true

	// Sleep for a specified amount of time before playing the sound
	time.Sleep(150 * time.Millisecond)

	// Start speaking.
	_ = vc.Speaking(true)

	// Send the buffer data.
	for _, buff := range model.Buffer {
		select {
		case <-stopChannel:
			// Stop sending buffer data if stop signal is received
			_ = vc.Speaking(false)
			botSpeaking = false
			model.Buffer = make([][]byte, 0)
			return nil
		default:
			vc.OpusSend <- buff
		}
	}

	// Stop speaking
	_ = vc.Speaking(false)

	// Sleep for a specified amount of time before ending.
	time.Sleep(250 * time.Millisecond)

	// Disconnect from the provided voice channel.
	//vc.Disconnect()

	// empty buffer to not play older sounds
	model.Buffer = make([][]byte, 0)
	botSpeaking = false

	return nil
}

func (a *API) PlayAudio(audioName string, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	// Find the channel that the interaction came from
	c, err := s.State.Channel(i.ChannelID)
	if err != nil {
		dlog.ErrorLog.Println("error finding channel:", err)
		return err
	}

	// Find the guild for that channel
	g, err := s.State.Guild(c.GuildID)
	if err != nil {
		dlog.ErrorLog.Println("error finding guild:", err)
		return err
	}

	// Look for the interaction user in that guild's current voice states
	for _, vs := range g.VoiceStates {
		if vs.UserID == i.Member.User.ID {
			userID, err := strconv.Atoi(i.Member.User.ID)
			if err != nil {
				dlog.ErrorLog.Println("error converting user ID to int:", err)
				return err
			} else {
				err = a.model.AddUser(userID, i.Member.User.GlobalName)
				if err != nil {
					dlog.ErrorLog.Println("error adding user:", err)
					return err
				}
			}

			content := []discordgo.MessageComponent{}
			row := discordgo.ActionsRow{}
			row.Components = append(row.Components, discordgo.Button{
				Label:    "Stop Sound",
				Style:    discordgo.DangerButton,
				CustomID: "stop_sound",
			})
			content = append(content, row)

			msg := &discordgo.MessageSend{
				Content:    "➡ Currently playing audio by <@" + i.Member.User.ID + "> ",
				Components: content,
			}

			// Send the message (+stop button)
			_, err = a.SendMessageComplex(msg, s, i, false)
			if err != nil {
				dlog.ErrorLog.Println("error sending message:", err)
				return err
			}

			err = a.UpdateInteractionResponse("🎶  Playing audio", s, i)
			if err != nil {
				dlog.ErrorLog.Println("error updating interaction response:", err)
				return err
			}

			dlog.InfoLog.Printf("User: %s played sound", i.Member.User.GlobalName)
			// Play the sound
			err = a.PlaySound(s, i, g.ID, vs.ChannelID, audioName)
			if err != nil {
				dlog.ErrorLog.Println("error playing sound:", err)
				return err
			}

			return nil
		}
	}

	// If the user is not in a voice channel, send an error message
	dlog.InfoLog.Printf("User %s tried to play audio but is not in a voice channel", i.Member.User.GlobalName)
	msg := "You need to be in a voice channel to play audio <@" + i.Member.User.ID + ">"

	_, err = a.SendMessage(msg, s, i, false)
	if err != nil {
		dlog.ErrorLog.Println("error sending message:", err)
		return fmt.Errorf("user not in voice channel")
	}
	return nil
}

// SyncDatabaseWithFileSystem will sync the database with the filesystem.
func (a *API) SyncDatabaseWithFileSystem(folderMap map[string][]string) error {
	existingCategories, _ := a.model.GetCategoriesM() // Get current folders/categories in DB
	existingSounds := a.model.GetSoundsM()            // Get current files/sounds in DB

	for folder, files := range folderMap {
		var categoryID int
		// Check if the folder (category) exists in the database
		if dbCategoryID, exists := existingCategories[folder]; exists {
			categoryID = dbCategoryID // The folder already exists
		} else {
			// The folder doesn't exist in the database, so we need to add it
			if err := a.model.AddCategory(folder); err != nil {
				if !strings.Contains(err.Error(), "UNIQUE constraint failed") {
					dlog.WarningLog.Printf("Failed to add category %s: %v", folder, err)
				}
				continue
			}
			// Fetch the new category ID after insertion
			categoryID = a.model.GetCategoryByID(folder)
		}

		// Add new sounds for this category
		for _, file := range files {
			soundPath := filepath.Join(a.model.Config.SoundsDir, folder, file+".mp3")
			fileData, err := os.ReadFile(soundPath)
			if err != nil {
				return fmt.Errorf("error reading dca sound file: %w", err)
			}

			fileHash, err := model.ComputeFileHash(soundPath)
			if err != nil {
				dlog.WarningLog.Printf("Failed to compute hash for file %s: %v", file, err)
				continue
			}

			if model.FileExistsInDB(existingSounds, categoryID, file, fileHash) {
				// File exists and hasn't changed, skip
				continue
			}

			//convert mp3 to dca
			err = a.ConvertMP3ToDCA(file, folder)
			if err != nil {
				dlog.ErrorLog.Println("error converting mp3 to dca:", err)
				continue
			}
			time.Sleep(3 * time.Second)

			// File does not exist in the DB, add it
			if err := a.model.AddSound(categoryID, file, fileHash, fileData); err != nil {
				//ignore this error if the sound already exists
				if !strings.Contains(err.Error(), "UNIQUE constraint failed") {
					dlog.WarningLog.Printf("Failed to add sound %s to category %s: %v", file, folder, err)
				}
			}
		}
	}

	// Remove entries that no longer exist on the filesystem
	for folder, categoryID := range existingCategories {
		if _, exists := folderMap[folder]; !exists {
			// Folder exists in the database but not in the filesystem
			if err := a.model.RemoveCategory(categoryID); err != nil {
				dlog.InfoLog.Printf("Failed to remove category %s (ID: %d): %v", folder, categoryID, err)
			}
			dlog.InfoLog.Printf("Removed category %s (ID: %d)", folder, categoryID)
		} else {
			// For existing folders, remove missing files
			dbFiles := existingSounds[categoryID] // Files in the DB
			fsFiles := folderMap[folder]          // Files in the filesystem

			for dbFile := range dbFiles {
				if !model.FileExistsInFS(fsFiles, dbFile) {
					// File exists in the database but not in the filesystem
					if err := a.model.RemoveSound(categoryID, dbFile); err != nil {
						dlog.InfoLog.Printf("Failed to remove sound %s from category %s: %v", dbFile, folder, err)
					}

					dlog.InfoLog.Printf("Removed sound %s from category %s", dbFile, folder)
				}
			}
		}
	}

	return nil
}

func (a *API) VoiceChannelCheck(s *discordgo.Session, i *discordgo.InteractionCreate) error {
	userInVS := false
	// Find the guild for that channel
	g, err := s.State.Guild(i.GuildID)
	if err != nil {
		dlog.ErrorLog.Println("error finding guild:", err)
		return err
	}
	for _, vs := range g.VoiceStates {
		if vs.UserID == i.Member.User.ID {
			userInVS = true
		}
	}
	if !userInVS {
		// If the user is not in a voice channel, send an error message and avoid processing the audio
		dlog.InfoLog.Printf("User %s tried to play sound but is not in a voice channel", i.Member.User.GlobalName)
		msg := "You need to be in a voice channel to play sounds <@" + i.Member.User.ID + ">"
		_, err = a.SendMessage(msg, s, i, false)
		if err != nil {
			dlog.ErrorLog.Println("error sending message:", err)
		}
		return fmt.Errorf("user not in voice channel, quitting early to avoid delay")
	}

	return nil
}
