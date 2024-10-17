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
		_, err := s.ChannelMessageSend(i.ChannelID, "Stop spamming the buttons <@"+i.Member.User.ID+"> you fucking idiot!!!")
		if err != nil {
			logger.ErrorLog.Println("Error sending message:", err)
		}
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

		// Delete the last "Now Playing" message
		// This should not be needed, since the actual function will finish normally when emptying the buffer
		// and this deletes the old message anyway. Keeping it here just to make sure
		_ = s.ChannelMessageDelete(lastChannelID, lastMessageID)
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
	subfolder := parts[0]
	soundName := parts[1]

	// Find the channel that the interaction came from
	c, err := s.State.Channel(i.ChannelID)
	if err != nil {
		logger.ErrorLog.Println("Error finding channel:", err)
		return
	}

	// Find the guild for that channel
	g, err := s.State.Guild(c.GuildID)
	if err != nil {
		logger.ErrorLog.Println("Error finding guild:", err)
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
				st, err := s.ChannelMessageSendComplex(i.ChannelID, &discordgo.MessageSend{
					Content:    "➡ Text2Speech playing by <@" + i.Member.User.GlobalName + ">",
					Components: content,
				})
				if err != nil {
					logger.ErrorLog.Println("Error sending message:", err)
				}
				logger.InfoLog.Printf("User: %s played sound: %s", i.Member.User.GlobalName, soundName)

				// Play the sound
				err = PlaySound(s, &discordgo.MessageCreate{Message: i.Message}, st, g.ID, vs.ChannelID, ttsOutput, "tts")
				if err != nil {
					logger.ErrorLog.Println("Error playing sound:", err)
				}
				_ = s.ChannelMessageDelete(st.ChannelID, st.ID)
				return
			} else {
				// add user and user statistics
				userID, err := strconv.Atoi(i.Member.User.ID)
				if err != nil {
					logger.ErrorLog.Println("Error converting user ID to int:", err)
				} else {
					err = utils.AddUser(userID, i.Member.User.GlobalName)
					if err != nil {
						logger.ErrorLog.Println("Error adding user:", err)
					}

					err = utils.AddUserStatistics(userID, soundName)
					if err != nil {
						logger.ErrorLog.Println("Error adding user statistics:", err)
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
				st, err := s.ChannelMessageSendComplex(i.ChannelID, &discordgo.MessageSend{
					Content:    "➡ Currently Playing by <@" + i.Member.User.ID + ">: " + soundName,
					Components: content,
				})
				if err != nil {
					logger.ErrorLog.Println("Error sending message:", err)
				}
				logger.InfoLog.Printf("User: %s played sound: %s", i.Member.User.GlobalName, soundName)
				// Construct the sound file path
				soundFile := fmt.Sprintf("%s/%s/%s.dca", model.Bot.Config.SoundsDir, subfolder, soundName)

				// Play the sound
				err = PlaySound(s, &discordgo.MessageCreate{Message: i.Message}, st, g.ID, vs.ChannelID, soundFile, soundName)
				if err != nil {
					logger.ErrorLog.Println("Error playing sound:", err)
				}
				_ = s.ChannelMessageDelete(st.ChannelID, st.ID)

				return
			}

		}

	}

	// If the user is not in a voice channel, send an error message
	logger.InfoLog.Printf("User %s tried to play sound \"%s\" but is not in a voice channel", i.Member.User.GlobalName, soundName)
	_, err = s.ChannelMessageSend(i.ChannelID, "You need to be in a voice channel to play sounds <@"+i.Member.User.ID+">")
	if err != nil {
		logger.ErrorLog.Println("Error sending message:", err)
	}

}

func HandleListSoundsInteraction(s *discordgo.Session, i *discordgo.InteractionCreate, customID string) {
	// Extract the category from the custom ID
	category := strings.TrimPrefix(customID, "list_sounds_")
	_, err := s.ChannelMessageSend(i.ChannelID, "➡ Sounds in category - "+category)
	if err != nil {
		logger.ErrorLog.Println("Error sending message:", err)
	}

	// Get all sound files in the subfolder
	sounds, err := getSounds(category)
	if err != nil {
		logger.ErrorLog.Println("Error listing sounds in subfolder:", err)
	}
	if len(sounds) == 0 {
		_, err := s.ChannelMessageSend(i.ChannelID, "No sounds found in this category.")
		if err != nil {
			logger.ErrorLog.Println("Error sending message:", err)
		}
		return
	}

	buttons, err := BuildSoundButtons(sounds, category)
	if err != nil {
		logger.ErrorLog.Println("Error listing sounds in category:", err)
		return
	}

	logger.InfoLog.Printf("User: %s listed sounds in category: %s", i.Member.User.GlobalName, category)

	// Split content into multiple messages if it exceeds 5 rows
	for len(buttons) > 0 {
		var messageContent []discordgo.MessageComponent
		if len(buttons) > 5 {
			messageContent, buttons = buttons[:5], buttons[5:]
		} else {
			messageContent, buttons = buttons, nil
		}
		_, err = s.ChannelMessageSendComplex(i.ChannelID, &discordgo.MessageSend{
			Components: messageContent,
		})
		if err != nil {
			logger.ErrorLog.Println("Error sending message:", err)
		}
	}
}
