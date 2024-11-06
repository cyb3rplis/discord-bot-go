package controller

import (
	"context"
	"github.com/cyb3rplis/discord-bot-go/dlog"
	"github.com/go-co-op/gocron"
	"sync"
	"time"
)

type backgroundFunc func(context.Context)

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

func startBackgroundFunctions(ctx context.Context, fncs ...backgroundFunc) {
	var wg sync.WaitGroup
	for _, fnc := range fncs {
		wg.Add(1)
		go func(fnc backgroundFunc) {
			defer wg.Done()
			fnc(ctx)
		}(fnc)
	}
	wg.Wait()
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
