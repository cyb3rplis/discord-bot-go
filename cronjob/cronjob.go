package cronjob

import (
	"github.com/cyb3rplis/discord-bot-go/logger"
	"github.com/cyb3rplis/discord-bot-go/sound"
	"github.com/robfig/cron/v3"
)

func InitCron() {
	cron := cron.New()
	cron.AddFunc("*/5 * * * *", func() {
		sound.InsertCategoriesAndSounds()
	})

	logger.InfoLog.Println("Initiated Cronjob")
	cron.Start()
}
