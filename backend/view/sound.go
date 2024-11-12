package view

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
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
	interactionUser := i.Member.User

	if i.Type == discordgo.InteractionApplicationCommand {
		switch i.ApplicationCommandData().Name {
		case "play":
			_ = a.SendInteractionRespond("➡ Playing sound", s, i)
			// Find the channel that the interaction came from
			guild, err := s.State.Guild(model.Meta.Guild.ID)
			if err != nil {
				dlog.ErrorLog.Println("error finding guild:", err)
				return
			}

			// Check if the user is in a voice channel
			vs, err := a.VoiceChannelCheck(s, i)
			if err != nil {
				dlog.ErrorLog.Println("error checking voice channel:", err)
				return
			}

			// Check if the user is in the Gulag
			user, err := a.model.SetUserGulaggedValue(interactionUser)
			if err != nil && err != sql.ErrNoRows {
				dlog.ErrorLog.Println("error getting user from username:", err)
			} else {
				if user, ok := SetUserGulagRemaining(user); ok {
					message := fmt.Sprintf(user.User.Mention()+" you are in the Gulag for another %s", user.Remaining)
					_, err = a.SendMessage(message, s, i, true)
					if err != nil {
						dlog.ErrorLog.Printf("error sending message: %v", err)
					}
					return
				}
			}

			soundName := i.ApplicationCommandData().Options[0].StringValue()

			content := []discordgo.MessageComponent{}
			row := discordgo.ActionsRow{}
			row.Components = append(row.Components, discordgo.Button{
				Label:    "Stop Sound",
				Style:    discordgo.DangerButton,
				CustomID: "stop_sound",
			})
			content = append(content, row)

			msg := &discordgo.MessageSend{
				Content:    "➡ Currently Playing by " + user.User.Mention() + ": " + soundName,
				Components: content,
			}

			// Send the message (+stop button)
			st, err := a.SendMessageComplex(msg, s, i, false)
			if err != nil {
				dlog.ErrorLog.Println("error sending message:", err)
				return
			}

			err = a.DeleteOldStopSoundButtons(s, st)
			if err != nil {
				dlog.ErrorLog.Println("error deleting stop sound buttons:", err)
			}

			err = a.UpdateInteractionResponse("➡ Playing sound", s, i)
			if err != nil {
				log.Printf("error executing play command: %v", err)
			}

			// Play the custom sound
			err = a.PlaySound(s, i, guild.ID, vs.ChannelID, soundName)
			if err != nil {
				dlog.ErrorLog.Println("error playing sound:", err)
				err = a.UpdateInteractionResponse("➡ Sound not found", s, i)
				if err != nil {
					log.Printf("error executing play command: %v", err)
				}
			}

			time.Sleep(250 * time.Millisecond)
			st, err = s.ChannelMessage(st.ChannelID, st.ID)
			if err == nil {
				err = s.ChannelMessageDelete(st.ChannelID, st.ID)
				if err != nil {
					dlog.ErrorLog.Println("error deleting stop sound button after sound finished:", err)
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
	var buffer [][]byte
	if soundName == a.model.Config.AudioTemp {
		buffer, err = a.model.LoadSoundFS(filepath.Join(a.model.Config.DataDir, soundName+".dca")) //play file from system if function (audio) is played
	} else {
		buffer, err = a.model.LoadSound(soundName)
	}
	if err != nil {
		dlog.ErrorLog.Printf("error loading sound %s, %v ", soundName, err)
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
	for _, buff := range buffer {
		select {
		case <-stopChannel:
			// Stop sending buffer data if stop signal is received
			_ = vc.Speaking(false)
			botSpeaking = false
			buffer = nil
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
	buffer = nil
	botSpeaking = false

	// Update the bot activity
	model.UpdateBotActivity()
	return nil
}

func (a *API) PlayAudio(audioName string, s *discordgo.Session, i *discordgo.InteractionCreate) error {
	guild, err := s.State.Guild(model.Meta.Guild.ID)
	if err != nil {
		dlog.ErrorLog.Println("error finding guild:", err)
		return err
	}

	// Check if the user is in a voice channel
	vs, err := a.VoiceChannelCheck(s, i)
	if err != nil {
		dlog.ErrorLog.Println("error checking voice channel:", err)
		return err
	}

	user := i.Member.User
	err = a.model.AddUser(user)
	if err != nil {
		dlog.ErrorLog.Println("error adding user:", err)
		return err
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
		Content:    "➡ Currently Playing audio by " + user.Mention(),
		Components: content,
	}

	// Send the message (+stop button)
	st, err := a.SendMessageComplex(msg, s, i, false)
	if err != nil {
		dlog.ErrorLog.Println("error sending message:", err)
		return err
	}

	err = a.UpdateInteractionResponse("🎶  Playing audio", s, i)
	if err != nil {
		dlog.ErrorLog.Println("error updating interaction response:", err)
		return err
	}

	err = a.DeleteOldStopSoundButtons(s, st)
	if err != nil {
		dlog.ErrorLog.Println("error deleting stop sound buttons:", err)
	}

	dlog.InfoLog.Printf("User: %s played sound", i.Member.User.GlobalName)
	// Play the sound
	err = a.PlaySound(s, i, guild.ID, vs.ChannelID, audioName)
	if err != nil {
		dlog.ErrorLog.Println("error playing sound:", err)
		return err
	}

	time.Sleep(250 * time.Millisecond)

	st, err = s.ChannelMessage(st.ChannelID, st.ID)
	if err == nil {
		err = s.ChannelMessageDelete(st.ChannelID, st.ID)
		if err != nil {
			dlog.ErrorLog.Println("error deleting stop sound button after sound finished:", err)
		}
	}

	return nil
}

// SyncDatabaseWithFileSystem synchronizes the database with the filesystem.
func (a *API) SyncDatabaseWithFileSystem(folderMap map[string][]string) error {
	existingCategories, _ := a.model.GetCategoriesM()

	var wg sync.WaitGroup                                        // WaitGroup to wait for all Goroutines to finish
	soundProcessingChan := make(chan struct{}, runtime.NumCPU()) // limit parallelism to maximum available threads Goroutines

	for folder, files := range folderMap {
		categoryID, err := a.getOrCreateCategoryID(folder, existingCategories)
		if err != nil {
			dlog.WarningLog.Printf("Failed to process category %s: %v", folder, err)
			continue
		}

		for _, file := range files {
			wg.Add(1)                         // increment the WaitGroup counter
			soundProcessingChan <- struct{}{} // acquire a token to limit concurrency

			go func(folder, file string, categoryID int) {
				defer wg.Done()
				defer func() { <-soundProcessingChan }()

				if err := a.processSoundFile(folder, file, categoryID); err != nil {
					dlog.WarningLog.Printf("failed to process file %s in folder %s: %v", file, folder, err)
				}
			}(folder, file, categoryID)
		}
	}

	wg.Wait()                                          // wait for all Goroutines to finish
	a.cleanUpCategories(folderMap, existingCategories) // remove orphaned entries from the database

	return nil
}

// getOrCreateCategoryID retrieves or creates a category and returns its ID.
func (a *API) getOrCreateCategoryID(folder string, existingCategories map[string]int) (int, error) {
	if categoryID, exists := existingCategories[folder]; exists {
		return categoryID, nil
	}
	if err := a.model.AddCategory(folder); err != nil {
		if !strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return 0, err
		}
	}
	return a.model.GetCategoryByID(folder), nil
}

// processSoundFile handles the conversion and addition of a sound file.
func (a *API) processSoundFile(folder, file string, categoryID int) error {
	soundPathMP3 := filepath.Join(a.model.Config.SoundsDir, folder, file+".mp3")
	soundPathDCA := filepath.Join(a.model.Config.SoundsDir, folder, file+".dca")

	if _, err := os.Stat(soundPathMP3); os.IsNotExist(err) {
		return fmt.Errorf("MP3 file not found: %s", soundPathMP3)
	}

	fileHash, err := model.ComputeFileHash(soundPathMP3)
	if err != nil {
		return fmt.Errorf("failed to compute hash: %v", err)
	}

	if err := a.ConvertMP3ToDCA(file, folder); err != nil {
		return fmt.Errorf("error converting mp3 to dca: %v", err)
	}

	soundBytes, err := os.ReadFile(soundPathDCA)
	if err != nil {
		return fmt.Errorf("failed to read DCA file: %v", err)
	}

	if err := a.model.AddSound(categoryID, file, fileHash, soundBytes); err != nil {
		return fmt.Errorf("failed to add sound: %v", err)
	}

	err = os.Remove(soundPathDCA)
	if err != nil {
		return fmt.Errorf("failed to remove dca file after insert into db: %v", err)
	}

	err = os.Rename(soundPathMP3, soundPathMP3+".done")
	if err != nil {
		return fmt.Errorf("failed to rename mp3 file after insert into db: %v", err)
	}

	return nil
}

// cleanUpCategories removes database categories (subfolders) that no longer exist in the filesystem.
func (a *API) cleanUpCategories(folderMap map[string][]string, existingCategories map[string]int) {
	for folder, categoryID := range existingCategories {
		if _, exists := folderMap[folder]; !exists {
			if err := a.model.RemoveCategory(categoryID); err == nil {
				dlog.InfoLog.Printf("Removed category %s (ID: %d)", folder, categoryID)
			}
			continue
		}
	}
}

func (a *API) VoiceChannelCheck(s *discordgo.Session, i *discordgo.InteractionCreate) (*discordgo.VoiceState, error) {
	userInVS := false
	var voiceState *discordgo.VoiceState
	// Find the guild for that channel
	guild, err := s.State.Guild(model.Meta.Guild.ID)
	if err != nil {
		dlog.ErrorLog.Println("error finding guild:", err)
		return voiceState, err
	}

	user := i.Member.User

	for _, vs := range guild.VoiceStates {
		if vs.UserID == user.ID {
			userInVS = true
			voiceState = vs
		}
	}
	if !userInVS {
		// If the user is not in a voice channel, send an error message and avoid processing the audio
		dlog.InfoLog.Printf("User %s tried to play sound but is not in a voice channel", user.GlobalName)
		msg := "You need to be in a voice channel to play sounds " + user.Mention()
		_, err = a.SendMessage(msg, s, i, false)
		if err != nil {
			dlog.ErrorLog.Println("error sending message:", err)
		}
		return voiceState, fmt.Errorf("user not in voice channel, quitting early to avoid delay")
	}

	return voiceState, nil
}
