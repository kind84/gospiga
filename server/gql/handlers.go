package gql

import (
	"context"

	"github.com/gin-gonic/gin"

	"github.com/99designs/gqlgen/graphql"
	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
)

// Defining the Graphql handler
func GraphqlHandler(app App) gin.HandlerFunc {
	// NewExecutableSchema and Config are in the generated.go file
	config := Config{Resolvers: NewResolver(app)}

	// match dgraph directives but ignore them
	config.Directives.Dgraph = func(ctx context.Context, obj interface{}, next graphql.Resolver, pred *string) (res interface{}, err error) {
		return next(ctx)
	}
	config.Directives.HasInverse = func(ctx context.Context, obj interface{}, next graphql.Resolver, field *string) (res interface{}, err error) {
		return next(ctx)
	}
	config.Directives.Search = func(ctx context.Context, obj interface{}, next graphql.Resolver, by []*string) (res interface{}, err error) {
		return next(ctx)
	}
	config.Directives.Id = func(ctx context.Context, obj interface{}, next graphql.Resolver) (res interface{}, err error) {
		return next(ctx)
	}

	h := handler.NewDefaultServer(NewExecutableSchema(config))

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// Defining the Playground handler
func PlaygroundHandler() gin.HandlerFunc {
	h := playground.Handler("GraphQL", "/query")

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
