//go:generate go run github.com/99designs/gqlgen

package gospiga

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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

func (r *queryResolver) Recipes(ctx context.Context, uid *string, tags []*string) ([]*models.Recipe, error) {
	c, err := newClient()
	if err != nil {
		return nil, err
	}

	txn := c.NewReadOnlyTxn()
	var fn, d, tg string

	if uid != nil {
		fn = fmt.Sprintf("uid(%s)", *uid)
	} else {
		fn = "has(title)"
	}
	if len(tags) > 0 {
		d = "@cascade"
		var ts []interface{}
		for _, t := range tags {
			ts = append(ts, *t)
		}
		tg = fmt.Sprintf(`tag @filter(anyofterms(name, %s)) {
			uid
		}`, ts...)
	} else {
		tg = `tag {
			uid
		}`
	}

	q := fmt.Sprintf(`{
		recipes (func: %s) %s{
			uid
			title
			subtitle
			description
			ingredient {
				uid
			}
			step {
				uid
			}
			conclusion
			%s
			createdAt
			updatedAt
		}
	}`, fn, d, tg)

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
			title: string @index(fulltext) .
			subtitle: string @index(fulltext) .
			description: string @index(fulltext) .
			ingredient: uid @reverse .
			step: uid @reverse .
			conclusion: string .
			tag: uid @reverse .
			createdAt: dateTime @index(day) .

			name: string @index(term) .
			quantity: int .

			index: int .
			excerpt: string @index(fulltext) .
			text: string @index(fulltext) .
		`,
	})
	if err != nil {
		return nil, err
	}

	mu := &api.Mutation{CommitNow: true}

	rcp := struct {
		NewRecipe
		CreatedAt time.Time `json:"createdAt,omitempty"`
		UpdatedAt time.Time `json:"updatedAt,omitempty"`
	}{
		NewRecipe: nr,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	rb, err := json.Marshal(rcp)
	if err != nil {
		return nil, err
	}

	mu.SetJson = rb
	assigned, err := c.NewTxn().Mutate(ctx, mu)
	if err != nil {
		return nil, err
	}

	res, err := getRecipe(ctx, assigned.Uids["blank-0"])
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (r *mutationResolver) UpdateRecipe(ctx context.Context, input UpRecipe) (*models.Recipe, error) {
	c, err := newClient()
	if err != nil {
		return nil, err
	}

	txn := c.NewTxn()
	variables := map[string]string{"$id": input.UID}
	q := `query Recipe($id: string){
		recipe(func: uid($id)) {
			uid
			title
			subtitle
			description
			ingredient {
				uid
			}
			step {
				uid
			}
			conclusion
			tag {
				uid
			}
			createdAt
			updatedAt
		}
	}`

	resp, err := txn.QueryWithVars(ctx, q, variables)
	if err != nil {
		return nil, err
	}

	type Root struct {
		Recipe []models.Recipe `json:"recipe"`
	}

	var rt Root
	err = json.Unmarshal(resp.Json, &rt)
	if err != nil {
		return nil, err
	}
	if len(rt.Recipe) == 0 {
		errMsg := fmt.Sprintf("Recipe UID %s not found.", input.UID)
		return nil, errors.New(errMsg)
	}

	upRc := struct {
		UpRecipe
		CreatedAt time.Time `json:"createdAt,omitempty"`
		UpdatedAt time.Time `json:"updatedAt,omitempty"`
	}{
		UpRecipe:  input,
		CreatedAt: rt.Recipe[0].CreatedAt,
		UpdatedAt: time.Now().UTC(),
	}

	rb, err := json.Marshal(upRc)
	if err != nil {
		return nil, err
	}

	mu := &api.Mutation{}

	mu.SetJson = rb
	_, err = txn.Mutate(ctx, mu)
	if err != nil {
		return nil, err
	}

	if err = txn.Commit(ctx); err != nil {
		return nil, err
	}

	res, err := getRecipe(ctx, rt.Recipe[0].UID)
	if err != nil {
		return nil, err
	}
	return res, nil
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

func (r *recipeResolver) Tag(ctx context.Context, obj *models.Recipe) ([]*Tag, error) {
	var tags []*Tag

	// Dgraph does not allow to pass muliple UIDs as func variable. Looping.
	for _, t := range obj.Tag {
		tt, err := getTag(ctx, t.UID)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tt)
	}
	return tags, nil
}

func getRecipe(ctx context.Context, uid string) (*models.Recipe, error) {
	c, err := newClient()
	if err != nil {
		return nil, err
	}

	variables := map[string]string{"$id": uid}
	q := `query Recipe($id: string){
		recipe(func: uid($id)) {
			uid
			title
			subtitle
			description
			ingredient {
				uid
			}
			step {
				uid
			}
			conclusion
			tag {
				uid
			}
			createdAt
			updatedAt
		}
	}`

	resp, err := c.NewTxn().QueryWithVars(ctx, q, variables)
	if err != nil {
		return nil, err
	}

	type Root struct {
		Recipe []models.Recipe `json:"recipe"`
	}

	var rt Root
	err = json.Unmarshal(resp.Json, &rt)
	if err != nil {
		return nil, err
	}

	return &rt.Recipe[0], nil
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

func getTag(ctx context.Context, uid string) (*Tag, error) {
	c, err := newClient()
	if err != nil {
		return nil, err
	}

	txn := c.NewReadOnlyTxn()

	vars := map[string]string{"$id": uid}
	const q = `query Tags($id: string){
		tags (func: uid($id)) {
			uid
			name
		}
	}`

	resp, err := txn.QueryWithVars(ctx, q, vars)
	if err != nil {
		return nil, err
	}

	var jres struct {
		Tags []*Tag `json:"tags"`
	}

	if err := json.Unmarshal(resp.GetJson(), &jres); err != nil {
		return nil, err
	}

	return jres.Tags[0], nil
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
