package cronjob

import (
	"github.com/cyb3rplis/discord-bot-go/logger"
	"github.com/cyb3rplis/discord-bot-go/sound"
	"github.com/cyb3rplis/discord-bot-go/utils"
	"github.com/robfig/cron/v3"
)

func InitCron() {
	cron := cron.New()
	cron.AddFunc("*/10 * * * *", func() {
		fsSounds, err := utils.ScanDirectory()
		if err != nil {
			logger.FatalLog.Printf("cron: error scanning sound directory: %v", err)
		}
		err = sound.SyncDatabaseWithFileSystem(fsSounds)
		if err != nil {
			logger.FatalLog.Printf("cron: error syncing database with filesystem: %v", err)
		}
	})

	logger.InfoLog.Println("Initiated Cronjob")
	cron.Start()
}
