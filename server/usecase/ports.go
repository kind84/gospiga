package usecase

import (
	"context"

	"github.com/kind84/gospiga/pkg/streamer"
	"github.com/kind84/gospiga/server/domain"
)

type DB interface {
}

type Service interface {
	SaveRecipe(context.Context, *domain.Recipe) error
	IDSaved(context.Context, string) (bool, error)
}

type Streamer interface {
	Ack(string, string, ...string) error
	Add(string, *streamer.Message) error
	AckAndAdd(from *streamer.StreamArgs, toStream string, id string, msg *streamer.Message) error
	ReadGroup(context.Context, *streamer.StreamArgs, chan streamer.Message, chan struct{})
}

type Provider interface {
	GetRecipe(context.Context, string) (*domain.Recipe, error)
}
