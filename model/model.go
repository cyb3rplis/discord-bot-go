package model

import "database/sql"

var Bot *Model

type Model struct {
	Db *sql.DB
}

func New(m *Model) *Model {
	return &Model{Db: m.Db}
}
