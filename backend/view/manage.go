package view

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/cyb3rplis/discord-bot-go/config"
	"github.com/cyb3rplis/discord-bot-go/model"

	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/dlog"
)

type Download struct {
	URL       string `json:"url"`
	Start     string `json:"start"`
	End       string `json:"end"`
	Category  string `json:"category"`
	SoundName string `json:"sound_name"`
}

func (a *API) PromptInteractionManage(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type == discordgo.InteractionApplicationCommand {
		switch i.ApplicationCommandData().Name {
		case "manage":
			//case create, delete, move:
			if len(i.ApplicationCommandData().Options) == 0 {
				err := a.handleList(s, i)
				if err != nil {
					dlog.ErrorLog.Println("error handling list:", err)
				}
			} else {
				option := i.ApplicationCommandData().Options[0]
				switch option.Name {
				case "create":
					err := a.SendInteractionRespond("👉  Creating button from URL.", s, i)
					if err != nil {
						dlog.ErrorLog.Println("error executing buttons command:", err)
					}
					url := option.Options[0].StringValue()
					soundName := option.Options[1].StringValue()
					category := option.Options[2].StringValue()
					startTime := "0"
					endTime := "0"
					if len(option.Options) > 3 {
						startTime = option.Options[3].StringValue()
					}
					if len(option.Options) > 4 {
						endTime = option.Options[4].StringValue()
					}
					// Download and convert the audio
					download := Download{URL: url, Start: startTime, End: endTime, Category: category, SoundName: soundName}
					// Check if the sound directory exists, if not create it
					categoryFolder := filepath.Join(config.GetConfig().SoundsDir, download.Category)
					if _, err := os.Stat(categoryFolder); os.IsNotExist(err) {
						err = os.MkdirAll(categoryFolder, os.ModePerm)
						if err != nil {
							return
						}
					}
					err = a.DownloadAudio(download)
					if err != nil {
						dlog.ErrorLog.Println("error loading audio:", err)
						return
					}
					err = a.UpdateInteractionResponse("🎶  Audio is ready, creating button...", s, i)
					if err != nil {
						dlog.ErrorLog.Println("error[manage1] sending message:", err)
					}
					time.Sleep(3 * time.Second)

					err = a.ConvertMP3ToDCA(soundName, download.Category)
					if err != nil {
						dlog.ErrorLog.Println("error converting audio:", err)
						return
					}

					soundPath := filepath.Join(a.model.Config.SoundsDir, download.Category, download.SoundName+".dca")
					fileData, err := os.ReadFile(soundPath)
					if err != nil {
						dlog.WarningLog.Printf("Failed to read sound file %s: %v", soundPath, err)

						// sometimes there is a race condition where the cronjob of the sound insert is faster
						// that causes the mp3 file to be renamed to .done, which causes the dca conversion to fail
						// but the dca conversion happened already, since the mp3 was renamed.
						soundID, err := a.model.GetSoundIDByName(soundName)
						if err != nil {
							dlog.WarningLog.Printf("Failed to get sound ID for %s: %v", soundName, err)
							message := "🎶  Something went wrong, try again"
							err = a.UpdateInteractionResponse(message, s, i)
							if err != nil {
								dlog.ErrorLog.Println("error[manage2] updating interaction response:", err)
							}
							return
						}

						// sound is not in the DB, there was an actual error
						if soundID == "" {
							dlog.WarningLog.Printf("Sound %s not found in database", soundName)
							message := "🎶  Something went wrong, try again"
							err = a.UpdateInteractionResponse(message, s, i)
							if err != nil {
								dlog.ErrorLog.Println("error[manage3] updating interaction response:", err)
							}
							return
						}

						dlog.InfoLog.Printf("Sound %s already exists, just send response to user", soundName)
						// if we reached this point, we can assume that the sound exists already
						// Build button for the new sound
						message := "🎶  New sound added, try it out"
						component := model.BuildSingleSoundButton(soundName, category, discordgo.SuccessButton)

						err = a.UpdateInteractionResponseWithButton(message, component, s, i)
						if err != nil {
							dlog.ErrorLog.Printf("error[manage4] updating interaction response: %v", err)
						}

						return
					}

					fileHash, err := model.ComputeFileHash(filepath.Join(a.model.Config.SoundsDir, download.Category, download.SoundName+".mp3")) //use mp3 hash for comparing
					if err != nil {
						dlog.WarningLog.Printf("Failed to compute hash for file %s: %v", download.SoundName, err)
						return
					}

					var categoryID int
					existingCategories, _ := a.model.GetCategoriesM() // Get current folders/categories in DB
					existingSounds := a.model.GetSoundsM()
					// Check if the folder (category) exists in the database
					if dbCategoryID, exists := existingCategories[download.Category]; exists {
						categoryID = dbCategoryID // The folder already exists
					} else {
						// The folder doesn't exist in the database, so we need to add it
						if err := a.model.AddCategory(download.Category); err != nil {
							if !strings.Contains(err.Error(), "UNIQUE constraint failed") {
								dlog.WarningLog.Printf("Failed to add category %s: %v", download.Category, err)
							}
							return
						}
						// Fetch the new category ID after insertion
						categoryID = a.model.GetCategoryByID(download.Category)
					}

					if model.FileExistsInDB(existingSounds, categoryID, download.SoundName, fileHash) {
						// File exists and hasn't changed, skip
						return
					}

					// File does not exist in the DB, add it
					if err := a.model.AddSound(categoryID, download.SoundName, fileHash, fileData); err != nil {
						//ignore this error if the sound already exists
						if !strings.Contains(err.Error(), "UNIQUE constraint failed") {
							dlog.WarningLog.Printf("Failed to add sound %s to category %s: %v", download.SoundName, download.Category, err)
						}
					}
					// Build button for the new sound
					message := "🎶  New sound added, try it out"
					component := model.BuildSingleSoundButton(soundName, category, discordgo.SuccessButton)

					err = a.UpdateInteractionResponseWithButton(message, component, s, i)
					if err != nil {
						dlog.ErrorLog.Printf("error[manage5] updating interaction response: %v", err)
					}
					return
				case "delete":
					soundName := option.Options[0].StringValue()
					response := fmt.Sprintf("👉  Deleting button: %s", soundName)
					err := a.SendInteractionRespond(response, s, i)
					if err != nil {
						dlog.ErrorLog.Println("error executing buttons command:", err)
					}

					// Check if the sound exists in the database by name
					sound, err := a.model.GetSound(soundName)
					if err != nil {
						dlog.ErrorLog.Println("error getting sound:", err)
						return
					}
					if sound.Name == "" {
						response = fmt.Sprintf("🎶  Sound not found: %s", soundName)
						err = a.UpdateInteractionResponse(response, s, i)
						if err != nil {
							dlog.ErrorLog.Println("error[manage6] sending message:", err)
						}
						return
					}

					//delete file from db
					err = a.model.DeleteSound(soundName)
					if err != nil {
						dlog.ErrorLog.Println("error deleting sound from db:", err)
						return
					}
					response = fmt.Sprintf("🎶  Sound deleted: %s", soundName)
					err = a.UpdateInteractionResponse(response, s, i)
					if err != nil {
						dlog.ErrorLog.Println("error[manage7] sending message:", err)
					}
				case "move":
					soundName := option.Options[0].StringValue()
					category := option.Options[1].StringValue()
					response := fmt.Sprintf("👉  Moving sound %s to category %s", soundName, category)
					err := a.SendInteractionRespond(response, s, i)
					if err != nil {
						dlog.ErrorLog.Println("error executing buttons command:", err)
					}

					var categoryID int
					existingCategories, _ := a.model.GetCategoriesM() // Get current folders/categories in DB

					// Check if the folder (category) exists in the database
					if dbCategoryID, exists := existingCategories[category]; exists {
						categoryID = dbCategoryID // The folder already exists
					} else {
						// The folder doesn't exist in the database, so we need to add it
						if err := a.model.AddCategory(category); err != nil {
							if !strings.Contains(err.Error(), "UNIQUE constraint failed") {
								dlog.WarningLog.Printf("Failed to add category %s: %v", category, err)
							}
							return
						}
						// Fetch the new category ID after insertion
						categoryID = a.model.GetCategoryByID(category)
					}

					err = a.model.MoveSound(categoryID, soundName)
					if err != nil {
						dlog.ErrorLog.Println("error moving sound to another category:", err)
						return
					}

					response = fmt.Sprintf("🎶  Moved sound %s to category %s", soundName, category)
					err = a.UpdateInteractionResponse(response, s, i)
					if err != nil {
						dlog.ErrorLog.Println("error executing buttons command:", err)
					}
				default:
					err := a.SendInteractionRespond("🎶  Something went wrong...", s, i)
					if err != nil {
						dlog.ErrorLog.Println("fallback to default manage handler", err)
					}
				}
			}
		}
	}
}

