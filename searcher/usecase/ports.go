package usecase

import (
	"context"

	"github.com/kind84/gospiga/pkg/streamer"
	"github.com/kind84/gospiga/searcher/domain"
)

type FT interface {
}

type Service interface {
	IndexRecipe(context.Context, *domain.Recipe) error
}

type Streamer interface {
	Add(context.Context, string, *streamer.Message) error
	ReadGroup(context.Context, *streamer.StreamArgs, chan streamer.Message, chan struct{})
}
