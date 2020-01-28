package usecase

import (
	"context"
	"sync"

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
	Ack(string, string, ...string) error
	Add(string, *streamer.Message) error
	AckAndAdd(fromStream, toStream, group, id string, msg *streamer.Message) error
	ReadGroup(context.Context, *streamer.StreamArgs, chan streamer.Message, chan struct{}, *sync.WaitGroup) error
}

type Provider interface {
	GetRecipe(context.Context, string) (*domain.Recipe, error)
}

type Stub interface {
	SearchRecipes(context.Context, string) ([]string, error)
}
