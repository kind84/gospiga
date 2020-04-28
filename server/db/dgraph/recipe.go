package dgraph

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/dgraph-io/dgo/v2/protos/api"

	"github.com/kind84/gospiga/pkg/errors"
	"github.com/kind84/gospiga/pkg/stemmer"
	"github.com/kind84/gospiga/server/domain"
)

var fm = template.FuncMap{
	"StemFood": func(term string) string {
		s, _ := stemmer.Stem(term, "italian")
		return s
	},
}

// Recipe represents repository version of the domain recipe.
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

// Step represents repository version of the domain step.
type Step struct {
	ID      string   `json:"uid,omitempty"`
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

// Image represents repository version of the domain image.
type Image struct {
	ID    string   `json:"uid,omitempty"`
	URL   string   `json:"url,omitempty"`
	DType []string `json:"dgraph.type,omitempty"`
}

func (i Image) MarshalJSON() ([]byte, error) {
	type Alias Image
	if len(i.DType) == 0 {
		i.DType = []string{"Image"}
	}
	return json.Marshal((Alias)(i))
}

// ToDomain converts a dgraph recipe into a domain recipe.
func (r *Recipe) ToDomain() *domain.Recipe {
	ings := make([]*domain.Ingredient, 0, len(r.Ingredients))
	for _, i := range r.Ingredients {
		ings = append(ings, i.ToDomain())
	}
	steps := make([]*domain.Step, 0, len(r.Steps))
	for _, s := range r.Steps {
		var i domain.Image
		if s.Image != nil {
			i.URL = s.Image.URL
		}
		steps = append(steps, &domain.Step{
			Heading: s.Heading,
			Body:    s.Body,
			Image:   &i,
		})
	}
	tags := make([]*domain.Tag, 0, len(r.Tags))
	for _, t := range r.Tags {
		dt := &domain.Tag{TagName: t.TagName}
		dt.Recipes = make([]*domain.Recipe, 0, len(t.Recipes))
		for _, r := range t.Recipes {
			dt.Recipes = append(dt.Recipes, r.ToDomain()) // /!\ recursive
		}
		tags = append(tags)
	}

	dr := &domain.Recipe{
		ID:          r.ID,
		ExternalID:  r.ExternalID,
		Title:       r.Title,
		Subtitle:    r.Subtitle,
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

	var mi domain.Image
	if r.MainImage != nil {
		mi.URL = r.MainImage.URL
	}

	return dr
}

// FromDomain converts a domain recipe into a dgraph recipe.
func (r *Recipe) FromDomain(dr *domain.Recipe) error {
	ings := make([]*Ingredient, 0, len(dr.Ingredients))
	for _, di := range dr.Ingredients {
		var i Ingredient
		err := i.FromDomain(di)
		if err != nil {
			return err
		}
		ings = append(ings, &i)
	}
	steps := make([]*Step, 0, len(dr.Steps))
	for _, s := range dr.Steps {
		var i Image
		if s.Image != nil {
			i = Image{
				URL:   s.Image.URL,
				DType: []string{"Image"},
			}
		}
		steps = append(steps, &Step{
			Heading: s.Heading,
			Body:    s.Body,
			Image:   &i,
			DType:   []string{"Step"},
		})
	}
	tags := make([]*Tag, 0, len(dr.Tags))
	for _, t := range dr.Tags {
		tags = append(tags, &Tag{
			TagName: t.TagName,
			DType:   []string{"Tag"},
		})
	}

	r.ExternalID = dr.ExternalID
	r.Title = dr.Title
	r.Subtitle = dr.Subtitle
	if dr.MainImage != nil {
		r.MainImage = &Image{
			URL:   dr.MainImage.URL,
			DType: []string{"Image"},
		}
	}
	r.Likes = dr.Likes
	r.Difficulty = dr.Difficulty
	r.Cost = dr.Cost
	r.PrepTime = dr.PrepTime
	r.CookTime = dr.CookTime
	r.Servings = dr.Servings
	r.ExtraNotes = dr.ExtraNotes
	r.Description = dr.Description
	r.Ingredients = ings
	r.Steps = steps
	r.Conclusion = dr.Conclusion
	r.Tags = tags
	r.Slug = dr.Slug
	r.DType = []string{"Recipe"}

	return nil
}

// CountRecipes total number.
func (db *DB) CountRecipes(ctx context.Context) (int, error) {
	return db.count(ctx, "Recipe")
}

// SaveRecipe if a recipe with the same external ID has not been saved yet.
func (db *DB) SaveRecipe(ctx context.Context, dr *domain.Recipe) error {
	var r Recipe
	err := r.FromDomain(dr)
	if err != nil {
		return err
	}

	var sb strings.Builder
	t := template.Must(template.New("save.tmpl").Funcs(fm).ParseFiles("./save.tmpl"))
	err = t.Execute(&sb, dr)
	if err != nil {
		return err
	}

	req := &api.Request{CommitNow: true}
	req.Vars = map[string]string{"$xid": dr.ExternalID}
	req.Query = sb.String()
	now := time.Now()

	mutations := make([]*api.Mutation, 0, len(dr.Ingredients)+1)

	for i := range dr.Ingredients {
		var r0, r1 Recipe

		err := r0.FromDomain(dr)
		if err != nil {
			return err
		}
		r0.ID = "_:recipe"
		r0.CretedAt = &now
		r0.ModifiedAt = &now

		err = r1.FromDomain(dr)
		if err != nil {
			return err
		}
		r1.ID = "_:recipe"
		r1.CretedAt = &now
		r1.ModifiedAt = &now

		// stem found
		r0.Ingredients[i].Food.ID = fmt.Sprintf("uid(f%d)", i)
		jr0, err := json.Marshal(r0)
		if err != nil {
			return err
		}

		mu0 := &api.Mutation{
			SetJson: jr0,
			Cond:    fmt.Sprintf("@if(eq(len(r), 0) AND eq(len(f%d), 1))", i),
		}

		// stem not found
		r1.Ingredients[i].Food.ID = fmt.Sprintf("_:f%d", i)
		jr1, err := json.Marshal(r1)
		if err != nil {
			return err
		}

		mu1 := &api.Mutation{
			SetJson: jr1,
			Cond:    fmt.Sprintf("@if(eq(len(r), 0) AND eq(len(f%d), 0))", i),
		}

		mutations = append(mutations, mu0, mu1)
	}

	req.Mutations = mutations

	res, err := db.Dgraph.NewTxn().Do(ctx, req)
	if err != nil {
		return err
	}

	if ruid, created := res.Uids["recipe"]; created {
		dr.ID = ruid
	} else {
		return errors.ErrDuplicateID{ID: dr.ExternalID}
	}

	return nil
}

// UpdateRecipe if already stored on db.
func (db *DB) UpdateRecipe(ctx context.Context, dr *domain.Recipe) error {
	var r Recipe
	err := r.FromDomain(dr)
	if err != nil {
		return err
	}

	var sb strings.Builder
	t := template.Must(template.New("update.tmpl").Funcs(fm).ParseFiles("./update.tmpl"))
	err = t.Execute(&sb, dr)
	if err != nil {
		return err
	}

	req := &api.Request{CommitNow: true}
	req.Vars = map[string]string{"$xid": dr.ExternalID}
	req.Query = sb.String()
	now := time.Now()

	mutations := []*api.Mutation{}

	// replace ingredients
	for i := range dr.Ingredients {
		var r0, r1 Recipe
		err := r0.FromDomain(dr)
		if err != nil {
			return err
		}
		err = r1.FromDomain(dr)
		if err != nil {
			return err
		}
		r0.ID = "uid(r)"
		r1.ID = "uid(r)"
		r0.ModifiedAt = &now
		r1.ModifiedAt = &now

		// stem found
		r0.Ingredients[i].Food.ID = fmt.Sprintf("uid(f%d)", i)
		jr0, err := json.Marshal(r0)
		if err != nil {
			return err
		}

		mu0 := &api.Mutation{
			SetJson: jr0,
			Cond:    fmt.Sprintf("@if(eq(len(r), 1) AND eq(len(f%d), 1))", i),
		}

		// stem not found
		r1.Ingredients[i].Food.ID = fmt.Sprintf("_:f%d", i)
		jr1, err := json.Marshal(r1)
		if err != nil {
			return err
		}

		mu1 := &api.Mutation{
			SetJson: jr1,
			Cond:    fmt.Sprintf("@if(eq(len(r), 1) AND eq(len(f%d), 0))", i),
		}

		mutations = append(mutations, mu0, mu1)
	}

	// remove old ingredients
	idel := map[string]interface{}{
		"uid":           "uid(i)",
		"name":          nil,
		"quantity":      nil,
		"unitOfMeasure": nil,
		"food":          nil,
	}
	rdel := map[string]interface{}{
		"uid":         "uid(r)",
		"ingredients": idel,
	}
	jdel, err := json.Marshal(rdel)
	if err != nil {
		return err
	}
	mdel := &api.Mutation{
		DeleteJson: jdel,
		Cond:       "@if(eq(len(r), 1))",
	}
	mutations = append(mutations, mdel)

	req.Mutations = mutations

	_, err = db.Dgraph.NewTxn().Do(ctx, req)
	if err != nil {
		return err
	}

	return nil
}

// DeleteRecipe matching the given external ID.
func (db *DB) DeleteRecipe(ctx context.Context, recipeID string) error {
	r, err := db.getRecipeByID(ctx, recipeID)
	if err != nil {
		return err
	}

	req := &api.Request{CommitNow: true}

	d := make([]interface{}, 0, len(r.Ingredients)+len(r.Steps)+1)
	d = append(d, r)
	for _, i := range r.Ingredients {
		i.Food = nil
		d = append(d, *i)
	}
	for _, s := range r.Steps {
		d = append(d, *s)
	}

	pb, err := json.Marshal(d)
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

	return r.ToDomain(), nil
}

func (db *DB) getRecipeByID(ctx context.Context, id string) (*Recipe, error) {
	vars := map[string]string{"$xid": id}
	q := `
		query Recipes($xid: string){
			recipes(func: eq(xid, $xid)) {
				uid
				xid
				title
				subtitle
				mainImage {
					uid
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
					uid
					name
					quantity
					unitOfMeasure
					food {
						uid
						term
						stem
					}
				}
				steps {
					uid
					heading
					body
					image {
						uid
						url
					}
				}
				tags {
					uid
					tagName
				}
				conclusion
				slug
				createdAt
				modifiedAt
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
				uid
				xid
				title
				subtitle
				mainImage {
					uid
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
					uid
					name
					quantity
					unitOfMeasure
					food {
						uid
						term
						stem
					}
				}
				steps {
					uid
					heading
					body
					image {
						uid
						url
					}
				}
				tags {
					uid
					tagName
				}
				conclusion
				slug
				createdAt
				modifiedAt
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
		stem: string @index(exact) .
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
