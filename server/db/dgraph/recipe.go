package dgraph

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/dgraph-io/dgo/v2/protos/api"

	"github.com/kind84/gospiga/server/domain"
)

type Recipe struct {
	domain.Recipe
	DType      []string  `json:"dgraph.type,omitempty"`
	CretedAt   time.Time `json:"createdAt,omitempty"`
	ModifiedAt time.Time `json:"modifiedAt,omitempty"`
}

func (r Recipe) MarshalJSON() ([]byte, error) {
	type Alias Recipe
	if len(r.DType) == 0 {
		r.DType = []string{"Recipe"}
	}
	return json.Marshal((Alias)(r))
}

// SaveRecipe on disk if a recipe with the same external ID is not already
// present.
func (db *DB) SaveRecipe(ctx context.Context, recipe *domain.Recipe) error {
	dRecipe, err := db.getRecipeByID(ctx, recipe.ExternalID)
	if err != nil {
		return err
	}
	if dRecipe != nil {
		return nil
	}

	mu := &api.Mutation{CommitNow: true}

	now := time.Now()
	dRecipe = &Recipe{*recipe, []string{}, now, now}
	dRecipe.ID = "_:recipe"

	rb, err := json.Marshal(dRecipe)
	if err != nil {
		return err
	}

	mu.SetJson = rb
	res, err := db.Dgraph.NewTxn().Mutate(ctx, mu)
	if err != nil {
		return err
	}
	ruid := res.Uids["recipe"]
	recipe.ID = ruid
	return nil
}

// UpdateRecipe matching the external ID. ModifiedAt field gets updated.
func (db *DB) UpdateRecipe(ctx context.Context, recipe *domain.Recipe) error {
	dRecipe, err := db.getRecipeByID(ctx, recipe.ExternalID)
	if err != nil {
		return err
	}
	if dRecipe == nil {
		return fmt.Errorf("recipe external ID [%s] not found", recipe.ExternalID)
	}

	mu := &api.Mutation{CommitNow: true}

	dRecipe.ModifiedAt = time.Now()

	rb, err := json.Marshal(dRecipe)
	if err != nil {
		return err
	}

	mu.SetJson = rb
	_, err = db.Dgraph.NewTxn().Mutate(ctx, mu)
	if err != nil {
		return err
	}
	return nil
}

// DeleteRecipe matching the given external ID.
func (db *DB) DeleteRecipe(ctx context.Context, recipeID string) error {
	req := &api.Request{CommitNow: true}
	req.Vars = map[string]string{"$id": recipeID}
	req.Query = `
		query RecipeUID($id: string){
			recipeUID(func: eq(id, $id)) {
				...fragmentA
			}
		}

		fragment fragmentA {
			v as uid
		}
	`
	del := map[string]string{"uid": "uid(v)"}
	pb, err := json.Marshal(del)
	if err != nil {
		return err
	}
	mu := &api.Mutation{
		DeleteJson: pb,
	}
	req.Mutations = []*api.Mutation{mu}

	_, err = db.Dgraph.NewTxn().Do(ctx, req)

	return err
}

// GetRecipeByID and return the domain recipe matching the external ID.
func (db *DB) GetRecipeByID(ctx context.Context, id string) (*domain.Recipe, error) {
	dRecipe, err := db.getRecipeByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if dRecipe == nil {
		return nil, nil
	}
	return &dRecipe.Recipe, nil
}

func (db *DB) getRecipeByID(ctx context.Context, id string) (*Recipe, error) {
	vars := map[string]string{"$id": id}
	q := `
		query IDSaved($id: string){
			recipes(func: eq(id, $id)) {
				expand(_all_)
			}
		}
	`

	resp, err := db.Dgraph.NewTxn().QueryWithVars(ctx, q, vars)
	if err != nil {
		return nil, err
	}

	var root struct {
		Recipes []Recipe `json:"recipes"`
	}
	err = json.Unmarshal(resp.Json, &root)
	if err != nil {
		return nil, err
	}
	if len(root.Recipes) == 0 {
		return nil, nil
	}
	return &root.Recipes[0], nil
}

// GetRecipesByUIDs and return domain recipes.
func (db *DB) GetRecipesByUIDs(ctx context.Context, uids []string) ([]*domain.Recipe, error) {
	uu := strings.Join(uids, ", ")
	vars := map[string]string{"$uids": uu}
	q := `
		query IDSaved($uid: []string){
			recipes(func: uid($uids)) {
				expand(_all_)
			}
		}
	`

	resp, err := db.Dgraph.NewTxn().QueryWithVars(ctx, q, vars)
	if err != nil {
		return nil, err
	}

	var root struct {
		Recipes []Recipe `json:"recipes"`
	}
	err = json.Unmarshal(resp.Json, &root)
	if err != nil {
		return nil, err
	}
	if len(root.Recipes) == 0 {
		return nil, nil
	}

	recipes := make([]*domain.Recipe, 0, len(root.Recipes))
	for _, recipe := range root.Recipes {
		recipes = append(recipes, &recipe.Recipe)
	}
	return recipes, nil
}

// IDSaved check if the given external ID is stored.
func (db *DB) IDSaved(ctx context.Context, id string) (bool, error) {
	vars := map[string]string{"$id": id}
	q := `
		query IDSaved($id: string){
			recipes(func: eq(id, $id)) {
				uid
			}
		}
	`

	resp, err := db.Dgraph.NewTxn().QueryWithVars(ctx, q, vars)
	if err != nil {
		return false, err
	}

	var root struct {
		Recipes []Recipe `json:"recipes"`
	}
	err = json.Unmarshal(resp.Json, &root)
	if err != nil {
		return false, err
	}
	return len(root.Recipes) > 0, nil
}

func loadRecipeSchema() *api.Operation {
	op := &api.Operation{}
	op.Schema = `
		type Recipe {
			id
			title
			subtitle
			mainImage
			likes
			difficulty
			cost
			prepTime
			cookTime
			servings
			extraNotes
			description
			ingredients
			steps
			conclusion
			createdAt
			modifiedAt
		}

		type Ingredient {
			name
			quantity
			unitOfMeasure
		}

		type Step {
			title
			description
			image
		}

		type Image {
			url
		}

		id: string @index(exact) .
		title: string @lang @index(fulltext) .
		subtitle: string @lang @index(fulltext) .
		mainImage: uid .
		likes: int @index(int) .
		difficulty: string .
		cost: string .
		prepTime: int @index(int) .
		cookTime: int @index(int) .
		servings: int .
		extraNotes: string .
		description: string @lang @index(fulltext) .
		ingredients: [uid] @count @reverse .
		steps: [uid] @count .
		conclusion: string .
		name: string @lang @index(term) .
		quantity: string .
		unitOfMeasure: string .
		image: uid .
		url: string .
		createdAt: dateTime @index(day).
		modifiedAt: dateTime @index(day).
	`
	return op
}
