package controller

import (
	"context"
	"github.com/cyb3rplis/discord-bot-go/logger"
	"github.com/cyb3rplis/discord-bot-go/model"
	"github.com/cyb3rplis/discord-bot-go/view"
)

type Controller struct {
	model *model.Model
	view  *view.API
}

func New(model *model.Model, view *view.API) *Controller {
	return &Controller{model: model, view: view}
}

func (c *Controller) Run(ctx context.Context) {
	// run background tasks here
	logger.InfoLog.Println("Initializing Background Tasks")
	startBackgroundFunctions(ctx,
		c.SyncFiles,
	)
}
