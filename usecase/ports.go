package usecase

import (
	"context"

	"github.com/kind84/gospiga/domain"
	"github.com/kind84/gospiga/streamer"
)

type DB interface {
}

type Service interface {
	SaveRecipe(context.Context, *domain.Recipe) error
}

type Streamer interface {
	Add(context.Context, string, *streamer.Message) error
	ReadGroup(context.Context, *streamer.StreamArgs, chan streamer.Message, chan struct{})
}

type Provider interface {
	GetRecipe(context.Context, string) (*domain.Recipe, error)
}
