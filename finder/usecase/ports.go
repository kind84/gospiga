package usecase

import (
	"context"

	"github.com/kind84/gospiga/finder/domain"
	"github.com/kind84/gospiga/pkg/streamer"
)

type DB interface {
	IDExists(id string) (bool, error)
}

type FT interface {
	IndexRecipe(*domain.Recipe) error
	DeleteRecipe(string) error
	SearchRecipes(string) ([]string, error)
}

type Streamer interface {
	Ack(stream, group string, ids ...string) error
	Add(string, *streamer.Message) error
	ReadGroup(context.Context, *streamer.StreamArgs) error
}
