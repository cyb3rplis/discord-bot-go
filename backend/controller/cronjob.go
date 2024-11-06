package controller

import (
	"context"
	"sync"
	"time"

	"github.com/cyb3rplis/discord-bot-go/dlog"
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

func (c *Controller) SyncFiles(ctx context.Context) {
	interval := time.Minute * 2
	dlog.InfoLog.Println("Run Background Task", "INTERVAL", interval)
	fsSounds, err := c.model.ScanDirectory()
	if err != nil {
		dlog.FatalLog.Printf("Cron: error scanning sound directory: %v", err)
	}
	err = c.view.SyncDatabaseWithFileSystem(fsSounds)
	if err != nil {
		dlog.FatalLog.Printf("Cron: error syncing database with filesystem: %v", err)
	}
	for {
		select {
		case <-time.After(interval):
			dlog.InfoLog.Println("Run Background Task", "INTERVAL", interval)
			fsSounds, err := c.model.ScanDirectory()
			if err != nil {
				dlog.FatalLog.Printf("Cron: error scanning sound directory: %v", err)
			}
			err = c.view.SyncDatabaseWithFileSystem(fsSounds)
			if err != nil {
				dlog.FatalLog.Printf("Cron: error syncing database with filesystem: %v", err)
			}
			dlog.InfoLog.Println("Cron: database synced with filesystem")
		case <-ctx.Done():
			return
		}
	}
}
