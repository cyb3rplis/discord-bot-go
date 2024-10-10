package main

import (
	"fmt"
	"github.com/cyb3rplis/discord-bot-go/config"
	"github.com/cyb3rplis/discord-bot-go/sound"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/message"
)

func main() {
	token := config.GetValueString("general", "token", "-")

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println("error creating Discord session: ", err)
		return
	}

	// Register messageCreate as a callback for the messageCreate events.
	dg.AddHandler(message.AudioMessageHandler)

	// Register interaction handler for button clicks
	dg.AddHandler(sound.InteractionCreate)

	// Register guildCreate as a callback for the guildCreate events.
	dg.AddHandler(guildCreate)

	// We need information about guilds (which includes their channels),
	// messages and voice states.
	dg.Identify.Intents = discordgo.IntentsGuilds | discordgo.IntentsGuildMessages | discordgo.IntentsGuildVoiceStates

	// Open the websocket and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening Discord session: ", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running..  Press CTRL-C to exit.")
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
	for _, channel := range event.Guild.Channels {
		if channel.ID == event.Guild.ID {
			_, _ = s.ChannelMessageSend(channel.ID, "Bot is ready to play sounds.")
			return
		}
	}
}
