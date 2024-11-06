package view

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/cyb3rplis/discord-bot-go/model"

	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/dlog"
)

var BotReady = false

// InteractionHandler handles interaction events (e.g., button clicks)
func (a *API) InteractionHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionMessageComponent {
		return
	}
	customID := i.MessageComponentData().CustomID

	// Acknowledge the interaction
	err := s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredMessageUpdate,
	})
	if err != nil {
		dlog.ErrorLog.Println("Failed to respond to interaction:", err)
		return
	}

	// Check if the user is in the Gulag
	user, err := a.model.GetUserFromUsername(i.Member.User.GlobalName)
	if err != nil {
		dlog.ErrorLog.Println("error getting user from username:", err)
		return
	} else {
		if remaining, ok := IsUserInGulag(user); ok {
			user.Remaining = remaining
			msg := fmt.Sprintf("<@"+user.ID+"> you are in the Gulag for another %s", user.Remaining)
			_, err = a.SendMessage(msg, s, i, false)
			if err != nil {
				dlog.ErrorLog.Printf("error sending message: %v", err)
			}
			return
		}
	}

	// Check if the user is spamming the buttons and limit the interactions
	mu.Lock()
	lastInteraction, exists := userLastInteraction[user.ID]    // Get the last interaction time
	if exists && time.Since(lastInteraction) < resetDuration { // Check if the user has interacted recently
		userInteractionCount[user.ID]++
	} else {
		userInteractionCount[user.ID] = 1 // Reset the interaction count
	}
	userLastInteraction[user.ID] = time.Now()            // Update the last interaction time
	if userInteractionCount[user.ID] > maxInteractions { // Check if the user has exceeded the interaction limit
		mu.Unlock()
		msg := "Stop spamming the buttons <@" + user.ID + ">, you are now being sent to the Gulag for one minute."
		_, err = a.SendMessage(msg, s, i, true)
		if err != nil {
			dlog.ErrorLog.Println("error sending message:", err)
		}
		err := a.model.GulagUser(user.Username, 1)
		if err != nil {
			dlog.ErrorLog.Println("error gulagging user:", err)
		}

		return
	}
	mu.Unlock()

	switch {
	case strings.HasPrefix(customID, "play_sound_"):
		a.handlePlaySoundInteraction(s, i)
	case strings.HasPrefix(customID, "list_sounds_"):
		a.handleListSoundsInteraction(s, i)
	case strings.HasPrefix(customID, "stop_sound"):
		a.handleStopSoundInteraction(s, i)
	default:
		dlog.ErrorLog.Println("unknown interaction:", customID)
	}
}

// handleStopSoundInteraction handles the stop sound interaction
func (a *API) handleStopSoundInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Check if the bot is currently speaking, and exit
	if botSpeaking {
		stopChannel <- struct{}{}
		time.Sleep(150 * time.Millisecond) // Give some time for the current sound to stop
	}

	// Delete the message that contains the stop button
	err := s.ChannelMessageDelete(i.ChannelID, i.Message.ID)
	if err != nil {
		dlog.ErrorLog.Println("error deleting message:", err)
	}
}

func (a *API) handlePlaySoundInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Extract the subfolder and sound name from the custom ID
	customID := ""
	if i.Interaction.Type != 3 {
		dlog.ErrorLog.Println("Invalid interaction type")
		return
	}

	customID = i.Interaction.MessageComponentData().CustomID

	parts := strings.SplitN(strings.TrimPrefix(customID, "play_sound_"), "_", 2)
	if len(parts) != 2 {
		dlog.ErrorLog.Println("Invalid custom ID format")
		return
	}
	//subfolder := parts[0]
	soundName := parts[1]

	// Find the channel that the interaction came from
	c, err := s.State.Channel(i.ChannelID)
	if err != nil {
		dlog.ErrorLog.Println("error finding channel:", err)
		return
	}

	// Find the guild for that channel
	g, err := s.State.Guild(c.GuildID)
	if err != nil {
		dlog.ErrorLog.Println("error finding guild:", err)
		return
	}

	// Look for the interaction user in that guild's current voice states
	for _, vs := range g.VoiceStates {
		if vs.UserID == i.Member.User.ID {
			// add user and user statistics
			userID, err := strconv.Atoi(i.Member.User.ID)
			if err != nil {
				dlog.ErrorLog.Println("error converting user ID to int:", err)
			} else {
				err = a.model.AddUser(userID, i.Member.User.GlobalName)
				if err != nil {
					dlog.ErrorLog.Println("error adding user:", err)
				}

				err = a.model.AddUserStatistics(userID, soundName)
				if err != nil {
					dlog.ErrorLog.Println("error adding user statistics:", err)
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

			msg := &discordgo.MessageSend{
				Content:    "➡ Currently Playing by <@" + i.Member.User.ID + ">: " + soundName,
				Components: content,
			}

			// Send the message (+stop button)
			_, err = a.SendMessageComplex(msg, s, i, false)
			if err != nil {
				dlog.ErrorLog.Println("error sending message:", err)
				return
			}

			dlog.InfoLog.Printf("User: %s played sound: %s", i.Member.User.GlobalName, soundName)
			// Construct the sound file path
			// soundFile := fmt.Sprintf("%s/%s/%s.dca", model.Bot.Config.SoundsDir, subfolder, soundName)

			// Play the sound
			err = a.PlaySound(s, i, g.ID, vs.ChannelID, soundName)
			if err != nil {
				dlog.ErrorLog.Println("error playing sound:", err)
			}

			return

		}

	}

	// If the user is not in a voice channel, send an error message
	dlog.InfoLog.Printf("User %s tried to play sound \"%s\" but is not in a voice channel", i.Member.User.GlobalName, soundName)
	msg := "You need to be in a voice channel to play sounds <@" + i.Member.User.ID + ">"

	_, err = a.SendMessage(msg, s, i, false)
	if err != nil {
		dlog.ErrorLog.Println("error sending message:", err)
	}
}

// HandleListSoundsInteraction handles the list sounds interaction
func (a *API) handleListSoundsInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	// Extract the category from the custom ID
	customID := ""
	if i.Interaction.Type != 3 {
		dlog.ErrorLog.Println("Invalid interaction type")
		return
	}

	customID = i.Interaction.MessageComponentData().CustomID

	category := strings.TrimPrefix(customID, "list_sounds_")
	msg := "➡ Sounds in category - " + category
	_, err := a.SendMessage(msg, s, i, false)
	if err != nil {
		dlog.ErrorLog.Println("error sending message:", err)
	}

	// Get all sound files in the subfolder
	sounds, err := a.model.GetSounds(category)
	if err != nil {
		dlog.ErrorLog.Println("error listing sounds in subfolder:", err)
	}
	if len(sounds) == 0 {
		msg := "No sounds found in this category."
		_, err = a.SendMessage(msg, s, i, false)
		if err != nil {
			dlog.ErrorLog.Println("error sending message:", err)
		}
		return
	}

	//build buttons for each sound
	buttons := model.BuildSoundButtons(sounds, category, discordgo.SecondaryButton)
	//build messages
	messages := model.BuildMessages(buttons, nil)

	dlog.InfoLog.Printf("User: %s listed sounds in category: %s", i.Member.User.GlobalName, category)
	for _, msg := range messages {
		_, err = a.SendMessageComplex(msg, s, i, false)
		if err != nil {
			dlog.ErrorLog.Println("error sending message:", err)
		}
	}
}
