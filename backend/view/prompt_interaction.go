package view

import (
	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/dlog"
	"github.com/cyb3rplis/discord-bot-go/model"
)

// RegisterPromptInteractionsAudio - Register prompt interactions
func RegisterPromptInteractionsAudio(s *discordgo.Session) {
	commandName := "audio"
	// Register the command with subcommands for URL and last
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        commandName,
			Description: "Play audio from a URL or replay the last audio",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "play",
					Description: "Play audio from a URL (e.g., YouTube, Vimeo, etc.)",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "url",
							Description: "The URL of the video you want to play",
							Required:    true,
						},
					},
				},
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "last",
					Description: "Replay the last played audio",
				},
			},
		},
	}

	//deletePromptInteraction(s, commandName)

	// Now register the new command
	ccmd, err := s.ApplicationCommandCreate(s.State.User.ID, model.Meta.Guild.ID, commands[0])
	if err != nil {
		dlog.FatalLog.Fatalf("failed to create '%s' command: %v", commandName, err)
	}

	dlog.InfoLog.Printf("'%s' command registered - ID: %s", commandName, ccmd.ID)
}

// RegisterPromptInteractionsButtons - Register prompt interactions
func RegisterPromptInteractionsButtons(s *discordgo.Session) {
	commandName := "buttons"
	// Register the command globally
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        commandName,
			Description: "Manage your favorite sounds",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "list",
					Description: "List all your favorite sound buttons",
				},
			},
		},
	}

	//deletePromptInteraction(s, commandName)

	// Now register the new command
	ccmd, err := s.ApplicationCommandCreate(s.State.User.ID, model.Meta.Guild.ID, commands[0])
	if err != nil {
		dlog.FatalLog.Fatalf("failed to create '%s' command: %v", commandName, err)
	}

	dlog.InfoLog.Printf("'%s' command registered - ID: %s", commandName, ccmd.ID)
}

// RegisterPromptInteractionsCreate - Register prompt interactions
func RegisterPromptInteractionsCreate(s *discordgo.Session) {
	commandName := "create"
	// Register the command globally
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        commandName,
			Description: "Manage your favorite sounds",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "button",
					Description: "Create a new sound button using a URL (e.G. YouTube, Vimeo, etc.)",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "url",
							Description: "The URL of the video you want to create a sound button for",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "name",
							Description: "The name of the sound button",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "category",
							Description: "The category of the sound button",
							Required:    true,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "start_time",
							Description: "The start time of the sound button",
							Required:    false,
						},
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "end_time",
							Description: "The end time of the sound button",
							Required:    false,
						},
					},
				}},
		},
	}
	// Now register the new command
	ccmd, err := s.ApplicationCommandCreate(s.State.User.ID, model.Meta.Guild.ID, commands[0])
	if err != nil {
		dlog.FatalLog.Fatalf("failed to create '%s' command: %v", commandName, err)
	}

	dlog.InfoLog.Printf("'%s' command registered - ID: %s", commandName, ccmd.ID)
}

// RegisterPromptInteractionsDelete - Register prompt interactions
func RegisterPromptInteractionsDelete(s *discordgo.Session) {
	commandName := "delete"
	// Register the command globally
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        commandName,
			Description: "Manage your favorite sounds",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionSubCommand,
					Name:        "button",
					Description: "Delete a sound button by name",
					Options: []*discordgo.ApplicationCommandOption{
						{
							Type:        discordgo.ApplicationCommandOptionString,
							Name:        "name",
							Description: "The name of the sound button to delete",
							Required:    true,
						},
					},
				},
			},
		},
	}
	// Now register the new command
	ccmd, err := s.ApplicationCommandCreate(s.State.User.ID, model.Meta.Guild.ID, commands[0])
	if err != nil {
		dlog.FatalLog.Fatalf("failed to create '%s' command: %v", commandName, err)
	}

	dlog.InfoLog.Printf("'%s' command registered - ID: %s", commandName, ccmd.ID)
}

