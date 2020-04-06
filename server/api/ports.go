package api

import (
	"context"
)

// App interface defines methods to be exposed by the api service.
type App interface {
	NewRecipe(context.Context, string) error
	UpdatedRecipe(context.Context, string) error
	DeletedRecipe(context.Context, string) error
}
