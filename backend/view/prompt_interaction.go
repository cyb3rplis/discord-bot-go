package view

import (
	"github.com/bwmarrin/discordgo"
	"log"
)

// RegisterPromptInteractionsAudio - Register prompt interactions
func RegisterPromptInteractionsAudio(s *discordgo.Session, event *discordgo.Ready) {
	commandName := "audio"
	// Register the command globally
	_, err := s.ApplicationCommandCreate(s.State.User.ID, "", &discordgo.ApplicationCommand{
		Name:        commandName,
		Description: "Use audio to play sound from a URL (e.G. YouTube, Vimeo, etc.)",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "url",
				Description: "The URL of the video you want to play",
				Required:    true,
			},
		},
	})
	if err != nil {
		log.Fatalf("cannot create '%s' command: %v", commandName, err)
	}
	log.Printf("> Prompt command registered: %s by %s\n", commandName, event.User.Username)
}

// RegisterPromptInteractionsList - Register prompt interactions
func RegisterPromptInteractionsList(s *discordgo.Session, event *discordgo.Ready) {
	commandName := "list"
	// Register the command globally
	_, err := s.ApplicationCommandCreate(s.State.User.ID, "", &discordgo.ApplicationCommand{
		Name:        commandName,
		Description: "Use list to list all button categories",
	})
	if err != nil {
		log.Fatalf("cannot create '%s' command: %v", commandName, err)
	}
	log.Printf("> Prompt command registered: %s by %s\n", commandName, event.User.Username)
}

// RegisterPromptInteractionsFavorite - Register prompt interactions for favorite
func RegisterPromptInteractionsFavorite(s *discordgo.Session, event *discordgo.Ready) {
	commandName := "favorite"
	// Register the command globally
	_, err := s.ApplicationCommandCreate(s.State.User.ID, "", &discordgo.ApplicationCommand{
		Name:        commandName,
		Description: "Manage your favorite sounds",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "list",
				Description: "List all your favorite sounds",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "add",
				Description: "Add a sound to your favorites",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "sound",
						Description: "The name of the sound to add",
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "remove",
				Description: "Remove a sound from your favorites",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "sound",
						Description: "The name of the sound to remove",
						Required:    true,
					},
				},
			},
		},
	})
	if err != nil {
		log.Fatalf("cannot create '%s' command: %v", commandName, err)
	}
	log.Printf("> Prompt command registered: %s by %s\n", commandName, event.User.Username)
}

// RegisterPromptInteractionsGulag - Register prompt interactions for gulag
func RegisterPromptInteractionsGulag(s *discordgo.Session, event *discordgo.Ready) {
	commandName := "gulag"
	// Register the command globally
	_, err := s.ApplicationCommandCreate(s.State.User.ID, "", &discordgo.ApplicationCommand{
		Name:        commandName,
		Description: "Manage gulag sounds",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "list",
				Description: "List all gulag sounds",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "add",
				Description: "Add a sound to the gulag",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "user",
						Description: "The name of the sound to add",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "timeout",
						Description: "Time in minutes to keep the user in the gulag",
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "remove",
				Description: "Remove a sound from the gulag",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "sound",
						Description: "The name of the sound to remove",
						Required:    true,
					},
				},
			},
		},
	})
	if err != nil {
		log.Fatalf("cannot create '%s' command: %v", commandName, err)
	}
	log.Printf("> Prompt command registered: %s by %s\n", commandName, event.User.Username)
}

// RegisterPromptInteractionsStats - Register prompt interactions for stats
func RegisterPromptInteractionsStats(s *discordgo.Session, event *discordgo.Ready) {
	commandName := "stats"
	// Register the command globally
	_, err := s.ApplicationCommandCreate(s.State.User.ID, "", &discordgo.ApplicationCommand{
		Name:        commandName,
		Description: "Get statistics for sounds and users",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "sounds",
				Description: "Get statistics for top sounds",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "users",
				Description: "Get statistics for top users",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "me",
				Description: "Get your personal sound statistics",
			},
		},
	})
	if err != nil {
		log.Fatalf("cannot create '%s' command: %v", commandName, err)
	}
	log.Printf("> Prompt command registered: %s by %s\n", commandName, event.User.Username)
}
