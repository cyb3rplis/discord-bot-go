package controller

import (
	"github.com/cyb3rplis/discord-bot-go/logger"
	"github.com/robfig/cron/v3"
)

func (c *Controller) SyncCronjob() {
	cronJob := cron.New()
	_, _ = cronJob.AddFunc("*/5 * * * *", func() {
		fsSounds, err := c.model.ScanDirectory()
		if err != nil {
			logger.FatalLog.Printf("Cron: error scanning sound directory: %v", err)
		}
		err = c.view.SyncDatabaseWithFileSystem(fsSounds)
		if err != nil {
			logger.FatalLog.Printf("Cron: error syncing database with filesystem: %v", err)
		}
		logger.InfoLog.Println("Cron: database synced with filesystem")
	})

	logger.InfoLog.Println("Initiated Sync Cronjob")
	cronJob.Start()
}
