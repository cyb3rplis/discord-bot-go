package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/cyb3rplis/discord-bot-go/utils"

	"github.com/cyb3rplis/discord-bot-go/logger"
	"github.com/cyb3rplis/discord-bot-go/model"

	"github.com/cyb3rplis/discord-bot-go/config"
	"github.com/cyb3rplis/discord-bot-go/cronjob"
	"github.com/cyb3rplis/discord-bot-go/db"
	"github.com/cyb3rplis/discord-bot-go/sound"

	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/message"
)

func init() {
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
	model.Bot = model.New(&m)

	// Check if the sound directory exists
	if _, err := os.Stat(model.Bot.Config.SoundsDir); os.IsNotExist(err) {
		os.Mkdir(model.Bot.Config.SoundsDir, 0755)
	}

	fsSounds, err := utils.ScanDirectory()
	if err != nil {
		logger.FatalLog.Printf("cron: error scanning sound directory: %v", err)
	}
	err = sound.SyncDatabaseWithFileSystem(fsSounds)
	if err != nil {
		logger.FatalLog.Printf("cron: error syncing database with filesystem: %v", err)
	}
}

func main() {
	cfg := config.GetConfig()
	cronjob.InitCron()

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		logger.ErrorLog.Println("error creating Discord session: ", err)
		return
	}

	// Register messageCreate as a callback for the messageCreate events.
	dg.AddHandler(message.AudioMessageHandler)

	// Register interaction handler for button clicks
	dg.AddHandler(sound.InteractionHandler)

	// Register guildCreate as a callback for the guildCreate events.
	dg.AddHandler(guildCreate)

	// We need information about guilds (which includes their channels),
	// messages and voice states.
	dg.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsGuildVoiceStates | discordgo.IntentDirectMessages

	// Open the websocket and begin listening.
	err = dg.Open()
	if err != nil {
		logger.ErrorLog.Println("error opening Discord session: ", err)
		return
	}

	// Scan the sound directory for sound files. Sync them with the DB.
	fsSounds, err := utils.ScanDirectory()
	if err != nil {
		logger.FatalLog.Fatalf("error scanning sound directory: %v", err)
	}
	err = sound.SyncDatabaseWithFileSystem(fsSounds)
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
}

// This function will be called (due to AddHandler above) every time a new
// guild is joined.
func guildCreate(s *discordgo.Session, event *discordgo.GuildCreate) {
	if event.Guild.Unavailable {
		return
	}

	logger.InfoLog.Printf("Joined guild: %s", event.Guild.Name)
}
