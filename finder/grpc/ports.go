package grpc

import "context"

type App interface {
	AllRecipeTags(context.Context) ([]string, error)
}
