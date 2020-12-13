package grpc

import "gospiga/pkg/types"

type App interface {
	AllRecipeTags() ([]string, error)
	SearchIDs(*types.SearchRecipesArgs) ([]string, error)
}
