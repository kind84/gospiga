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
	DeleteRecipe(context.Context, string) error
	GetRecipeByID(context.Context, string) (*domain.Recipe, error)
	GetRecipesByIDs(context.Context, []string) ([]*domain.Recipe, error)
	IDSaved(context.Context, string) (bool, error)
}

type Streamer interface {
	Ack(stream, group string, ids ...string) error
	Add(string, *streamer.Message) error
	AckAndAdd(fromStream, toStream, group, id string, msg *streamer.Message) error
	ReadGroup(*streamer.StreamArgs) error
}

type Provider interface {
	GetRecipe(context.Context, string) (*domain.Recipe, error)
}

type Stub interface {
	SearchRecipes(context.Context, string) ([]string, error)
}
