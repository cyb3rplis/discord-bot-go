package view

import (
	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/dlog"
)

// RegisterPromptInteractionsAudio - Register prompt interactions
func RegisterPromptInteractionsAudio(s *discordgo.Session, i *discordgo.InteractionCreate) {
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
		dlog.FatalLog.Fatalf("cannot create '%s' command: %v", commandName, err)
	}
}

// RegisterPromptInteractionsButtons - Register prompt interactions
func RegisterPromptInteractionsButtons(s *discordgo.Session, i *discordgo.InteractionCreate) {
	commandName := "buttons"
	// Register the command globally
	_, err := s.ApplicationCommandCreate(s.State.User.ID, "", &discordgo.ApplicationCommand{
		Name:        commandName,
		Description: "Use buttons to list all button categories",
	})
	if err != nil {
		dlog.FatalLog.Fatalf("cannot create '%s' command: %v", commandName, err)
	}
}

// RegisterPromptInteractionsFavorite - Register prompt interactions for favorite
func (a *API) RegisterPromptInteractionsFavorite(s *discordgo.Session, i *discordgo.InteractionCreate) {
	commandName := "favorite"
	userID := i.Member.User.ID
	sounds, err := a.model.GetSoundsAll()
	if err != nil {
		dlog.FatalLog.Fatalf("cannot get sounds: %v", err)
		return
	}
	var soundsChoices []*discordgo.ApplicationCommandOptionChoice
	for _, sound := range sounds {
		soundChoice := &discordgo.ApplicationCommandOptionChoice{
			Name:  sound,
			Value: sound,
		}
		soundsChoices = append(soundsChoices, soundChoice)
	}
	favoriteSounds, err := a.model.GetUserFavorites(userID)
	if err != nil {
		dlog.FatalLog.Fatalf("cannot get sounds: %v", err)
		return
	}
	var favoriteSoundsChoices []*discordgo.ApplicationCommandOptionChoice
	for _, sound := range favoriteSounds {
		soundChoice := &discordgo.ApplicationCommandOptionChoice{
			Name:  sound.SoundName,
			Value: sound.SoundName,
		}
		favoriteSoundsChoices = append(favoriteSoundsChoices, soundChoice)
	}
	// Register the command globally
	_, err = s.ApplicationCommandCreate(s.State.User.ID, "", &discordgo.ApplicationCommand{
		Name:        commandName,
		Description: "Manage your favorite sounds",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "buttons",
				Description: "List all your favorite sound buttons",
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
						Choices:     soundsChoices,
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
						Choices:     favoriteSoundsChoices,
					},
				},
			},
		},
	})
	if err != nil {
		dlog.FatalLog.Fatalf("cannot create '%s' command: %v", commandName, err)
	}
}

// RegisterPromptInteractionsGulag - Register prompt interactions for gulag
func (a *API) RegisterPromptInteractionsGulag(s *discordgo.Session, i *discordgo.InteractionCreate) {
	commandName := "gulag"
	users, err := a.model.GetUsers()
	if err != nil {
		dlog.FatalLog.Fatalf("cannot get sounds: %v", err)
		return
	}
	var usersChoices []*discordgo.ApplicationCommandOptionChoice
	for _, user := range users {
		soundChoice := &discordgo.ApplicationCommandOptionChoice{
			Name:  user.Username,
			Value: user.Username,
		}
		usersChoices = append(usersChoices, soundChoice)
	}
	// Register the command globally
	_, err = s.ApplicationCommandCreate(s.State.User.ID, "", &discordgo.ApplicationCommand{
		Name:        commandName,
		Description: "Manage gulag",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "list",
				Description: "List all users in the gulag",
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "add",
				Description: "Add a user to the gulag",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "user",
						Description: "The name of the user to add to the gulag",
						Required:    true,
						Choices:     usersChoices,
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
				Description: "Remove a user from the gulag",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Name:        "user",
						Description: " The name of the user to remove from the gulag",
						Required:    true,
						Choices:     usersChoices,
					},
				},
			},
		},
	})
	if err != nil {
		dlog.FatalLog.Fatalf("cannot create '%s' command: %v", commandName, err)
	}
}

// RegisterPromptInteractionsStats - Register prompt interactions for stats
func RegisterPromptInteractionsStats(s *discordgo.Session, i *discordgo.InteractionCreate) {
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
		dlog.FatalLog.Fatalf("cannot create '%s' command: %v", commandName, err)
	}
}

func (a *API) RegisterPromptInteractionsPlaySound(s *discordgo.Session, i *discordgo.InteractionCreate) {
	commandName := "play"
	// Register the command globally
	sounds, err := a.model.GetSoundsAll()
	if err != nil {
		dlog.FatalLog.Fatalf("cannot get sounds: %v", err)
		return
	}
	var soundsChoices []*discordgo.ApplicationCommandOptionChoice
	for _, sound := range sounds {
		soundChoice := &discordgo.ApplicationCommandOptionChoice{
			Name:  sound,
			Value: sound,
		}
		soundsChoices = append(soundsChoices, soundChoice)
	}
	_, err = s.ApplicationCommandCreate(s.State.User.ID, "", &discordgo.ApplicationCommand{
		Name:        commandName,
		Description: "Play a sound by name",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "sound",
				Description: "The name of the sound to play",
				Required:    true,
				Choices:     soundsChoices,
			},
		},
	})
	if err != nil {
		dlog.FatalLog.Fatalf("cannot create '%s' command: %v", commandName, err)
	}
}
