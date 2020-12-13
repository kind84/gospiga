package gql

//go:generate go run github.com/99designs/gqlgen --verbose

import (
	"context"
	"gospiga/pkg/types"
)

func NewResolver(app App) *Resolver {
	return &Resolver{app: app}
}

type Resolver struct {
	app App
}

func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

type queryResolver struct{ *Resolver }

func (r *queryResolver) Recipes(ctx context.Context, first *int, after *string,
	tags []*string, ingredients []*string, query *string) ([]*Recipe, error) {

	searchArgs := types.SearchRecipesArgs{
		First: first,
		After: after,
		Query: query,
	}

	if len(tags) > 0 {
		ts := make([]string, len(tags))
		for i, t := range tags {
			ts[i] = *t
		}
		searchArgs.Tags = ts
	}
	if len(ingredients) > 0 {
		is := make([]string, len(ingredients))
		for i, s := range ingredients {
			is[i] = *s
		}
		searchArgs.Ingredients = is
	}

	_, err := r.app.SearchRecipes(ctx, &searchArgs)
	if err != nil {
		return nil, err
	}

	return nil, nil
}
