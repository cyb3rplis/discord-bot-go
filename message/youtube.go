package message

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/logger"
	"github.com/cyb3rplis/discord-bot-go/model"
	"github.com/cyb3rplis/discord-bot-go/sound"
	"github.com/cyb3rplis/discord-bot-go/utils"
)

func HandleYoutube(s *discordgo.Session, m *discordgo.MessageCreate, arg, command string) error {
	prefix := model.Bot.Config.Prefix
	if len(arg) == 0 {
		message := fmt.Sprintf("🎶  Youtube: Type the URL of the video you want to play\n > » %syoutube https://...\n", prefix)
		utils.NewMessageRoutine(command, message, s, m, true)
		return nil
	}
	if !strings.Contains(arg, "https://") {
		message := fmt.Sprintf("🎶  Youtube: Invalid URL\n > » %syoutube https://...\n", prefix)
		utils.NewMessageRoutine(command, message, s, m, true)
		return nil
	}
	utils.CleanUpSoundFile("youtube")
	err := utils.VoiceChannelCheck(s, m)
	if err != nil {
		logger.ErrorLog.Println("error checking voice channel:", err)
		return err
	}
	err = DownloadAndConvertYoutubeAudio(arg, s, m)
	if err != nil {
		logger.ErrorLog.Println("error loading youtube audio:", err)
		return err
	}

	err = sound.PlayCustomAudio(s, m, "youtube")
	if err != nil {
		logger.ErrorLog.Println("error playing youtube audio:", err)
	}

	return nil
}

func DownloadAndConvertYoutubeAudio(videoURL string, s *discordgo.Session, m *discordgo.MessageCreate) error {
	message := "🎶  Preparing Youtube Audio, this might take a few seconds..."
	st := utils.NewMessageRoutine(".youtubedl", message, s, m, true)
	s.ChannelTyping(m.ChannelID)

	// setting up a context to cancel the process after x seconds
	timeout := time.Duration(model.Bot.Config.YTTimeout) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Create ffmpeg and dcaenc pipeline to convert YouTube stream to DCA format
	cmd := exec.CommandContext(ctx, "bash", "-c", fmt.Sprintf("yt-dlp -x --audio-format mp3 --force-overwrites -o %s %s", model.Bot.Config.YTTemp, videoURL))
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to run yt-dlp, make sure it is installed (python venv): %w", err)
	}

	err = cmd.Wait()
	if ctx.Err() == context.DeadlineExceeded {
		utils.DeleteMessageRoutine(s, ".youtubedl")
		utils.DeleteMessageRoutine(s, ".youtubedlerr")

		message := "❗  Downloading Youtube Audio failed, song is probably too long <@" + m.Author.ID + ">?"
		utils.NewMessageRoutine(".youtubedlerr", message, s, m, false)
		return fmt.Errorf("error downloading youtube audio: %w", err)
	} else if err != nil {
		utils.DeleteMessageRoutine(s, ".youtubedl")
		utils.DeleteMessageRoutine(s, ".youtubedlerr")

		message := "❗  Downloading Youtube Audio failed, did you use the correct link <@" + m.Author.ID + ">?"
		utils.NewMessageRoutine(".youtubedlerr", message, s, m, false)
		return fmt.Errorf("error downloading youtube audio: %w", err)
	}

	utils.DeleteMessageRoutine(s, ".youtubedlerr")

	cmd = exec.Command("bash", "-c", fmt.Sprintf("ffmpeg -i %s -filter:a \"loudnorm=I=-14:LRA=7:TP=-2, compand=attacks=0:points=-80/-80|-10/-5|0/-1\" -f s16le -ar 48000 -ac 2 pipe:1 | dca > %s", model.Bot.Config.YTTemp, model.Bot.Config.YTOutput))
	err = cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to to convert youtube audio from mp3 to dca: %w", err)
	}

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("error converting youtube audio: %w", err)
	}

	err = s.ChannelMessageDelete(st.ChannelID, st.ID)
	if err != nil {
		logger.ErrorLog.Println("error deleting preparing messages: ", err)
	}

	err = utils.DeleteMessageID(st.ID)
	if err != nil {
		logger.ErrorLog.Printf("error deleting message from DB: %v", err)
	}

	return nil
}
