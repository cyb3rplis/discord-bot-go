package controller

import (
	"context"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/cyb3rplis/discord-bot-go/dlog"
	"github.com/cyb3rplis/discord-bot-go/model"
	"github.com/go-co-op/gocron"
)

func (c *Controller) SyncCronjob(ctx context.Context) {
	interval := time.Minute
	for {
		select {
		case <-time.After(interval):

		case <-ctx.Done():
			return
		}
	}
}

func (c *Controller) SyncSoundDirectories() {
	interval := time.Minute * 1
	s := gocron.NewScheduler(time.UTC)
	s.Every(interval).Do(func() {
		fsSounds, err := c.model.ScanDirectory()
		if err != nil {
			dlog.FatalLog.Printf("Cron: error scanning sound directory: %v", err)
		}
		err = c.view.SyncDatabaseWithFileSystem(fsSounds)
		if err != nil {
			dlog.FatalLog.Printf("Cron: error syncing database with filesystem: %v", err)
		}
		//dlog.InfoLog.Println("Cron: database synced with filesystem")
	})
	// starts the scheduler asynchronously
	s.StartAsync()
	// starts the scheduler and blocks current execution path
	//s.StartBlocking()
}

func (c *Controller) SyncUsers(s *discordgo.Session, m *model.Model) {
	interval := time.Minute * 15
	scheduler := gocron.NewScheduler(time.UTC)
	scheduler.Every(interval).Do(func() {
		m.FetchAndStoreGuildMembers(s)
	})
	scheduler.StartAsync()
}

func (c *Controller) CheckBotActivity(s *discordgo.Session, m *model.Model) {
	botTimeout, err := strconv.Atoi(m.Config.BotTimeout)
	if err != nil {
		dlog.FatalLog.Fatalf("Failed to convert bot timeout to integer: %v", err)
	}

	interval := time.Minute * 5
	scheduler := gocron.NewScheduler(time.UTC)
	scheduler.Every(interval).Do(func() {
		thresholdTime := model.Meta.BotActivity.Add(time.Duration(botTimeout) * time.Minute)

		if time.Now().After(thresholdTime) {
			m.LeaveVoiceChannel(s)
		}
	})
	scheduler.StartAsync()
}

func (c *Controller) CheckNewSounds(s *discordgo.Session, m *model.Model) {
	interval := time.Minute * 1
	scheduler := gocron.NewScheduler(time.UTC)
	scheduler.Every(interval).Do(func() {
		m.PinNewSoundButtons(s)
	})
	scheduler.StartAsync()
}
