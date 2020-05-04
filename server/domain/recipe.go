package domain

import (
	"strings"

	"github.com/kind84/gospiga/pkg/types"
)

type Recipe struct {
	ID          string           `json:"uid,omitempty"`
	ExternalID  string           `json:"xid,omitempty"`
	Title       string           `json:"title,omitempty"`
	Subtitle    string           `json:"subtitle,omitempty"`
	MainImage   *Image           `json:"mainImage,omitempty"`
	Likes       int              `json:"likes,omitempty"`
	Difficulty  RecipeDifficulty `json:"difficulty,omitempty"`
	Cost        RecipeCost       `json:"cost,omitempty"`
	PrepTime    int              `json:"prepTime,omitempty"`
	CookTime    int              `json:"cookTime,omitempty"`
	Servings    int              `json:"servings,omitempty"`
	ExtraNotes  string           `json:"extraNotes,omitempty"`
	Description string           `json:"description,omitempty"`
	Ingredients []*Ingredient    `json:"ingredients,omitempty"`
	Steps       []*Step          `json:"steps,omitempty"`
	Tags        []*Tag           `json:"tags,omitempty"`
	Conclusion  string           `json:"conclusion,omitempty"`
	Slug        string           `json:"slug,omitempty"`
}

type RecipeDifficulty string

const (
	DifficultyEasy = "Bassa"
	DifficultyMid  = "Media"
	DifficultyHard = "Alta"
)

type RecipeCost string

const (
	CostLow  = "Basso"
	CostMid  = "Medio"
	CostHigh = "Alto"
)

type Ingredient struct {
	Name          string      `json:"name,omitempty"`
	Quantity      interface{} `json:"quantity,omitempty"`
	UnitOfMeasure string      `json:"unitOfMeasure,omitempty"`
}

type Step struct {
	Heading string `json:"heading,omitempty"`
	Body    string `json:"body,omitempty"`
	Image   *Image `json:"image,omitempty"`
}

type Image struct {
	URL string `json:"url,omitempty"`
}

type Tag struct {
	TagName string    `json:"tagName,omitempty"`
	Recipes []*Recipe `json:"recipes,omitempty"`
}

func (r *Recipe) ToType() *types.Recipe {
	var rt types.Recipe
	var img types.Image
	if r.MainImage != nil {
		img.URL = r.MainImage.URL
	}

	rt.ID = r.ID
	rt.ExternalID = r.ExternalID
	rt.Title = r.Title
	rt.Subtitle = r.Subtitle
	rt.MainImage = &img
	rt.Likes = r.Likes
	rt.Description = r.Description
	rt.Conclusion = r.Conclusion
	rt.Difficulty = types.RecipeDifficulty(r.Difficulty)
	rt.Cost = types.RecipeCost(r.Cost)
	rt.PrepTime = r.PrepTime
	rt.CookTime = r.CookTime
	rt.Servings = r.Servings
	rt.ExtraNotes = r.ExtraNotes
	rt.Slug = r.Slug

	for _, ingr := range r.Ingredients {
		rt.Ingredients = append(rt.Ingredients, &types.Ingredient{
			Name:          ingr.Name,
			Quantity:      ingr.Quantity,
			UnitOfMeasure: ingr.UnitOfMeasure,
		})
	}

	for _, step := range r.Steps {
		var img *types.Image
		if step.Image != nil {
			img = &types.Image{URL: step.Image.URL}
		}
		rt.Steps = append(rt.Steps, &types.Step{
			Heading: step.Heading,
			Body:    step.Body,
			Image:   img,
		})
	}

	tags := make([]string, 0, len(r.Tags))
	for _, t := range r.Tags {
		tags = append(tags, t.TagName)
	}
	rt.Tags = strings.Join(tags, ", ")

	return &rt
}

func FromType(rt *types.Recipe) *Recipe {
	var r Recipe

	r.ExternalID = rt.ExternalID
	r.Title = strings.TrimSpace(rt.Title)
	r.Subtitle = strings.TrimSpace(rt.Subtitle)
	r.MainImage = &Image{URL: rt.MainImage.URL}
	r.Likes = rt.Likes
	r.Description = rt.Description
	r.Conclusion = rt.Conclusion
	r.Difficulty = RecipeDifficulty(rt.Difficulty)
	r.Cost = RecipeCost(rt.Cost)
	r.PrepTime = rt.PrepTime
	r.CookTime = rt.CookTime
	r.Servings = rt.Servings
	r.ExtraNotes = rt.ExtraNotes
	r.Slug = rt.Slug

	for _, ingr := range rt.Ingredients {
		r.Ingredients = append(r.Ingredients, &Ingredient{
			Name:          strings.ToLower(strings.TrimSpace(ingr.Name)),
			Quantity:      ingr.Quantity,
			UnitOfMeasure: strings.ToLower(strings.TrimSpace(ingr.UnitOfMeasure)),
		})
	}

	for _, step := range rt.Steps {
		var img *Image
		if step.Image != nil {
			img = &Image{URL: step.Image.URL}
		}
		r.Steps = append(r.Steps, &Step{
			Heading: strings.TrimSpace(step.Heading),
			Body:    step.Body,
			Image:   img,
		})
	}

	tags := strings.Split(rt.Tags, ", ")

	for _, tag := range tags {
		r.Tags = append(r.Tags, &Tag{TagName: strings.ToLower(strings.TrimSpace(tag))})
	}

	return &r
}

func (t *Tag) ToType() *types.Tag {
	recipes := make([]*types.Recipe, 0, len(t.Recipes))
	for _, r := range t.Recipes {
		recipes = append(recipes, r.ToType())
	}

	return &types.Tag{
		TagName: t.TagName,
		Recipes: recipes,
	}
}

func NewRecipe(id string, title string) (*Recipe, error) {
	r := &Recipe{
		ID:    id,
		Title: title,
	}
	return r, nil
}

func (r *Recipe) Hello() string {
	return r.Title
}
