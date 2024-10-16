package utils

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/cyb3rplis/discord-bot-go/config"
	"github.com/cyb3rplis/discord-bot-go/logger"
)

func TextToSpeech(text string) error {
	// piper in path ./tts has to be present
	// also the model
	cfg := config.GetConfig()
	ttsOutput := cfg.TTSInput // this is correct, the value gets used twice
	piperPath := fmt.Sprintf("%s/piper", cfg.TTS)
	speechFile := fmt.Sprintf("%s/de_DE-thorsten-medium.onnx", cfg.TTS)

	// Construct the shell command to echo and pipe it
	cmd := exec.Command("sh", "-c", fmt.Sprintf("echo %q | %s --model %s --output_file %s", text, piperPath, speechFile, ttsOutput))

	// Run the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.ErrorLog.Println("Error:", err)
		logger.ErrorLog.Println("Output:", string(output))
		return err
	}

	// Print the output if the command succeeds
	logger.InfoLog.Printf("Successfully converted text to speech: %s", ttsOutput)

	return nil
}

func WAVtoDCA() error {
	cfg := config.GetConfig()
	ttsInput := cfg.TTSInput
	ttsOutput := cfg.TTSOutput
	dca := cfg.DCA

	// Construct the shell command to echo and pipe it
	// ffmpeg -i tta.wav -f s16le -ar 48000 -ac 2 pipe:1 | ./dca > tta.dca
	cmd := exec.Command("sh", "-c", fmt.Sprintf("ffmpeg -i %s -f s16le -ar 48000 -ac 2 pipe:1 | %s > %s", ttsInput, dca, ttsOutput))

	// Run the command
	output, err := cmd.CombinedOutput()
	if err != nil {
		logger.ErrorLog.Println("Error:", err)
		logger.ErrorLog.Println("Output:", string(output))
		return err
	}

	// Print the output if the command succeeds
	logger.InfoLog.Printf("Successfully converted wav to dca: %s", ttsOutput)

	return nil
}

func CleanUpTTS() error {
	cfg := config.GetConfig()
	ttsInput := cfg.TTSInput
	ttsOutput := cfg.TTSOutput

	err := os.Remove(ttsInput)
	if err != nil {
		// Handle error if file deletion fails
		logger.ErrorLog.Printf("Error deleting file: %v\n", err)
		return err
	}

	err = os.Remove(ttsOutput)
	if err != nil {
		// Handle error if file deletion fails
		logger.ErrorLog.Printf("Error deleting file: %v\n", err)
		return err
	}

	return nil
}
