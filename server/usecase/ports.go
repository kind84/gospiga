package usecase

import (
	"context"
	"sync"

	"gospiga/pkg/streamer"
	"gospiga/pkg/types"
	"gospiga/server/domain"
)

type DB interface {
	AllTagsImages(context.Context) ([]*domain.Tag, error)
}

type Service interface {
	SaveRecipe(context.Context, *domain.Recipe) error
	UpdateRecipe(context.Context, *domain.Recipe) (string, error)
	DeleteRecipe(context.Context, string) error
	GetRecipeByID(context.Context, string) (*domain.Recipe, error)
	GetRecipesByIDs(context.Context, []string) ([]*domain.Recipe, error)
	IDSaved(context.Context, string) (bool, error)
}

type Streamer interface {
	Ack(ctx context.Context, stream, group string, ids ...string) error
	Add(context.Context, string, *streamer.Message) error
	AckAndAdd(ctx context.Context, fromStream, toStream, group, id string, msg *streamer.Message) error
	ReadGroup(context.Context, *sync.WaitGroup, *streamer.StreamArgs) error
}

type Provider interface {
	GetRecipe(ctx context.Context, recipeID string) (*types.Recipe, error)
	GetAllRecipeIDs(context.Context) ([]string, error)
}

type Stub interface {
	AllRecipeTags(context.Context) ([]string, error)
}
