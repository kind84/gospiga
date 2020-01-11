package api

import (
	"context"
)

type App interface {
	Search(context.Context, string) ([]string, error)
}
