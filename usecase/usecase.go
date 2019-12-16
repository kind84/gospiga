package usecase

import (
	"context"
)

type App struct {
	service  Service
	db       DB
	streamer Streamer
}

func NewApp(ctx context.Context, service Service, db DB, streamer Streamer) *App {
	app := &App{
		service:  service,
		db:       db,
		streamer: streamer,
	}

	// start streamer to listen for new recipes.

	return app
}
