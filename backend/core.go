package backend

import (
	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/config"
	"github.com/cyb3rplis/discord-bot-go/controller"
	"github.com/cyb3rplis/discord-bot-go/db"
	"github.com/cyb3rplis/discord-bot-go/logger"
	"github.com/cyb3rplis/discord-bot-go/model"
	"github.com/cyb3rplis/discord-bot-go/view"
	"os"
	"os/signal"
	"syscall"
)

func Init() {
	m, dbClose, err := db.InitModel()
	if err != nil {
		logger.FatalLog.Fatalf("Failed to initialize database: %v", err)
	}
	defer func() {
		if err != nil {
			if err := dbClose(); err != nil {
				logger.FatalLog.Fatalf("Failed to close database: %v", err)
			}
		}
	}()
	modelInstance := model.New(&m)
	viewInstance := view.New(modelInstance)
	ctrl := controller.New(modelInstance, viewInstance)

	// Check if the sound directory exists
	if _, err := os.Stat(modelInstance.Config.SoundsDir); os.IsNotExist(err) {
		if err != nil {
			logger.FatalLog.Fatalf("Failed to check sound directory: %v", err)
			return
		}
		err = os.Mkdir(m.Config.SoundsDir, 0755)
		if err != nil {
			logger.FatalLog.Fatalf("Failed to create sound directory: %v", err)
			return
		}
	}

	fsSounds, err := modelInstance.ScanDirectory()
	if err != nil {
		logger.FatalLog.Printf("cron: error scanning sound directory: %v", err)
	}
	err = viewInstance.SyncDatabaseWithFileSystem(fsSounds)
	if err != nil {
		logger.FatalLog.Printf("cron: error syncing database with filesystem: %v", err)
	}

	cfg := config.GetConfig()

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		logger.ErrorLog.Println("error creating Discord session: ", err)
		return
	}

	// Register messageCreate as a callback for the messageCreate events.
	dg.AddHandler(viewInstance.AudioMessageHandler)

	// Register interaction handler for button clicks
	dg.AddHandler(viewInstance.InteractionHandler)

	// Register guildCreate as a callback for the guildCreate events.
	dg.AddHandler(guildCreate)

	// We need information about guilds (which includes their channels),
	// messages and voice states.
	dg.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsGuildVoiceStates | discordgo.IntentDirectMessages | discordgo.IntentGuildMembers

	// Open the websocket and begin listening.
	err = dg.Open()
	if err != nil {
		logger.ErrorLog.Println("error opening Discord session: ", err)
		return
	}

	// Scan the sound directory for sound files. Sync them with the DB.
	fsSounds, err = modelInstance.ScanDirectory()
	if err != nil {
		logger.FatalLog.Fatalf("error scanning sound directory: %v", err)
	}
	err = viewInstance.SyncDatabaseWithFileSystem(fsSounds)
	if err != nil {
		logger.FatalLog.Printf("error syncing sound files: %v", err)
	}

	// Wait here until CTRL-C or other term signal is received.
	logger.InfoLog.Println("Bot is now running")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()

	ctrl.Run()
}

func NewBot() {
	Init()
}

// This function will be called (due to AddHandler above) every time a new
// guild is joined.
func guildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	if event.Guild.Unavailable {
		return
	}
	// load the guild into the
	config.LoadGuild(event)
	model.Meta = model.NewInfo()
	logger.InfoLog.Printf("Joined guild: %s", event.Guild.Name)
}
