package utils

import (
	"fmt"
	"os/exec"

	"github.com/cyb3rplis/discord-bot-go/model"

	"github.com/cyb3rplis/discord-bot-go/logger"
)

func TextToSpeech(text string) error {
	// piper in path ./tts has to be present
	// also the model
	ttsOutput := model.Bot.Config.TTSInput // this is correct, the value gets used twice
	piperPath := fmt.Sprintf("%s/piper", model.Bot.Config.TTS)
	speechFile := fmt.Sprintf("%s/de_DE-thorsten-medium.onnx", model.Bot.Config.TTS)

	// Construct the shell command to echo and pipe it
	cmd := exec.Command("sh", "-c", fmt.Sprintf("echo %q | %s --model %s --output_file %s", text, piperPath, speechFile, ttsOutput))
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to run piper, make sure it is installed (.tts folder): %w", err)
	}

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("error creating tts wav audio: %w", err)
	}

	// Print the output if the command succeeds
	logger.InfoLog.Printf("Successfully converted text to speech: %s", ttsOutput)

	return nil
}

func WAVtoDCA() error {
	ttsInput := model.Bot.Config.TTSInput
	ttsOutput := model.Bot.Config.TTSOutput

	// Construct the shell command to echo and pipe it
	// ffmpeg -i tta.wav -f s16le -ar 48000 -ac 2 pipe:1 | ./dca > tta.dca
	cmd := exec.Command("sh", "-c", fmt.Sprintf("ffmpeg -i %s -filter:a \"loudnorm=I=-14:LRA=7:TP=-2, compand=attacks=0:points=-80/-80|-10/-5|0/-1\" -f s16le -ar 48000 -ac 2 pipe:1 | dca > %s", ttsInput, ttsOutput))

	// Run the command
	err := cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to run ffmpeg, make sure it is installed: %w", err)
	}

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("error creating tts mp3 audio: %w", err)
	}

	// Print the output if the command succeeds
	logger.InfoLog.Printf("Successfully converted wav to dca: %s", ttsOutput)

	return nil
}
