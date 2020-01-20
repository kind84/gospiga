package api

import (
	"context"

	"github.com/kind84/gospiga/server/domain"
)

type App interface {
	NewRecipe(context.Context, string) error
	UpdatedRecipe(context.Context, string) error
	SearchRecipes(context.Context, string) ([]*domain.Recipe, error)
}
