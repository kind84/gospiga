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
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Image       *Image `json:"image,omitempty"`
}

type Image struct {
	URL string `json:"url,omitempty"`
}

type Tag struct {
	Tag string `json:"tag,omitempty"`
}

func (r *Recipe) ToType() *types.Recipe {
	var rt types.Recipe

	rt.ExternalID = r.ExternalID
	rt.Title = r.Title
	rt.Subtitle = r.Subtitle
	rt.Likes = r.Likes
	rt.Description = r.Description
	rt.Conclusion = r.Conclusion
	rt.Difficulty = types.RecipeDifficulty(r.Difficulty)
	rt.Cost = types.RecipeCost(r.Cost)
	rt.PrepTime = r.PrepTime
	rt.CookTime = r.CookTime
	rt.Servings = r.Servings
	rt.ExtraNotes = r.ExtraNotes

	for _, ingr := range r.Ingredients {
		rt.Ingredients = append(rt.Ingredients, &types.Ingredient{
			Name:          ingr.Name,
			Quantity:      ingr.Quantity,
			UnitOfMeasure: ingr.UnitOfMeasure,
		})
	}

	for _, step := range r.Steps {
		rt.Steps = append(rt.Steps, &types.Step{
			Title:       step.Title,
			Description: step.Description,
			Image:       &types.Image{URL: step.Image.URL},
		})
	}

	var sb strings.Builder
	numTags := len(r.Tags)
	for i := 0; i < numTags-1; i++ {
		sb.WriteString(r.Tags[i].Tag)
		sb.WriteString(", ")
	}
	sb.WriteString(r.Tags[numTags-1].Tag)
	rt.Tags = sb.String()

	return &rt
}

func FromType(rt *types.Recipe) *Recipe {
	var r Recipe

	r.ExternalID = rt.ExternalID
	r.Title = rt.Title
	r.Subtitle = rt.Subtitle
	r.Likes = rt.Likes
	r.Description = rt.Description
	r.Conclusion = rt.Conclusion
	r.Difficulty = RecipeDifficulty(rt.Difficulty)
	r.Cost = RecipeCost(rt.Cost)
	r.PrepTime = rt.PrepTime
	r.CookTime = rt.CookTime
	r.Servings = rt.Servings
	r.ExtraNotes = rt.ExtraNotes

	for _, ingr := range rt.Ingredients {
		r.Ingredients = append(r.Ingredients, &Ingredient{
			Name:          ingr.Name,
			Quantity:      ingr.Quantity,
			UnitOfMeasure: ingr.UnitOfMeasure,
		})
	}

	for _, step := range rt.Steps {
		r.Steps = append(r.Steps, &Step{
			Title:       step.Title,
			Description: step.Description,
			Image:       &Image{URL: step.Image.URL},
		})
	}

	tags := strings.Split(rt.Tags, ", ")

	for _, tag := range tags {
		r.Tags = append(r.Tags, &Tag{Tag: tag})
	}

	return &r
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
