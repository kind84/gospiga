package domain

import (
	"context"
)

type DB interface {
	SaveRecipe(context.Context, *Recipe) error
}
