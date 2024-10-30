package cronjob

import (
	"github.com/cyb3rplis/discord-bot-go/logger"
	"github.com/cyb3rplis/discord-bot-go/sound"
	"github.com/cyb3rplis/discord-bot-go/utils"
	"github.com/robfig/cron/v3"
)

func InitCron() {
	syncCronjob()
}

func syncCronjob() {
	cronJob := cron.New()
	_, _ = cronJob.AddFunc("*/5 * * * *", func() {
		fsSounds, err := utils.ScanDirectory()
		if err != nil {
			logger.FatalLog.Printf("Cron: error scanning sound directory: %v", err)
		}
		err = sound.SyncDatabaseWithFileSystem(fsSounds)
		if err != nil {
			logger.FatalLog.Printf("Cron: error syncing database with filesystem: %v", err)
		}
		logger.InfoLog.Println("Cron: database synced with filesystem")
	})

	logger.InfoLog.Println("Initiated Sync Cronjob")
	cronJob.Start()
}