// DownloadAudio downloads the audio from the provided URL
func (a *API) DownloadAudio(download Download) error {
	timeout := time.Duration(a.model.Config.AudioTimeout) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	destFile := filepath.Join(a.model.Config.SoundsDir, download.Category, download.SoundName+".mp3")
	if download.Category == "" {
		destFile = filepath.Join(a.model.Config.DataDir, download.SoundName+".mp3")
	}
	dlog.InfoLog.Printf("Downloading %s to %s. START: %s, END: %s", download.URL, destFile, download.Start, download.End)
	cmdStr := fmt.Sprintf("yt-dlp -x --audio-format mp3 --force-overwrites -o %s %s", destFile, download.URL)
	if download.Start != "" && download.End != "" {
		cmdStr = fmt.Sprintf("yt-dlp -x --audio-format mp3 --download-sections \"*%s-%s\" --force-overwrites -o %s %s", download.Start, download.End, destFile, download.URL)
	}

	cmd := exec.CommandContext(ctx, "bash", "-c", cmdStr)

	// Start the command asynchronously and wait for completion
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run yt-dlp: %w, output: %s", err, string(output))
	}

	// If the context has expired or was canceled before the command completed, handle that here
	if ctx.Err() == context.DeadlineExceeded {
		dlog.ErrorLog.Println("Download Audio operation timed out")
		return ctx.Err()
	} else if ctx.Err() == context.Canceled {
		dlog.ErrorLog.Println("Download Audio operation was canceled")
		return ctx.Err()
	}

	return nil
}

