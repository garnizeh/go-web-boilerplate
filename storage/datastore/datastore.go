package datastore

import (
	"embed"

	"github.com/garnizeh/go-web-boilerplate/storage"
)

//go:embed sql/migrations/*
var Migrations embed.FS

func Factory(tx storage.DBTX) *Queries {
	return New(tx)
}
