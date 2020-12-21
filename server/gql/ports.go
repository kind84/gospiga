package gql

// ** TODO: fix missing mockgen in CI //go:generate mockgen -source ports.go -destination portsmock_test.go -package gql_test

import (
	"context"

	"gospiga/pkg/types"
	"gospiga/server/domain"
)

type App interface {
	SearchRecipes(context.Context, *types.SearchRecipesArgs) ([]*domain.Recipe, error)
}
