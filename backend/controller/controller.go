package controller

import (
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

func (c *Controller) Run() {
	// do something
	// run background tasks here
	c.SyncCronjob()
}
