package gql

import "context"

//go:generate go run github.com/99designs/gqlgen --verbose

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

	askFinder := false
	if first != nil {
		askFinder = true
	}
	if after != nil {
		askFinder = true
	}
	if len(tags) > 0 {
		askFinder = true
	}
	if len(ingredients) > 0 {
		askFinder = true
	}
	if query != nil {
		askFinder = true
	}

	if askFinder {
		// grpc call to finder here
	}

	return nil, nil
}
