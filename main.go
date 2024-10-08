package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/jonas747/dca"
)

//GOTO: https://discord.com/developers/applications/
var Token = "<your-token-here>"

func main() {
	//new session
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
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
		fmt.Println("Empty message content.")
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
		soundList := "sounds List:\n"
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
		// find the guild and voice state to identify which channel the user is in
		guildID := m.GuildID
		userID := m.Author.ID
		voiceState := getVoiceState(s, guildID, userID)

		if voiceState == nil {
			s.ChannelMessageSend(m.ChannelID, "You need to be in a voice channel to play sounds.")
			return
		}

		//play sound
		err := playSound(s, guildID, voiceState.ChannelID, soundFile)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, "error playing sound. Check if sound exists")
			return
		}
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("playing sound: %s", args[1]))
	}
}

// function to find the user's voice state in a guild
func getVoiceState(s *discordgo.Session, guildID, userID string) *discordgo.VoiceState {
	guild, err := s.State.Guild(guildID)
	if err != nil {
		return nil
	}

	for _, vs := range guild.VoiceStates {
		if vs.UserID == userID {
			return vs
		}
	}

	return nil
}

func listSounds() ([]string, error) {
	files, err := ioutil.ReadDir("./sounds")
	if err != nil {
		return nil, err
	}

	var soundFiles []string
	baseDir, err := filepath.Abs("./sounds")
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		// resolve the absolute path of each file
		filePath, err := filepath.Abs(filepath.Join("./sounds", file.Name()))
		if err != nil {
			return nil, err
		}
		// ensure the file is inside the base directory
		relPath, err := filepath.Rel(baseDir, filePath)
		if err != nil {
			return nil, err
		}
		// check if the relative path does not escape the base directory
		if relPath == ".." || relPath[:3] == "../" {
			return nil, errors.New("potential path traversal detected")
		}
		// only append if symlink
		if filepath.Ext(file.Name()) == ".mp3" && !isSymlink(filePath) {
			soundFiles = append(soundFiles, file.Name())
		}
	}

	return soundFiles, nil
}

//play sound test
func playSound(s *discordgo.Session, guildID, channelID, soundFile string) error {
	// connect to the voice channel
	vc, err := s.ChannelVoiceJoin(guildID, channelID, false, true)
	if err != nil {
		return fmt.Errorf("failed to join voice channel: %w", err)
	}

	// open the sound file
	sound, err := os.Open(soundFile)
	if err != nil {
		return fmt.Errorf("failed to open sound file: %w", err)
	}
	defer sound.Close()

	// create a new DCA encoder
	encodeSession, err := dca.EncodeFile(soundFile, dca.StdEncodeOptions)
	if err != nil {
		return fmt.Errorf("failed to encode sound file: %w", err)
	}
	defer encodeSession.Cleanup()

	// play the audio stream
	done := make(chan error)
	dca.NewStream(encodeSession, vc, done)

	// wait for the sound to finish playing
	err = <-done
	if err != nil && err != io.EOF {
		return fmt.Errorf("error while playing sound: %w", err)
	}

	// disconnect after playing
	vc.Disconnect()

	return nil
}

// helper
func isSymlink(path string) bool {
	fileInfo, err := os.Lstat(path)
	if err != nil {
		return false
	}
	return (fileInfo.Mode() & os.ModeSymlink) != 0
}