// RegisterPromptInteractionsFavorite - Register prompt interactions for favorite
func (a *API) RegisterPromptInteractionsFavorite(s *discordgo.Session) {
	commandName := "favorite"
	/*sounds, err := a.model.GetSoundsAll()
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
	}*/
	// Register the command globally
	commands := []*discordgo.ApplicationCommand{
		{
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
		},
	}

	// deletePromptInteraction(s, commandName)

	// Now register the new command
	ccmd, err := s.ApplicationCommandCreate(s.State.User.ID, model.Meta.Guild.ID, commands[0])
	if err != nil {
		dlog.FatalLog.Fatalf("failed to create '%s' command: %v", commandName, err)
	}

	dlog.InfoLog.Printf("'%s' command registered - ID: %s", commandName, ccmd.ID)
}

// RegisterPromptInteractionsGulag - Register prompt interactions for gulag
func (a *API) RegisterPromptInteractionsGulag(s *discordgo.Session) {
	commandName := "gulag"
	/*users, err := a.model.GetUsers()
	if err != nil {
		dlog.FatalLog.Fatalf("cannot get sounds: %v", err)
		return
	}
	var usersChoices []*discordgo.ApplicationCommandOptionChoice
	for _, user := range users {
		soundChoice := &discordgo.ApplicationCommandOptionChoice{
			Name:  user.User.GlobalName,
			Value: user.User.GlobalName,
		}
		usersChoices = append(usersChoices, soundChoice)
	}*/
	// Register the command globally
	commands := []*discordgo.ApplicationCommand{
		{
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
						},
					},
				},
			},
		},
	}

	// deletePromptInteraction(s, commandName)

	// Now register the new command
	ccmd, err := s.ApplicationCommandCreate(s.State.User.ID, model.Meta.Guild.ID, commands[0])
	if err != nil {
		dlog.FatalLog.Fatalf("failed to create '%s' command: %v", commandName, err)
	}

	dlog.InfoLog.Printf("'%s' command registered - ID: %s", commandName, ccmd.ID)
}

// RegisterPromptInteractionsStats - Register prompt interactions for stats
func RegisterPromptInteractionsStats(s *discordgo.Session) {
	commandName := "stats"
	// Register the command globally
	commands := []*discordgo.ApplicationCommand{
		{
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
		},
	}

	// deletePromptInteraction(s, commandName)

	// Now register the new command
	ccmd, err := s.ApplicationCommandCreate(s.State.User.ID, model.Meta.Guild.ID, commands[0])
	if err != nil {
		dlog.FatalLog.Fatalf("failed to create '%s' command: %v", commandName, err)
	}

	dlog.InfoLog.Printf("'%s' command registered - ID: %s", commandName, ccmd.ID)
}

func (a *API) RegisterPromptInteractionsPlaySound(s *discordgo.Session) {
	commandName := "play"
	// Register the command globally
	/*sounds, err := a.model.GetSoundsAll()
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
	}*/
	commands := []*discordgo.ApplicationCommand{
		{
			Name:        commandName,
			Description: "Play a sound by name",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "sound",
					Description: "The name of the sound to play",
					Required:    true,
				},
			},
		},
	}

	// deletePromptInteraction(s, commandName)

	// Now register the new command
	ccmd, err := s.ApplicationCommandCreate(s.State.User.ID, model.Meta.Guild.ID, commands[0])
	if err != nil {
		dlog.FatalLog.Fatalf("failed to create '%s' command: %v", commandName, err)
	}

	dlog.InfoLog.Printf("'%s' command registered - ID: %s", commandName, ccmd.ID)
}

func deletePromptInteraction(s *discordgo.Session, commandName string) {

	// Fetch all the global commands
	commandsList, err := s.ApplicationCommands(s.State.User.ID, model.Meta.Guild.ID)
	if err != nil {
		dlog.FatalLog.Fatalf("cannot fetch commands: %v", err)
	}
	// Check if the "audio" command already exists
	var existingCmd *discordgo.ApplicationCommand
	for _, cmd := range commandsList {
		if cmd.Name == commandName {
			existingCmd = cmd
			break
		}
	}

	// If the command exists, delete it before creating a new one
	if existingCmd != nil {
		err := s.ApplicationCommandDelete(s.State.User.ID, model.Meta.Guild.ID, existingCmd.ID)
		if err != nil {
			dlog.FatalLog.Printf("failed to delete existing command %s: %v", commandName, err)
		}
	}
}
