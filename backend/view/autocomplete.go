package view

import (
	"github.com/bwmarrin/discordgo"
	log "github.com/cyb3rplis/discord-bot-go/logger"
)

// AutocompleteHandler handles autocomplete interactions for sound name fields.
func (a *API) AutocompleteHandler(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommandAutocomplete {
		return
	}

	// Extract the focused option's current value
	query := ""
	data := i.ApplicationCommandData()

	// Walk through options to find the focused one (handles both top-level and subcommand options)
	options := data.Options
	for _, opt := range options {
		if opt.Type == discordgo.ApplicationCommandOptionSubCommand || opt.Type == discordgo.ApplicationCommandOptionSubCommandGroup {
			options = opt.Options
			break
		}
	}
	for _, opt := range options {
		if opt.Focused {
			query = opt.StringValue()
			break
		}
	}

	sounds, err := a.model.SearchSounds(query)
	if err != nil {
		log.ErrorLog.Printf("autocomplete search error: %v", err)
		sounds = []string{}
	}

	choices := make([]*discordgo.ApplicationCommandOptionChoice, 0, len(sounds))
	for _, sound := range sounds {
		choices = append(choices, &discordgo.ApplicationCommandOptionChoice{
			Name:  sound,
			Value: sound,
		})
	}

	err = s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionApplicationCommandAutocompleteResult,
		Data: &discordgo.InteractionResponseData{
			Choices: choices,
		},
	})
	if err != nil {
		log.ErrorLog.Printf("failed to send autocomplete response: %v", err)
	}
}
