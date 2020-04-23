package dgraph

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/dgraph-io/dgo/v2/protos/api"

	"github.com/kind84/gospiga/pkg/errors"
	"github.com/kind84/gospiga/pkg/stemmer"
	"github.com/kind84/gospiga/server/domain"
)

// TODO: add dgraph type on ingredients and steps.

// recipe represents repository version of the domain recipe.
type Recipe struct {
	ID          string                  `json:"uid,omitempty"`
	ExternalID  string                  `json:"xid,omitempty"`
	Title       string                  `json:"title,omitempty"`
	Subtitle    string                  `json:"subtitle,omitempty"`
	MainImage   *Image                  `json:"mainImage,omitempty"`
	Likes       int                     `json:"likes,omitempty"`
	Difficulty  domain.RecipeDifficulty `json:"difficulty,omitempty"`
	Cost        domain.RecipeCost       `json:"cost,omitempty"`
	PrepTime    int                     `json:"prepTime,omitempty"`
	CookTime    int                     `json:"cookTime,omitempty"`
	Servings    int                     `json:"servings,omitempty"`
	ExtraNotes  string                  `json:"extraNotes,omitempty"`
	Description string                  `json:"description,omitempty"`
	Ingredients []*Ingredient           `json:"ingredients,omitempty"`
	Steps       []*Step                 `json:"steps,omitempty"`
	Tags        []*Tag                  `json:"tags,omitempty"`
	Conclusion  string                  `json:"conclusion,omitempty"`
	Slug        string                  `json:"slug,omitempty"`
	DType       []string                `json:"dgraph.type,omitempty"`
	CretedAt    *time.Time              `json:"createdAt,omitempty"`
	ModifiedAt  *time.Time              `json:"modifiedAt,omitempty"`
}

func (r Recipe) MarshalJSON() ([]byte, error) {
	type Alias Recipe
	if len(r.DType) == 0 {
		r.DType = []string{"Recipe"}
	}
	return json.Marshal((Alias)(r))
}

// step represents repository version of the domain step.
type Step struct {
	Heading string   `json:"heading,omitempty"`
	Body    string   `json:"body,omitempty"`
	Image   *Image   `json:"image,omitempty"`
	DType   []string `json:"dgraph.type,omitempty"`
}

func (s Step) MarshalJSON() ([]byte, error) {
	type Alias Step
	if len(s.DType) == 0 {
		s.DType = []string{"Step"}
	}
	return json.Marshal((Alias)(s))
}

// image represents repository version of the domain image.
type Image struct {
	domain.Image
	DType []string `json:"dgraph.type,omitempty"`
}

func (i Image) MarshalJSON() ([]byte, error) {
	type Alias Image
	if len(i.DType) == 0 {
		i.DType = []string{"Image"}
	}
	return json.Marshal((Alias)(i))
}

func (r *Recipe) ToDomain() *domain.Recipe {
	ings := make([]*domain.Ingredient, 0, len(r.Ingredients))
	for _, i := range r.Ingredients {
		ings = append(ings, &i.Ingredient)
	}
	steps := make([]*domain.Step, 0, len(r.Steps))
	for _, s := range r.Steps {
		var i *domain.Image
		if &s.Image.Image != nil {
			i = &s.Image.Image
		}
		steps = append(steps, &domain.Step{
			Heading: s.Heading,
			Body:    s.Body,
			Image:   i,
		})
	}
	tags := make([]*domain.Tag, 0, len(r.Tags))
	for _, t := range r.Tags {
		tags = append(tags, &t.Tag)
	}

	return &domain.Recipe{
		ID:          r.ID,
		ExternalID:  r.ExternalID,
		Title:       r.Title,
		Subtitle:    r.Subtitle,
		MainImage:   &r.MainImage.Image,
		Likes:       r.Likes,
		Difficulty:  r.Difficulty,
		Cost:        r.Cost,
		PrepTime:    r.PrepTime,
		CookTime:    r.CookTime,
		Servings:    r.Servings,
		ExtraNotes:  r.ExtraNotes,
		Description: r.Description,
		Ingredients: ings,
		Steps:       steps,
		Conclusion:  r.Conclusion,
		Tags:        tags,
		Slug:        r.Slug,
	}
}

func FromDomain(r *domain.Recipe) (*Recipe, error) {
	ings := make([]*Ingredient, 0, len(r.Ingredients))
	for _, i := range r.Ingredients {
		s, err := stemmer.Stem(i.Name, "italian")
		if err != nil {
			return nil, err
		}
		ings = append(ings, &Ingredient{
			Ingredient: *i,
			Food: &Food{
				Term:  i.Name,
				Stem:  s,
				DType: []string{"Food"},
			},
			DType: []string{"Ingredient"},
		})
	}
	steps := make([]*Step, 0, len(r.Steps))
	for _, s := range r.Steps {
		var i *Image
		if s.Image != nil {
			i = &Image{
				Image: *s.Image,
				DType: []string{"Image"},
			}
		}
		steps = append(steps, &Step{
			Heading: s.Heading,
			Body:    s.Body,
			Image:   i,
			DType:   []string{"Step"},
		})
	}
	tags := make([]*Tag, 0, len(r.Tags))
	for _, t := range r.Tags {
		tags = append(tags, &Tag{
			Tag:   *t,
			DType: []string{"Tag"},
		})
	}

	now := time.Now()
	dr := &Recipe{
		ExternalID: r.ExternalID,
		Title:      r.Title,
		Subtitle:   r.Subtitle,
		MainImage: &Image{
			Image: *r.MainImage,
			DType: []string{"Image"},
		},
		Likes:       r.Likes,
		Difficulty:  r.Difficulty,
		Cost:        r.Cost,
		PrepTime:    r.PrepTime,
		CookTime:    r.CookTime,
		Servings:    r.Servings,
		ExtraNotes:  r.ExtraNotes,
		Description: r.Description,
		Ingredients: ings,
		Steps:       steps,
		Conclusion:  r.Conclusion,
		Tags:        tags,
		Slug:        r.Slug,
		DType:       []string{"Recipe"},
		CretedAt:    &now,
		ModifiedAt:  &now,
	}

	return dr, nil
}

