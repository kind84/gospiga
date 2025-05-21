package usecase

import (
	"context"
	"gospiga/finder/domain"
	"gospiga/finder/fulltext"
	"gospiga/pkg/streamer"
	"sync"
)

type DB interface {
	IDExists(id string) (bool, error)
	Tags(ctx context.Context, index, field string) ([]string, error)
}

type FT interface {
	IndexRecipe(*domain.Recipe) error
	DeleteRecipe(string) error
	SearchRecipes(string) ([]*fulltext.Recipe, error)
	SearchByTag([]string) ([]*fulltext.Recipe, error)
}

type Streamer interface {
	Ack(ctx context.Context, stream, group string, ids ...string) error
	Add(context.Context, string, *streamer.Message) error
	ReadGroup(context.Context, *sync.WaitGroup, *streamer.StreamArgs) error
}
