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

// TODO: add dgraph type on ingredients and steps.

// recipe represents repository version of the domain recipe.
type recipe struct {
	domain.Recipe
	DType      []string   `json:"dgraph.type,omitempty"`
	CretedAt   *time.Time `json:"createdAt,omitempty"`
	ModifiedAt *time.Time `json:"modifiedAt,omitempty"`
}

func (r recipe) MarshalJSON() ([]byte, error) {
	type Alias recipe
	if len(r.DType) == 0 {
		r.DType = []string{"Recipe"}
	}
	return json.Marshal((Alias)(r))
}

// Ingredient represents repository verison of the domain ingredient.
type ingredient struct {
	domain.Ingredient
	DType []string `json:"dgraph.type,omitempty"`
}

func (i ingredient) MarshalJSON() ([]byte, error) {
	type Alias ingredient
	if len(i.DType) == 0 {
		i.DType = []string{"Ingredient"}
	}
	return json.Marshal((Alias)(i))
}

func newRecipe(r *domain.Recipe) *recipe {
	// ings := make([]ingredient, 0, len(r.Ingredients))
	// for _, i := range r.Ingredients {
	// 	ings = append(ings, ingredient{*i, []string{}})
	// }
	now := time.Now()
	return &recipe{*r, []string{}, &now, &now}
}

// CountRecipes total number.
func (db *DB) CountRecipes(ctx context.Context) (int, error) {
	return db.count(ctx, "Recipe")
}

// SaveRecipe on disk with upsert. If a recipe with the same external ID is
// already present it gets replaced with the given recipe.
func (db *DB) SaveRecipe(ctx context.Context, recipe *domain.Recipe) error {
	req := &api.Request{CommitNow: true}
	req.Vars = map[string]string{"$xid": recipe.ExternalID}
	req.Query = `
		query RecipeUID($xid: string){
			recipeUID(func: eq(xid, $xid)) {
				...fragmentA
			}
		}

		fragment fragmentA {
			v as uid
			c as createdAt
		}
	`
	now := time.Now()
	dRecipe := newRecipe(recipe)
	dRecipe.ID = "uid(v)"

	pb, err := json.Marshal(dRecipe)
	if err != nil {
		return err
	}

	mu := &api.Mutation{
		SetJson: pb,
	}

	mm := map[string]interface{}{
		"cond": fmt.Sprintf("@if(NOT(lt(val(c), %s)))", now.Format(time.RFC3339Nano)),
		"set": map[string]interface{}{
			"uid":       "uid(v)",
			"createdAt": "val(c)",
		},
	}
	pb2, err := json.Marshal(mm)
	if err != nil {
		return err
	}
	mu2 := &api.Mutation{
		SetJson: pb2,
	}

	req.Mutations = []*api.Mutation{mu, mu2}

	res, err := db.Dgraph.NewTxn().Do(ctx, req)
	if err != nil {
		return err
	}

	if ruid, created := res.Uids["uid(v)"]; created && recipe.ID == "" {
		recipe.ID = ruid
	}
	return nil
}

// DeleteRecipe matching the given external ID.
func (db *DB) DeleteRecipe(ctx context.Context, recipeID string) error {
	req := &api.Request{CommitNow: true}
	req.Vars = map[string]string{"$xid": recipeID}
	req.Query = `
		query RecipeUID($xid: string){
			recipeUID(func: eq(xid, $xid)) {
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

func (db *DB) getRecipeByID(ctx context.Context, id string) (*recipe, error) {
	vars := map[string]string{"$xid": id}
	q := `
		query Recipes($xid: string){
			recipes(func: eq(xid, $xid)) {
				expand(_all_)
			}
		}
	`

	resp, err := db.Dgraph.NewReadOnlyTxn().QueryWithVars(ctx, q, vars)
	if err != nil {
		return nil, err
	}

	var root struct {
		Recipes []recipe `json:"recipes"`
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
		query Recipes($uids: string){
			recipes(func: uid($uids)) {
				id
				title
				subtitle
				mainImage {
					url
				}
				likes
				difficulty
				cost
				prepTime
				cookTime
				servings
				extraNotes
				description
				ingredients {
					name
					quantity
					unitOfMeasure
				}
				steps {
					title
					description
					image {
						url
					}
				}
				conclusion
			}
		}
	`

	resp, err := db.Dgraph.NewTxn().QueryWithVars(ctx, q, vars)
	if err != nil {
		return nil, err
	}

	var root struct {
		Recipes []recipe `json:"recipes"`
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
		Recipes []recipe `json:"recipes"`
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
			xid
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
			tags
			finalImage
			createdAt
			modifiedAt
		}

		type Ingredient {
			name
			quantity
			unitOfMeasure
			<~ingredients>
		}

		type Step {
			index
			title
			description
			image
		}

		type Image {
			url
		}

		type Tag {
			tag
		}

		xid: string @index(hash) .
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
		finalImage: uid .
		tags: [uid] .
		name: string @lang @index(term) .
		quantity: string .
		unitOfMeasure: string .
		index: int @index(int) .
		image: uid .
		url: string .
		createdAt: dateTime @index(hour) @upsert .
		modifiedAt: dateTime @index(hour) @upsert .
		tag: string @index(term) .
	`
	return op
}
