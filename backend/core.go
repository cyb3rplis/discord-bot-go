package backend

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/config"
	"github.com/cyb3rplis/discord-bot-go/controller"
	"github.com/cyb3rplis/discord-bot-go/db"
	"github.com/cyb3rplis/discord-bot-go/dlog"
	"github.com/cyb3rplis/discord-bot-go/model"
	"github.com/cyb3rplis/discord-bot-go/view"
)

var readyMutex = &sync.Mutex{}

func Init() {
	m, dbClose, err := db.InitModel()
	if err != nil {
		dlog.FatalLog.Fatalf("Failed to initialize database: %v", err)
	}
	defer func() {
		if err != nil {
			if err := dbClose(); err != nil {
				dlog.FatalLog.Fatalf("Failed to close database: %v", err)
			}
		}
	}()
	modelInstance := model.New(&m)
	viewInstance := view.New(modelInstance)
	ctrl := controller.New(modelInstance, viewInstance)

	// Check if the sound directory exists
	if _, err := os.Stat(modelInstance.Config.SoundsDir); os.IsNotExist(err) {
		if err != nil {
			dlog.FatalLog.Fatalf("Failed to check sound directory: %v", err)
		}
		err = os.Mkdir(m.Config.SoundsDir, 0755)
		if err != nil {
			dlog.FatalLog.Fatalf("Failed to create sound directory: %v", err)
		}
	}

	fsSounds, err := modelInstance.ScanDirectory()
	if err != nil {
		dlog.FatalLog.Fatalf("cron: error scanning sound directory: %v", err)
	}
	err = viewInstance.SyncDatabaseWithFileSystem(fsSounds)
	if err != nil {
		dlog.FatalLog.Fatalf("cron: error syncing database with filesystem: %v", err)
	}

	cfg := config.GetConfig()
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + cfg.Token)
	if err != nil {
		dlog.FatalLog.Fatalf("error creating Discord session: %v", err)
	}

	//set bot ready
	dg.AddHandlerOnce(func(s *discordgo.Session, event *discordgo.Ready) {
		readyMutex.Lock()
		defer readyMutex.Unlock()
		dlog.InfoLog.Println("Bot is ready")
		view.BotReady = true
	})

	// Register guildCreate as a callback for the guildCreate events.
	// This will be called every time a new guild is joined.
	dg.AddHandlerOnce(func(s *discordgo.Session, event *discordgo.GuildCreate) {
		guildCreate(event)
		modelInstance.FetchAndStoreGuildMembers(s)
		view.RegisterPromptInteractionsButtons(s)
		view.RegisterPromptInteractionsCreate(s)
		view.RegisterPromptInteractionsDelete(s)
		view.RegisterPromptInteractionsAudio(s)
		viewInstance.RegisterPromptInteractionsFavorite(s)
		viewInstance.RegisterPromptInteractionsGulag(s)
		view.RegisterPromptInteractionsStats(s)
		viewInstance.RegisterPromptInteractionsPlaySound(s)
	})

	// Register prompt interaction handlers
	// This will be called every time a new interaction is created.
	dg.AddHandler(viewInstance.InteractionHandler)         //interaction handler
	dg.AddHandler(viewInstance.PromptInteractionButtons)   //buttons
	dg.AddHandler(viewInstance.PromptInteractionCreate)    //create
	dg.AddHandler(viewInstance.PromptInteractionDelete)    //delete
	dg.AddHandler(viewInstance.PromptInteractionAudio)     //audio
	dg.AddHandler(viewInstance.PromptInteractionFavorite)  //favorite
	dg.AddHandler(viewInstance.PromptInteractionGulag)     //gulag
	dg.AddHandler(viewInstance.PromptInteractionStats)     //stats
	dg.AddHandler(viewInstance.PromptInteractionPlaySound) //play sound

	// messages and voice states.
	dg.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsGuildVoiceStates | discordgo.IntentDirectMessages | discordgo.IntentGuildMembers

	// Open the websocket and begin listening.
	err = dg.Open()
	if err != nil {
		dlog.ErrorLog.Fatalf("error opening Discord session: %v", err)
	}

	// Scan the sound directory for sound files. Sync them with the DB.
	fsSounds, err = modelInstance.ScanDirectory()
	if err != nil {
		dlog.FatalLog.Fatalf("error scanning sound directory: %v", err)
	} else {
		dlog.InfoLog.Println("Scanning sound directory for new audio")
	}
	err = viewInstance.SyncDatabaseWithFileSystem(fsSounds)
	if err != nil {
		dlog.FatalLog.Printf("error syncing sound files: %v", err)
	} else {
		dlog.InfoLog.Println("Inserting new sounds into database")
	}

	//background TODO: move this to controller
	go ctrl.SyncSoundDirectories()

	// Wait here until CTRL-C or other term signal is received.
	dlog.InfoLog.Println("Bot is now running")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	ctx, cancel := context.WithCancel(context.Background())
	s := make(chan os.Signal, 1)
	signal.Notify(s, os.Interrupt)
	go func() {
		<-s
		cancel()
	}()

	// Cleanly close down the Discord session.
	_ = dg.Close()

	ctrl.Run(ctx)
}

func NewBot() {
	Init()
}

// This function will be called (due to AddHandler above) every time a new
// guild is joined.
func guildCreate(event *discordgo.GuildCreate) {
	if event.Guild.Unavailable {
		dlog.FatalLog.Fatalf("Guild is unavailable")
	}
	dlog.InfoLog.Printf("Joined guild: %s", event.Guild.Name)
	dlog.InfoLog.Printf("Guild ID: %s", event.Guild.ID)
	config.LoadGuild(event.Guild)
	model.Meta = model.NewInfo()
}
