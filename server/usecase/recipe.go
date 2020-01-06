package usecase

import (
	"context"

	"github.com/kind84/gospiga/pkg/streamer"
)

const stream = "new-recipes"

func (a *App) NewRecipe(ctx context.Context, recipeID string) error {
	msg := &streamer.Message{Payload: recipeID}
	return a.streamer.Add(ctx, stream, msg)
}
