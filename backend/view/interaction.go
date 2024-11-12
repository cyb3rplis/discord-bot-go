package view

import (
	"database/sql"
	"fmt"
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
	user, err := a.model.SetUserGulaggedValue(i.Member.User)
	if err != nil && err != sql.ErrNoRows {
		dlog.ErrorLog.Println("error getting user from username:", err)
		return
	} else {
		if user, ok := SetUserGulagRemaining(user); ok {
			msg := fmt.Sprintf(user.User.Mention()+" you are in the Gulag for another %s", user.Remaining)
			_, err = a.SendMessage(msg, s, i, false)
			if err != nil {
				dlog.ErrorLog.Printf("error sending message: %v", err)
			}
			return
		}
	}

	// Check if the user is spamming the buttons and limit the interactions
	mu.Lock()
	lastInteraction, exists := userLastInteraction[user.User.ID] // Get the last interaction time
	if exists && time.Since(lastInteraction) < resetDuration {   // Check if the user has interacted recently
		userInteractionCount[user.User.ID]++
	} else {
		userInteractionCount[user.User.ID] = 1 // Reset the interaction count
	}
	userLastInteraction[user.User.ID] = time.Now()            // Update the last interaction time
	if userInteractionCount[user.User.ID] > maxInteractions { // Check if the user has exceeded the interaction limit
		mu.Unlock()
		msg := "Stop spamming the buttons " + user.User.Mention() + ", you are now being sent to the Gulag for one minute."
		_, err = a.SendMessage(msg, s, i, true)
		if err != nil {
			dlog.ErrorLog.Println("error sending message:", err)
		}
		err := a.model.GulagUser(user, 1)
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
	st, err := s.ChannelMessage(i.ChannelID, i.ID)
	if err == nil {
		err = s.ChannelMessageDelete(st.ChannelID, st.ID)
		if err != nil {
			dlog.ErrorLog.Println("error deleting stop sound button after sound finished:", err)
		}
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

	// Find the guild for that channel
	guild, err := s.State.Guild(model.Meta.Guild.ID)
	if err != nil {
		dlog.ErrorLog.Println("error finding guild:", err)
		return
	}

	interactionUser := i.Member.User

	// Check if the user is in a voice channel
	vs, err := a.VoiceChannelCheck(s, i)
	if err != nil {
		dlog.ErrorLog.Println("error checking voice channel:", err)
		return
	}

	// add user and user statistics
	err = a.model.AddUser(interactionUser)
	if err != nil {
		dlog.ErrorLog.Println("error adding user:", err)
	}

	err = a.model.AddUserStatistics(interactionUser, soundName)
	if err != nil {
		dlog.ErrorLog.Println("error adding user statistics:", err)
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
		Content:    "➡ Currently Playing by " + interactionUser.Mention() + " " + soundName,
		Components: content,
	}

	// Send the message (+stop button)
	st, err := a.SendMessageComplex(msg, s, i, false)
	if err != nil {
		dlog.ErrorLog.Println("error sending message:", err)
		return
	}

	err = a.DeleteOldStopSoundButtons(s, st)
	if err != nil {
		dlog.ErrorLog.Println("error deleting all stop sound buttons:", err)
	}

	dlog.InfoLog.Printf("User: %s played sound: %s", interactionUser.GlobalName, soundName)
	// Construct the sound file path
	// soundFile := fmt.Sprintf("%s/%s/%s.dca", model.Bot.Config.SoundsDir, subfolder, soundName)

	// Play the sound
	err = a.PlaySound(s, i, guild.ID, vs.ChannelID, soundName)
	if err != nil {
		dlog.ErrorLog.Println("error playing sound:", err)
	}

	time.Sleep(250 * time.Millisecond)
	st, err = s.ChannelMessage(st.ChannelID, st.ID)
	if err == nil {
		err = s.ChannelMessageDelete(st.ChannelID, st.ID)
		if err != nil {
			dlog.ErrorLog.Println("error deleting stop sound button after sound finished:", err)
		}
	}
}

// HandleListSoundsInteraction handles the list sounds interaction
func (a *API) handleListSoundsInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	interactionUser := i.Member.User
	dlog.InfoLog.Println("Handling list sounds interaction")
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

	dlog.InfoLog.Printf("User: %s listed sounds in category: %s", interactionUser.GlobalName, category)
	for _, msg := range messages {
		_, err = a.SendMessageComplex(msg, s, i, false)
		if err != nil {
			dlog.ErrorLog.Println("error sending message:", err)
		}
	}
}
