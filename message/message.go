package message

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/cyb3rplis/discord-bot-go/config"

	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/sound"
)

// This function will be called (due to AudioMessageHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func AudioMessageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	prefix := config.GetValueString("general", "prefix", ".")

	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	if len(m.Content) == 0 {
		fmt.Println("Empty content..")
		return
	}

	// Extract the command and arguments
	args := strings.Split(m.Content, " ")
	command := args[0]

	switch {
	//if the command starts with the prefix and is not a list or stop command
	case command == fmt.Sprintf("%shelp", prefix):
		// Default case: show help message
		_, err := s.ChannelMessageSend(m.ChannelID, "🧐 Usage: \n > » List Categories: <.l> \n > » List Sounds: <.c> <.l> \n > » Play Sound: <.c> <sound_name>")
		if err != nil {
			fmt.Println("error sending message:", err)
		}
	// LIST CATEGORIES
	case command == fmt.Sprintf("%sl", prefix):
		soundCategories, err := sound.ListSoundsCategories()
		if err != nil {
			fmt.Println("error listing sound categories:", err)
			return
		}
		if len(soundCategories) == 0 {
			_, err := s.ChannelMessageSend(m.ChannelID, "No sound categories found.")
			if err != nil {
				fmt.Println("error sending message:", err)
			}
			return
		}

		list := fmt.Sprintf("> Available Sound Categories \n")
		for _, folder := range soundCategories {
			list += fmt.Sprintf("> * %s\n", folder)
		}
		_, err = s.ChannelMessageSend(m.ChannelID, list)
		if err != nil {
			fmt.Println("error sending message:", err)
		}
	// LIST SOUNDS IN SUBFOLDER
	case len(args) == 1 && strings.HasPrefix(command, prefix):
		subfolder := strings.TrimPrefix(command, prefix)
		sounds, err := sound.ListSoundsInSubfolder(subfolder)
		if err != nil {
			fmt.Println("error listing sounds in subfolder:", err)
			return
		}
		if len(sounds) == 0 {
			_, err := s.ChannelMessageSend(m.ChannelID, "No sounds found in this category.")
			if err != nil {
				fmt.Println("error sending message:", err)
			}
			return
		}
		soundList := fmt.Sprintf("> Available sounds for this category %s \n Usage: %s%s <sound-name> \n", subfolder, prefix, subfolder)
		for _, soundName := range sounds {
			soundList += fmt.Sprintf("> * %s\n", soundName[:len(soundName)-4])
		}
		_, err = s.ChannelMessageSend(m.ChannelID, soundList)
		if err != nil {
			fmt.Println("error sending message:", err)
		}
	// PLAY SOUND
	case len(args) == 2 && strings.HasPrefix(command, prefix):
		subfolder := strings.TrimPrefix(command, prefix)
		soundName := args[1]
		// Validate the sound name
		var validPattern = regexp.MustCompile(`^[a-z\-0-9]+$`)
		if !validPattern.MatchString(soundName) {
			_, err := s.ChannelMessageSend(m.ChannelID, "Sound contains invalid characters, only [a-z0-9] allowed")
			if err != nil {
				fmt.Println("error sending message:", err)
			}
			return
		}

		// Check if subfolder exists
		subfolderPath := fmt.Sprintf("%s/%s", config.GetValueString("general", "sounds_dir", "-"), subfolder)
		if _, err := os.Stat(subfolderPath); os.IsNotExist(err) {
			_, err := s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Category %s does not exist. \n  Usage: \n > List Categories: <.l> \n", subfolder))
			if err != nil {
				fmt.Println("error sending message:", err)
			}
			return
		}

		// Construct the sound file path
		soundFile := fmt.Sprintf("%s/%s/%s.dca", config.GetValueString("general", "sounds_dir", "-"), subfolder, soundName)

		// Find the channel that the message came from.
		c, err := s.State.Channel(m.ChannelID)
		if err != nil {
			// Could not find channel.
			return
		}

		// Find the guild for that channel.
		g, err := s.State.Guild(c.GuildID)
		if err != nil {
			// Could not find guild.
			return
		}

		// Look for the message sender in that guild's current voice states.
		for _, vs := range g.VoiceStates {
			if vs.UserID == m.Author.ID {
				err = sound.PlaySound(s, m, g.ID, vs.ChannelID, soundFile, soundName)
				if err != nil {
					fmt.Println("error playing sound:", err)
				}
				return
			}
		}

	default:
		return
	}
}
