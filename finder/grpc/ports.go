package grpc

import "gospiga/pkg/types"

type App interface {
	AllRecipeTags() ([]string, error)
	SearchIDs(types.SearchIDsArgs) ([]uint64, error)
}
