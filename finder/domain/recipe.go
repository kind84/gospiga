package domain

import (
	"fmt"
	"strconv"

	"github.com/kind84/gospiga/pkg/types"
)

type Recipe struct {
	ID           string           `json:"id,omitempty"`
	ExternalID   string           `json:"xid,omitempty"`
	Title        string           `json:"title,omitempty"`
	Subtitle     string           `json:"subtitle,omitempty"`
	MainImageURL string           `json:"mainImageURL,omitempty"`
	Likes        int              `json:"likes,omitempty"`
	Difficulty   RecipeDifficulty `json:"difficulty,omitempty"`
	Cost         RecipeCost       `json:"cost,omitempty"`
	PrepTime     int              `json:"prepTime,omitempty"`
	CookTime     int              `json:"cookTime,omitempty"`
	Servings     int              `json:"servings,omitempty"`
	ExtraNotes   string           `json:"extraNotes,omitempty"`
	Description  string           `json:"description,omitempty"`
	Ingredients  []string         `json:"ingredients,omitempty"`
	Steps        []string         `json:"steps,omitempty"`
	Tags         string           `json:"tags,omitempty"`
	Conclusion   string           `json:"conclusion,omitempty"`
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

func FromType(rt *types.Recipe) *Recipe {
	var r Recipe

	r.ID = rt.ID
	r.ExternalID = rt.ExternalID
	r.Title = rt.Title
	r.Subtitle = rt.Subtitle
	r.MainImageURL = rt.MainImage.URL
	r.Likes = rt.Likes
	r.Description = rt.Description
	r.Conclusion = rt.Conclusion
	r.Difficulty = RecipeDifficulty(rt.Difficulty)
	r.Cost = RecipeCost(rt.Cost)
	r.PrepTime = rt.PrepTime
	r.CookTime = rt.CookTime
	r.Servings = rt.Servings
	r.ExtraNotes = rt.ExtraNotes
	r.Tags = rt.Tags

	for _, ingr := range rt.Ingredients {
		var qty string
		if q, ok := ingr.Quantity.(string); ok {
			qty = q
		} else if ingr.Quantity != nil {
			qty = strconv.Itoa(int(ingr.Quantity.(float64)))
		}
		r.Ingredients = append(r.Ingredients, fmt.Sprintf("%s %s %s", qty, ingr.UnitOfMeasure, ingr.Name))
	}

	for _, step := range rt.Steps {
		r.Steps = append(r.Steps, fmt.Sprintf("%s %s", step.Title, step.Description))
	}

	return &r
}
