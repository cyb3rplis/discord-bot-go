package message

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/cyb3rplis/discord-bot-go/config"

	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/sound"
)

// This function will be called (due to AudioMessageHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func AudioMessageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {

	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}

	if len(m.Content) == 0 {
		fmt.Println("Empty content..")
		return
	}

	//LIST SOUNDS
	if m.Content == fmt.Sprintf("%slist", config.GetValueString("general", "prefix", ".")) {
		sounds, err := sound.ListSounds()
		if err != nil {
			fmt.Println("error listing sounds:", err)
			return
		}
		if len(sounds) == 0 {
			_, err := s.ChannelMessageSend(m.ChannelID, "No sounds found.")
			if err != nil {
				fmt.Println("error sending message:", err)
			}
			return
		}

		soundList := fmt.Sprintf("> Available sounds (play with: %ssound <sound-name>) \n", config.GetValueString("general", "prefix", "."))
		for _, soundName := range sounds {
			soundList += fmt.Sprintf("> * %s", soundName[:len(soundName)-4]+"\n")
		}
		_, err = s.ChannelMessageSend(m.ChannelID, soundList)
		if err != nil {
			fmt.Println("error sending message:", err)
		}

		//PLAY SOUND
		// check if the message is "<prefix>sound"
	} else if strings.HasPrefix(m.Content, fmt.Sprintf("%ssound", config.GetValueString("general", "prefix", "."))) {

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

		args := strings.Split(m.Content, " ")
		if len(args) < 2 || args[1] == "" || len(args) > 2 {
			_, err := s.ChannelMessageSend(m.ChannelID, "Usage: .sound <sound_name>")
			if err != nil {
				fmt.Println("error sending message:", err)
			}
			return
		}

		var validPattern = regexp.MustCompile(`^[a-z\-0-9]+$`)
		if !validPattern.MatchString(args[1]) {
			_, err := s.ChannelMessageSend(m.ChannelID, "sound contains invalid characters, only [a-z0-9] allowed")
			if err != nil {
				fmt.Println("error sending message:", err)
			}
			return
		}

		soundFile := fmt.Sprintf("%s/%s.dca", config.GetValueString("general", "sounds_dir", "-"), args[1])
		fmt.Println("playing sound file: ", soundFile)

		// Look for the message sender in that guild's current voice states.
		for _, vs := range g.VoiceStates {
			if vs.UserID == m.Author.ID {
				err = sound.PlaySound(s, m, g.ID, vs.ChannelID, soundFile)
				if err != nil {
					fmt.Println("error playing sound:", err)
				}
				return
			}
		}
	}
}