func (a *API) ConvertMP3ToDCA(fileName, categoryName string) error {
	timeout := time.Duration(a.model.Config.AudioTimeout) * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	sourceFile := filepath.Join(a.model.Config.SoundsDir, categoryName, fileName+".mp3")
	destFile := filepath.Join(a.model.Config.SoundsDir, categoryName, fileName+".dca")
	if categoryName == "" {
		sourceFile = filepath.Join(a.model.Config.DataDir, fileName+".mp3")
		destFile = filepath.Join(a.model.Config.DataDir, fileName+".dca")
	}

	cmdStr := fmt.Sprintf("ffmpeg -i %s -filter:a \"loudnorm=I=-14:LRA=7:TP=-2, compand=attacks=0:points=-80/-80|-10/-5|0/-1\" -f s16le -ar 48000 -ac 2 pipe:1 | dca > %s", sourceFile, destFile)

	cmd := exec.CommandContext(ctx, "bash", "-c", cmdStr)

	// Run the command and capture the output
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to run ffmpeg/dca: %w, output: %s", err, string(output))
	}

	// Check if the context was canceled or timed out
	if ctx.Err() == context.DeadlineExceeded {
		dlog.ErrorLog.Println("Converting Audio operation timed out")
		return ctx.Err()
	} else if ctx.Err() == context.Canceled {
		dlog.ErrorLog.Println("Converting Audio operation was canceled")
		return ctx.Err()
	}

	return nil
}
