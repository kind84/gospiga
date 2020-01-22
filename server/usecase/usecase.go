package usecase

import (
	"context"
)

type App struct {
	service  Service
	db       DB
	streamer Streamer
	provider Provider
	stub     Stub
}

func NewApp(ctx context.Context, service Service, db DB, streamer Streamer, provider Provider, stub Stub) (*App, error) {
	app := &App{
		service:  service,
		db:       db,
		streamer: streamer,
		provider: provider,
		stub:     stub,
	}

	// start streamer to listen for new recipes.
	err := app.readRecipes(ctx)
	if err != nil {
		return nil, err
	}

	return app, nil
}
