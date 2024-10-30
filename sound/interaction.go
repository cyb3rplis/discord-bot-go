package sound

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/logger"
	"github.com/cyb3rplis/discord-bot-go/model"
	"github.com/cyb3rplis/discord-bot-go/utils"
)

// InteractionHandler handles interaction events (e.g., button clicks)
func InteractionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}
	customID := i.MessageComponentData().CustomID

	// Acknowledge the interaction
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
	if err != nil {
		logger.ErrorLog.Println("Failed to respond to interaction:", err)
		return
	}

	// Check if the user is in the Gulag
	user, err := utils.GetUserFromUsername(i.Member.User.GlobalName)
	if err != nil {
		logger.ErrorLog.Println("error getting user from username:", err)
	} else {
		if remaining, ok := utils.IsUserInGulag(user); ok {
			user.Remaining = remaining
			message := fmt.Sprintf("<@"+user.ID+"> you are in the Gulag for another %s", user.Remaining)
			utils.NewMessageRoutine(".gulag"+user.ID, message, s, &discordgo.MessageCreate{Message: i.Message})
			return
		}
	}

	// Check if the user is spamming the buttons and limit the interactions
	mu.Lock()
	lastInteraction, exists := userLastInteraction[i.Member.User.ID] // Get the last interaction time
	if exists && time.Since(lastInteraction) < resetDuration {       // Check if the user has interacted recently
		userInteractionCount[i.Member.User.ID]++
	} else {
		userInteractionCount[i.Member.User.ID] = 1 // Reset the interaction count
	}
	userLastInteraction[i.Member.User.ID] = time.Now()            // Update the last interaction time
	if userInteractionCount[i.Member.User.ID] > maxInteractions { // Check if the user has exceeded the interaction limit
		mu.Unlock()
		message := "Stop spamming the buttons <@" + i.Member.User.ID + "> you fucking idiot!!!"

		utils.NewMessageRoutine(".idiot", message, s, &discordgo.MessageCreate{Message: i.Message})
		return
	}
	mu.Unlock()

	switch {
	case strings.HasPrefix(customID, "play_sound_"):
		HandlePlaySoundInteraction(s, i, customID)
	case strings.HasPrefix(customID, "list_sounds_"):
		HandleListSoundsInteraction(s, i, customID)
	case strings.HasPrefix(customID, "stop_sound"):
		handleStopSoundInteraction(s)
	default:
		logger.ErrorLog.Println("unknown interaction:", customID)
	}
}

func handleStopSoundInteraction(s *discordgo.Session) {
	// check if the bot is currently speaking, and exit
	if botSpeaking {
		stopChannel <- struct{}{}
		utils.DeleteMessageRoutine(s, ".stopbutton")
		time.Sleep(150 * time.Millisecond) // Give some time for the current sound to stop
	}
}

