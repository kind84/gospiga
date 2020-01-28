package usecase

import (
	"context"
)

func (a *app) SearchRecipes(ctx context.Context, query string) ([]string, error) {
	return a.ft.SearchRecipes(query)
}
