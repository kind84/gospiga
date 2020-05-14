package usecase

import (
	"context"

	"gospiga/pkg/streamer"
	"gospiga/pkg/types"
	"gospiga/server/domain"
)

type DB interface {
	AllTagsImages(context.Context) ([]*domain.Tag, error)
}

type Service interface {
	SaveRecipe(context.Context, *domain.Recipe) error
	UpdateRecipe(context.Context, *domain.Recipe) error
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
	GetRecipe(ctx context.Context, recipeID string) (*types.Recipe, error)
	GetAllRecipeIDs(context.Context) ([]string, error)
}

type Stub interface {
	AllRecipeTags(context.Context) ([]string, error)
}
