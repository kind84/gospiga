package gql

// ** TODO: fix missing mockgen in CI //go:generate mockgen -source ports.go -destination portsmock_test.go -package gql_test

import (
	"context"

	"gospiga/finder/domain"
	"gospiga/pkg/types"
)

type App interface {
	SearchRecipes(context.Context, *types.SearchRecipesArgs) ([]*domain.Recipe, error)
}
