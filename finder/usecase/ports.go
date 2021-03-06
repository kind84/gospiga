package usecase

import (
	"gospiga/finder/domain"
	"gospiga/finder/fulltext"
	"gospiga/pkg/streamer"
)

type DB interface {
	IDExists(id string) (bool, error)
	Tags(index, field string) ([]string, error)
}

type FT interface {
	IndexRecipe(*domain.Recipe) error
	DeleteRecipe(string) error
	SearchRecipes(string) ([]*fulltext.Recipe, error)
	SearchByTag([]string) ([]*fulltext.Recipe, error)
}

type Streamer interface {
	Ack(stream, group string, ids ...string) error
	Add(string, *streamer.Message) error
	ReadGroup(*streamer.StreamArgs) error
}
