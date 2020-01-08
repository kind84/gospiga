package api

import (
	"context"
)

type App interface {
	SearchRecipe(context.Context, string) error
}
