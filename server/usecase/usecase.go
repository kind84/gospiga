package usecase

import (
	"context"
)

type app struct {
	service  Service
	db       DB
	streamer Streamer
	provider Provider
	stub     Stub
}

func NewApp(ctx context.Context, service Service, db DB, streamer Streamer, provider Provider, stub Stub) (*app, error) {
	a := &app{
		service:  service,
		db:       db,
		streamer: streamer,
		provider: provider,
		stub:     stub,
	}

	// start streamer to listen for new recipes.
	err := a.readRecipes(ctx)
	if err != nil {
		return nil, err
	}

	return a, nil
}
