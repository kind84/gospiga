package domain

import (
	"context"

	"gospiga/pkg/types"
)

// DB defines the domain database capabilities.
type DB interface {
	SaveRecipe(context.Context, *Recipe) error
	UpdateRecipe(context.Context, *Recipe) (string, error)
	DeleteRecipe(context.Context, string) error
	GetRecipeByID(context.Context, string) (*Recipe, error)
	GetRecipesByUIDs(context.Context, []string) ([]*Recipe, error)
	IDSaved(context.Context, string) (bool, error)
	SearchRecipes(context.Context, *types.SearchRecipesArgs) ([]*Recipe, error)
}
