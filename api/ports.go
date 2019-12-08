package api

import (
	"context"
)

type Storer interface {
	Save(context.Context, interface{}) error
}

type App interface {
	NewRecipe(context.Context, interface{}) error
}
