package usecase

import (
	"context"
	"sync"

	"github.com/kind84/gospiga/finder/domain"
	"github.com/kind84/gospiga/pkg/streamer"
)

type DB interface {
	IDExists(id string) (bool, error)
}

type FT interface {
	IndexRecipe(recipe *domain.Recipe) error
	SearchRecipes(query string) ([]string, error)
}

type Streamer interface {
	Ack(string, string, ...string) error
	Add(string, *streamer.Message) error
	ReadGroup(context.Context, *streamer.StreamArgs, chan streamer.Message, chan struct{}, *sync.WaitGroup) error
}