func HandlePlaySoundInteraction(s *discordgo.Session, i *discordgo.InteractionCreate, customID string) {
	ttsOutput := model.Bot.Config.TTSOutput
	// Extract the subfolder and sound name from the custom ID
	parts := strings.SplitN(strings.TrimPrefix(customID, "play_sound_"), "_", 2)
	if len(parts) != 2 {
		logger.ErrorLog.Println("Invalid custom ID format")
		return
	}
	//subfolder := parts[0]
	soundName := parts[1]

	// Find the channel that the interaction came from
	c, err := s.State.Channel(i.ChannelID)
	if err != nil {
		logger.ErrorLog.Println("error finding channel:", err)
		return
	}

	// Find the guild for that channel
	g, err := s.State.Guild(c.GuildID)
	if err != nil {
		logger.ErrorLog.Println("error finding guild:", err)
		return
	}

	// Look for the interaction user in that guild's current voice states
	for _, vs := range g.VoiceStates {
		if vs.UserID == i.Member.User.ID {
			if customID == "play_sound_temp_tts" {
				content := []discordgo.MessageComponent{}
				row := discordgo.ActionsRow{}
				row.Components = append(row.Components, discordgo.Button{
					Label:    "Stop Sound",
					Style:    discordgo.DangerButton,
					CustomID: "stop_sound",
				})
				content = append(content, row)
				message := &discordgo.MessageSend{
					Content:    "➡ Text2Speech playing by <@" + i.Member.User.GlobalName + ">",
					Components: content,
				}

				utils.NewComplexMessageRoutine(".stopbutton", i.ChannelID, i.ID, message, s)

				logger.InfoLog.Printf("User: %s played sound: %s", i.Member.User.GlobalName, soundName)

				// Play the sound
				err = PlaySound(s, &discordgo.MessageCreate{Message: i.Message}, g.ID, vs.ChannelID, ttsOutput)
				if err != nil {
					logger.ErrorLog.Println("error playing sound:", err)
				}
				return
			} else {
				// add user and user statistics
				userID, err := strconv.Atoi(i.Member.User.ID)
				if err != nil {
					logger.ErrorLog.Println("error converting user ID to int:", err)
				} else {
					err = utils.AddUser(userID, i.Member.User.GlobalName)
					if err != nil {
						logger.ErrorLog.Println("error adding user:", err)
					}

					err = utils.AddUserStatistics(userID, soundName)
					if err != nil {
						logger.ErrorLog.Println("error adding user statistics:", err)
					}
				}

				content := []discordgo.MessageComponent{}
				row := discordgo.ActionsRow{}
				row.Components = append(row.Components, discordgo.Button{
					Label:    "Stop Sound",
					Style:    discordgo.DangerButton,
					CustomID: "stop_sound",
				})
				content = append(content, row)

				message := &discordgo.MessageSend{
					Content:    "➡ Currently Playing by <@" + i.Member.User.ID + ">: " + soundName,
					Components: content,
				}

				utils.NewComplexMessageRoutine(".stopbutton", i.ChannelID, i.ID, message, s)

				logger.InfoLog.Printf("User: %s played sound: %s", i.Member.User.GlobalName, soundName)
				// Construct the sound file path
				// soundFile := fmt.Sprintf("%s/%s/%s.dca", model.Bot.Config.SoundsDir, subfolder, soundName)

				// Play the sound
				err = PlaySound(s, &discordgo.MessageCreate{Message: i.Message}, g.ID, vs.ChannelID, soundName)
				if err != nil {
					logger.ErrorLog.Println("error playing sound:", err)
				}

				return
			}

		}

	}

	// If the user is not in a voice channel, send an error message
	logger.InfoLog.Printf("User %s tried to play sound \"%s\" but is not in a voice channel", i.Member.User.GlobalName, soundName)
	message := "You need to be in a voice channel to play sounds <@" + i.Member.User.ID + ">"

	utils.NewMessageRoutine(".novc"+i.Member.User.ID, message, s, &discordgo.MessageCreate{Message: i.Message})
}

// HandleListSoundsInteraction handles the list sounds interaction
func HandleListSoundsInteraction(s *discordgo.Session, i *discordgo.InteractionCreate, customID string) {
	// Extract the category from the custom ID
	category := strings.TrimPrefix(customID, "list_sounds_")
	message := "➡ Sounds in category - " + category

	longCategory := fmt.Sprintf(".listAll%s", category)
	utils.NewMessageRoutine(longCategory, message, s, &discordgo.MessageCreate{Message: i.Message})

	// Get all sound files in the subfolder
	sounds, err := getSounds(category)
	if err != nil {
		logger.ErrorLog.Println("error listing sounds in subfolder:", err)
	}
	if len(sounds) == 0 {
		message := "No sounds found in this category."
		utils.NewMessageRoutine(".list"+"no"+category, message, s, &discordgo.MessageCreate{Message: i.Message})
		return
	}

	//build buttons for each sound
	buttons := utils.BuildSoundButtons(sounds, category, discordgo.SecondaryButton)
	//build messages
	messages := utils.BuildMessages(buttons, nil)

	logger.InfoLog.Printf("User: %s listed sounds in category: %s", i.Member.User.GlobalName, category)
	for idx, message := range messages {
		longCategory := fmt.Sprintf(".list%s%d", category, idx+1)
		utils.NewComplexMessageRoutine(longCategory, i.ChannelID, i.ID, message, s)
	}
}
