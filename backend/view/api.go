package view

import "github.com/cyb3rplis/discord-bot-go/model"

type API struct {
	model *model.Model
}

func New(model *model.Model) *API {
	a := &API{model: model}
	return a
}
