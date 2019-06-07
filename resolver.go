//go:generate go run github.com/99designs/gqlgen

package gospiga

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/dgraph-io/dgo"
	"github.com/dgraph-io/dgo/protos/api"
	"google.golang.org/grpc"

	"github.com/kind84/gospiga/models"
) // THIS CODE IS A STARTING POINT ONLY. IT WILL NOT BE UPDATED WITH SCHEMA CHANGES.

type Resolver struct{}

func (r *Resolver) Query() QueryResolver {
	return &queryResolver{r}
}

func (r *Resolver) Mutation() MutationResolver {
	return &mutationResolver{r}
}

func (r *Resolver) Recipe() RecipeResolver {
	return &recipeResolver{r}
}

type queryResolver struct{ *Resolver }

type mutationResolver struct{ *Resolver }

type recipeResolver struct{ *Resolver }

func (r *queryResolver) Recipes(ctx context.Context) ([]*models.Recipe, error) {
	c, err := newClient()
	if err != nil {
		return nil, err
	}

	txn := c.NewReadOnlyTxn()

	const q = `{
		recipes (func: has(title)) {
			uid
			title
			ingredient {
				uid
			}
			step {
				uid
			}
			createdAt
		}
	}`

	resp, err := txn.Query(ctx, q)
	if err != nil {
		return nil, err
	}

	var jres struct {
		Recipes []*models.Recipe `json:"recipes"`
	}

	if err := json.Unmarshal(resp.GetJson(), &jres); err != nil {
		return nil, err
	}

	return jres.Recipes, nil
}

func (r *mutationResolver) CreateRecipe(ctx context.Context, nr NewRecipe) (*models.Recipe, error) {
	c, err := newClient()
	if err != nil {
		return nil, err
	}
	txn := c.NewTxn()
	defer txn.Discard(ctx)

	err = c.Alter(context.Background(), &api.Operation{
		Schema: `
			title: string @index(term) .
			ingredient: uid @reverse .
			steps: uid @reverse .
			createdAt: dateTime @index(day) .
		`,
	})
	if err != nil {
		return nil, err
	}

	mu := &api.Mutation{CommitNow: true}

	rcp := struct {
		NewRecipe
		CreatedAt time.Time `json:"createdAt,omitempty"`
	}{
		NewRecipe: nr,
		CreatedAt: time.Now().UTC(),
	}

	rb, err := json.Marshal(rcp)
	if err != nil {
		return nil, err
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
			ingredient {
				uid
			}
			step {
				uid
			}
			createdAt
		}
	}`

	resp, err := c.NewTxn().QueryWithVars(ctx, q, variables)
	if err != nil {
		log.Fatal(err)
	}

	type Root struct {
		Recipe []models.Recipe `json:"recipe"`
	}

	var rt Root
	err = json.Unmarshal(resp.Json, &rt)
	if err != nil {
		log.Fatal(err)
	}

	return &rt.Recipe[0], nil
}

func (r *recipeResolver) Ingredient(ctx context.Context, obj *models.Recipe) ([]*Ingredient, error) {
	var igs []*Ingredient

	// Dgraph does not allow to pass muliple UIDs as func variable. Looping.
	for _, i := range obj.Ingredient {
		ii, err := getIngredient(ctx, i.UID)
		if err != nil {
			return nil, err
		}
		igs = append(igs, ii)
	}
	return igs, nil
}

func (r *recipeResolver) Step(ctx context.Context, obj *models.Recipe) ([]*Step, error) {
	var stps []*Step

	// Dgraph does not allow to pass muliple UIDs as func variable. Looping.
	for _, s := range obj.Step {
		ss, err := getStep(ctx, s.UID)
		if err != nil {
			return nil, err
		}
		stps = append(stps, ss)
	}
	return stps, nil
}

func getIngredient(ctx context.Context, uid string) (*Ingredient, error) {
	c, err := newClient()
	if err != nil {
		return nil, err
	}

	txn := c.NewReadOnlyTxn()

	vars := map[string]string{"$id": uid}
	const q = `query Ingredients($id: string){
		ingredients (func: uid($id)) {
			uid
			name
			quantity
		}
	}`

	resp, err := txn.QueryWithVars(ctx, q, vars)
	if err != nil {
		return nil, err
	}

	var jres struct {
		Ingredients []*Ingredient `json:"ingredients"`
	}

	if err := json.Unmarshal(resp.GetJson(), &jres); err != nil {
		return nil, err
	}

	return jres.Ingredients[0], nil
}

func getStep(ctx context.Context, uid string) (*Step, error) {
	c, err := newClient()
	if err != nil {
		return nil, err
	}

	txn := c.NewReadOnlyTxn()

	vars := map[string]string{"$id": uid}
	const q = `query Steps($id: string){
		steps (func: uid($id)) {
			uid
			index
			excerpt
			text
		}
	}`

	resp, err := txn.QueryWithVars(ctx, q, vars)
	if err != nil {
		return nil, err
	}

	var jres struct {
		Steps []*Step `json:"steps"`
	}

	if err := json.Unmarshal(resp.GetJson(), &jres); err != nil {
		return nil, err
	}

	return jres.Steps[0], nil
}

func newClient() (*dgo.Dgraph, error) {
	// Dial a gRPC connection. The address to dial to can be configured when
	// setting up the dgraph cluster.
	d, err := grpc.Dial("localhost:9080", grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return dgo.NewDgraphClient(
		api.NewDgraphClient(d),
	), nil
}
