package view

import (
	"fmt"
	"os/exec"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/logger"
)

func (a *API) HandleTTS(s *discordgo.Session, m *discordgo.MessageCreate, command string) error {
	prefix := a.model.Config.Prefix
	if m.Content == fmt.Sprintf("%stts", prefix) {
		msg := fmt.Sprintf("📢  TTS: Type text which will be played via Text to Speech in your Voice Channel\n > » %stts \"This is Text to Speech\"\n", prefix)

		a.NewMessageRoutine(command+"help", msg, s, m)
		_ = s.MessageReactionAdd(m.ChannelID, m.ID, "✅")
		return nil
	}

	_ = s.ChannelMessageDelete(m.ChannelID, m.ID)

	// Check if the user is in the Gulag
	user, err := a.model.GetUserFromUsername(m.Message.Author.GlobalName)
	if err != nil {
		logger.ErrorLog.Println("error getting user from username:", err)
	} else {
		if remaining, ok := IsUserInGulag(user); ok {
			user.Remaining = remaining
			msg := fmt.Sprintf("<@"+user.ID+"> you are in the Gulag for another %s", user.Remaining)
			a.NewMessageRoutine(".gulag"+user.ID, msg, s, &discordgo.MessageCreate{Message: m.Message})
			return fmt.Errorf("user is in the Gulag: %s", user.ID)
		}
	}

	ttsText := m.Content[5:len(m.Content)]
	if strings.HasPrefix(ttsText, "\"") && strings.HasSuffix(ttsText, "\"") {
		pattern := `^\"[öäüÖÄÜa-zA-Z0-9\.!:,? ]+\"$`
		re, err := regexp.Compile(pattern)
		if err != nil {
			logger.ErrorLog.Println("error compiling regex:", err)
			return err
		}
		if re.MatchString(ttsText) {
			err := a.VoiceChannelCheck(s, m)
			if err != nil {
				logger.ErrorLog.Println("error checking voice channel:", err)
				return err
			}

			a.model.CleanUpSoundFile("tts")

			err = a.textToSpeech(ttsText)
			if err != nil {
				logger.ErrorLog.Println("error converting text to speech:", err)
				return err
			}

			err = a.wavToDCA()
			if err != nil {
				logger.ErrorLog.Println("error converting wav to dca:", err)
				return err
			}

			// play sound and clean up files
			err = a.PlayCustomAudio(s, m, "tts")
			if err != nil {
				logger.ErrorLog.Println("error playing youtube audio:", err)
			}
		} else {
			logger.InfoLog.Println("TTS Text does not match regex pattern: ", ttsText)
			return err
		}
		return nil
	}

	msg := fmt.Sprintf("Text has to be in Quotes\n > » %stts \"This is Text to Speech\"\n", prefix)
	a.NewMessageRoutine(command+"quote", msg, s, m)
	return nil
}

func (a *API) textToSpeech(text string) error {
	piperPath := fmt.Sprintf("%s/piper", a.model.Config.TTS)
	speechFile := fmt.Sprintf("%s/de_DE-thorsten-medium.onnx", a.model.Config.TTS)

	// Construct the shell command to echo and pipe it
	cmd := exec.Command("sh", "-c", fmt.Sprintf("echo %q | %s --model %s --output_file %s", text, piperPath, speechFile, a.model.Config.TTSTemp))
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to run piper, make sure it is installed (.tts folder): %w", err)
	}

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("error creating tts wav audio: %w", err)
	}

	// Print the output if the command succeeds
	logger.InfoLog.Printf("Successfully converted text to speech: %s", a.model.Config.TTSTemp)

	return nil
}

func (a *API) wavToDCA() error {
	// Construct the shell command to echo and pipe it
	// ffmpeg -i tta.wav -f s16le -ar 48000 -ac 2 pipe:1 | dca > tta.dca
	cmd := exec.Command("sh", "-c", fmt.Sprintf("ffmpeg -i %s -filter:a \"loudnorm=I=-14:LRA=7:TP=-2, compand=attacks=0:points=-80/-80|-10/-5|0/-1\" -f s16le -ar 48000 -ac 2 pipe:1 | dca > %s", a.model.Config.TTSTemp, a.model.Config.TTSOutput))

	// Run the command
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to run ffmpeg, make sure it is installed: %w", err)
	}

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("error creating tts dac audio: %w", err)
	}

	// Print the output if the command succeeds
	logger.InfoLog.Printf("Successfully converted wav to dca: %s", a.model.Config.TTSOutput)

	return nil
}
