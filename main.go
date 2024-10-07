package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/faiface/beep"
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
)

//GOTO: https://discord.com/developers/applications/
var Token = "<your-token>"

func main() {
	//new session
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating discord session,", err)
		return
	}

	//event handler for messages
	dg.AddHandler(messageHandler)

	//websocket to Discord-API
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}

	fmt.Println("Bot is now running. Press CTRL+C to exit.")

	//ctrl stop
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-stop

	dg.Close()
}

// handler for incoming messages
func messageHandler(s *discordgo.Session, m *discordgo.MessageCreate) {
	//dont allow bot to response to own messages
	if m.Author.ID == s.State.User.ID {
		return
	}
	if len(m.Content) == 0 {
		fmt.Println("empty message > no content")
		return
	}

	if m.Content == ".sound list" {
		sounds, err := listSounds()
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("error listing sounds: %v", err))
			return
		}
		if len(sounds) == 0 {
			s.ChannelMessageSend(m.ChannelID, "No sounds found in the directory.")
			return
		}
		soundList := "Sounds List:\n"
		for _, sound := range sounds {
			soundList += "- " + sound[:len(sound)-4] + "\n" // Remove .mp3 extension
		}

		// Send the list of sounds to the channel
		s.ChannelMessageSend(m.ChannelID, soundList)
		return
	}

	if strings.HasPrefix(m.Content, ".sound") {
		args := strings.Split(m.Content, " ")
		fmt.Println("sound arguments:  ", args)

		if len(args) < 2 || len(args) > 2 {
			s.ChannelMessageSend(m.ChannelID, "Usage: .sound <soundname>")
			return
		}

		soundFile := fmt.Sprintf("./sounds/%s.mp3", args[1])

		//play sound
		err := playSound(soundFile)
		if err != nil {
			//s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("error playing sound: %v", err))
			s.ChannelMessageSend(m.ChannelID, "error playing sound. Check if sound exists")
		} else {
			s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("playing sound: %s", args[1]))
		}
	}
}

func listSounds() ([]string, error) {
	files, err := ioutil.ReadDir("./sounds")
	if err != nil {
		return nil, err
	}
	var soundFiles []string
	for _, file := range files {
		// Only include .mp3 files
		if filepath.Ext(file.Name()) == ".mp3" {
			soundFiles = append(soundFiles, file.Name())
		}
	}
  
	return soundFiles, nil
}

//play sound test
func playSound(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	streamer, format, err := mp3.Decode(file)
	if err != nil {
		return fmt.Errorf("error decoding mp3: %v", err)
	}
	defer streamer.Close()

	speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))

	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))

	<-done
	return nil
}
