package dgraph

import (
	"context"
	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"google.golang.org/grpc"

	"github.com/kind84/gospiga/domain"
)

type DB struct {
	*dgo.Dgraph
}

func NewDB() (*DB, error) {
	d, err := grpc.Dial("localhost:9080", grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	dgraph := dgo.NewDgraphClient(
		api.NewDgraphClient(d),
	)

	return &DB{dgraph}, nil
}

func (db *DB) SaveRecipe(ctx context.Context, recipe *domain.Recipe) error {
	return nil
}
