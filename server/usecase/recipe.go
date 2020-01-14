package usecase

import (
	"context"

	"github.com/kind84/gospiga/pkg/streamer"
	"github.com/kind84/gospiga/server/domain"
)

const stream = "new-recipes"

// NewRecipe inform of a new recipe id sending it over the stream.
func (a *App) NewRecipe(ctx context.Context, recipeID string) error {
	msg := &streamer.Message{Payload: recipeID}
	return a.streamer.Add(stream, msg)
}

// SearchRecipes matching the query string.
func (a *App) SearchRecipes(ctx context.Context, query string) ([]*domain.Recipe, error) {
	ids, err := a.stub.SearchRecipes(ctx, query)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return []*domain.Recipe{}, nil
	}

	return a.service.GetRecipesByIDs(ctx, ids)
}
