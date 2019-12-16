package api

import (
	"context"
)

type App interface {
	NewRecipe(context.Context, string) error
}
