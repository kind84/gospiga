package api

import (
	"context"

	"github.com/kind84/gospiga/server/domain"
)

// App interface defines methods to be exposed by the api service.
type App interface {
	NewRecipe(context.Context, string) error
	UpdatedRecipe(context.Context, string) error
	DeletedRecipe(context.Context, string) error
	SearchRecipes(context.Context, string) ([]*domain.Recipe, error)
}