// CountRecipes total number.
func (db *DB) CountRecipes(ctx context.Context) (int, error) {
	return db.count(ctx, "Recipe")
}

// SaveRecipe if a recipe with the same external ID has not been saved yet.
func (db *DB) SaveRecipe(ctx context.Context, r *domain.Recipe) error {
	req := &api.Request{CommitNow: true}
	req.Vars = map[string]string{"$xid": r.ExternalID}
	req.Query = `
		query RecipeUID($xid: string){
			recipeUID(func: eq(xid, $xid)) {
				v as uid
			}
		}
	`
	dRecipe, err := FromDomain(r)
	if err != nil {
		return err
	}
	dRecipe.ID = "_:recipe"

	rb, err := json.Marshal(dRecipe)
	if err != nil {
		return err
	}

	mu := &api.Mutation{
		SetJson: rb,
		Cond:    "@if(eq(len(v), 0))",
	}

	req.Mutations = []*api.Mutation{mu}

	res, err := db.Dgraph.NewTxn().Do(ctx, req)
	if err != nil {
		return err
	}

	if ruid, created := res.Uids["recipe"]; created {
		r.ID = ruid
	} else {
		return errors.ErrDuplicateID{ID: r.ExternalID}
	}

	return nil
}

// UpsertRecipe, if a recipe with the same external ID is already present it
// gets replaced with the given recipe.
func (db *DB) UpsertRecipe(ctx context.Context, recipe *domain.Recipe) error {
	req := &api.Request{CommitNow: true}
	req.Vars = map[string]string{"$xid": recipe.ExternalID}
	req.Query = `
		query RecipeUID($xid: string){
			recipeUID(func: eq(xid, $xid)) {
				v as uid
				c as createdAt
			}
		}
	`
	dRecipe, err := FromDomain(recipe)
	if err != nil {
		return err
	}
	dRecipe.ID = "uid:uid(v)"

	rb, err := json.Marshal(dRecipe)
	if err != nil {
		return err
	}
	var m map[string]interface{}
	err = json.Unmarshal(rb, &m)
	if err != nil {
		return err
	}
	m["createdAt"] = "val(c)"
	rb, err = json.Marshal(m)
	if err != nil {
		return err
	}

	del := map[string]string{"uid": "uid(v)"}
	delb, err := json.Marshal(del)
	if err != nil {
		return err
	}
	mu := &api.Mutation{
		SetJson:    rb,
		DeleteJson: delb,
		Cond:       "@if(gt(len(v), 0))",
	}

	req.Mutations = []*api.Mutation{mu}

	res, err := db.Dgraph.NewTxn().Do(ctx, req)
	if err != nil {
		return err
	}

	if ruid, created := res.Uids["recipe"]; created {
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
				v as uid
			}
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
	r, err := db.getRecipeByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, nil
	}

	ings := make([]*domain.Ingredient, 0, len(r.Ingredients))
	for _, i := range r.Ingredients {
		ings = append(ings, &i.Ingredient)
	}
	steps := make([]*domain.Step, 0, len(r.Steps))
	for _, s := range r.Steps {
		steps = append(steps, &domain.Step{
			Heading: s.Heading,
			Body:    s.Body,
			Image:   &s.Image.Image,
		})
	}

	return r.ToDomain(), nil
}

func (db *DB) getRecipeByID(ctx context.Context, id string) (*Recipe, error) {
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
				tags {
					tagName
				}
				conclusion
				slug
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
	for _, r := range root.Recipes {
		recipes = append(recipes, r.ToDomain())
	}
	return recipes, nil
}

// IDSaved check if the given external ID is stored.
func (db *DB) IDSaved(ctx context.Context, id string) (bool, error) {
	vars := map[string]string{"$id": id}
	q := `
		query IDSaved($id: string){
			recipes(func: eq(xid, $id)) {
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
			slug
			createdAt
			modifiedAt
		}

		type Ingredient {
			name
			quantity
			unitOfMeasure
			food
			<~ingredients>
		}

		type Food {
			term
			stem
			<~food>
		}

		type Step {
			index
			heading
			body
			image
		}

		type Image {
			url
		}

		type Tag {
			tagName
			<~tags>
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
		heading: string @lang @index(fulltext) .
		body: string @lang @index(fulltext) .
		conclusion: string .
		finalImage: uid .
		tags: [uid] @reverse .
		name: string @lang @index(term) .
		quantity: string .
		unitOfMeasure: string .
		food: uid @reverse .
		term: string @index(term) .
		stem: string .
		index: int @index(int) .
		image: uid .
		url: string .
		createdAt: dateTime @index(hour) @upsert .
		modifiedAt: dateTime @index(hour) @upsert .
		tagName: string @index(term) .
		slug: string .
	`
	return op
}
