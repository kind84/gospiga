package api

import (
	"context"
)

type App interface {
	SearchRecipes(context.Context, string) ([]string, error)
}
