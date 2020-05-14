package api

import (
	"context"

	"gospiga/pkg/types"
)

// App interface defines methods to be exposed by the api service.
type App interface {
	NewRecipe(context.Context, string) error
	UpdatedRecipe(context.Context, string) error
	DeletedRecipe(context.Context, string) error
	AllTagsImages(context.Context) ([]*types.Tag, error)
	LoadRecipes(ctx context.Context) error
}
