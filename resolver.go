//go:generate go run github.com/99designs/gqlgen

package gospiga

import (
	"context"
	"encoding/json"
	"log"

	"github.com/dgraph-io/dgo"
	"github.com/dgraph-io/dgo/protos/api"
	"google.golang.org/grpc"
) // THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

type Resolver struct{}

func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}

type queryResolver struct{ *Resolver }

type mutationResolver struct{ *Resolver }

func (r *queryResolver) Recipes(ctx context.Context) ([]Recipe, error) {
	c := newClient()
	txn := c.NewReadOnlyTxn()

	const q = `
		{
			recipes (func: has(title)) {
				title
				ingredients {
					name
					quantity
				}
			}
		}
	`
	resp, err := txn.Query(context.Background(), q)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}

	var jres struct {
		Recipes []Recipe `json:"recipes"`
	}

	if err := json.Unmarshal(resp.GetJson(), &jres); err != nil {
		return nil, err
	}

	return jres.Recipes, nil
}

func (r *mutationResolver) CreateRecipe(ctx context.Context, nr NewRecipe) (*Recipe, error) {
	c := newClient()
	txn := c.NewTxn()
	defer txn.Discard(ctx)

	mu := &api.Mutation{
		CommitNow: true,
	}

	rb, err := json.Marshal(nr)
	if err != nil {
		log.Fatal(err)
	}

	mu.SetJson = rb
	assigned, err := c.NewTxn().Mutate(ctx, mu)
	if err != nil {
		log.Fatal(err)
	}

	variables := map[string]string{"$id": assigned.Uids["blank-0"]}
	q := `query Recipe($id: string){
		recipe(func: uid($id)) {
			uid
			title
			ingredients {
				uid
				name
				quantity
			}
		}
	}`

	resp, err := c.NewTxn().QueryWithVars(ctx, q, variables)
	if err != nil {
		log.Fatal(err)
	}

	type Root struct {
		Recipe []Recipe `json:"recipe"`
	}

	var rt Root
	err = json.Unmarshal(resp.Json, &rt)
	if err != nil {
		log.Fatal(err)
	}

	return &rt.Recipe[0], nil
}

func newClient() *dgo.Dgraph {
	// Dial a gRPC connection. The address to dial to can be configured when
	// setting up the dgraph cluster.
	d, err := grpc.Dial("localhost:9080", grpc.WithInsecure())
	if err != nil {
		log.Fatal(err)
	}

	return dgo.NewDgraphClient(
		api.NewDgraphClient(d),
	)
}
