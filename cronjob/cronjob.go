package cronjob

import (
	"github.com/cyb3rplis/discord-bot-go/logger"
	"github.com/cyb3rplis/discord-bot-go/sound"
	"github.com/robfig/cron/v3"
)

func InitCron() {
	cron := cron.New()
	cron.AddFunc("*/10 * * * *", func() {
		fsSounds, err := sound.ScanDirectory()
		if err != nil {
			logger.FatalLog.Fatalf("cron: error scanning sound directory: %v", err)
		}
		sound.SyncDatabaseWithFileSystem(fsSounds)
	})

	logger.InfoLog.Println("Initiated Cronjob")
	cron.Start()
}
