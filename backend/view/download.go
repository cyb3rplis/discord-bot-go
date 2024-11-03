package view

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/logger"
)

type Download struct {
	URL       string `json:"url"`
	Start     string `json:"start"`
	End       string `json:"end"`
	Category  string `json:"category"`
	SoundName string `json:"sound_name"`
}

func (a *API) DownloadAndConvertAudio(download Download, s *discordgo.Session, m *discordgo.MessageCreate) error {
	message := "🎶  Preparing Audio, this might take a few seconds..."
	st := a.NewMessageRoutine(".audio-dl", message, s, m)
	err := s.ChannelTyping(m.ChannelID)
	if err != nil {
		logger.ErrorLog.Println("error setting typing status:", err)
	}

	// setting up a context to cancel the process after x seconds
	timeout := time.Duration(a.model.Config.YTTimeout) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	var cmd *exec.Cmd

	if download.Start != "" && download.End != "" {
		cmd = exec.CommandContext(ctx, "bash", "-c", fmt.Sprintf("yt-dlp -x --audio-format mp3 --download-sections \"*%s-%s\" --force-overwrites -o %s %s", download.Start, download.End, a.model.Config.YTTemp, download.URL))
	} else {
		cmd = exec.CommandContext(ctx, "bash", "-c", fmt.Sprintf("yt-dlp -x --audio-format mp3 --force-overwrites -o %s %s", a.model.Config.YTTemp, download.URL))
	}

	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to run yt-dlp, make sure it is installed (python venv): %w", err)
	}

	err = cmd.Wait()
	if ctx.Err() == context.DeadlineExceeded {
		a.DeleteMessageRoutine(s, ".audio-dl")
		a.DeleteMessageRoutine(s, ".audio-dl-err")

		message := "❗  Downloading audio failed, song is probably too long <@" + m.Author.ID + ">?"
		a.NewMessageRoutine(".audio-dl-err", message, s, m)
		return fmt.Errorf("error downloading audio: %w", err)
	} else if err != nil {
		a.DeleteMessageRoutine(s, ".audio-dl")
		a.DeleteMessageRoutine(s, ".audio-dl-err")

		message := "❗  Downloading audio failed, did you use the correct link <@" + m.Author.ID + ">?"
		a.NewMessageRoutine(".audio-dl-err", message, s, m)
		return fmt.Errorf("error downloading audio: %w", err)
	}

	a.DeleteMessageRoutine(s, ".audio-dl-err")

	cmd = exec.Command("bash", "-c", fmt.Sprintf("ffmpeg -i %s -filter:a \"loudnorm=I=-14:LRA=7:TP=-2, compand=attacks=0:points=-80/-80|-10/-5|0/-1\" -f s16le -ar 48000 -ac 2 pipe:1 | dca > %s", a.model.Config.YTTemp, a.model.Config.YTOutput))
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to to convert audio from mp3 to dca: %w", err)
	}

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("error converting audio: %w", err)
	}

	err = s.ChannelMessageDelete(st.ChannelID, st.ID)
	if err != nil {
		logger.ErrorLog.Println("error deleting preparing messages: ", err)
	}

	err = a.model.DeleteMessageID(st.ID)
	if err != nil {
		logger.ErrorLog.Printf("error deleting message from DB: %v", err)
	}

	return nil
}
