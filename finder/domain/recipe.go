package domain

import (
	"fmt"
	"github.com/kind84/gospiga/pkg/types"
	"strings"
)

type Recipe struct {
	ID          string                 `json:"id,omitempty"`
	Title       string                 `json:"title,omitempty"`
	Subtitle    string                 `json:"subtitle,omitempty"`
	Likes       int                    `json:"likes,omitempty"`
	Difficulty  types.RecipeDifficulty `json:"difficulty,omitempty"`
	Cost        types.RecipeCost       `json:"cost,omitempty"`
	PrepTime    int                    `json:"prepTime,omitempty"`
	CookTime    int                    `json:"cookTime,omitempty"`
	Servings    int                    `json:"servings,omitempty"`
	ExtraNotes  string                 `json:"extraNotes,omitempty"`
	Description string                 `json:"description,omitempty"`
	Ingredients []string               `json:"ingredients,omitempty"`
	Steps       []string               `json:"steps,omitempty"`
	Conclusion  string                 `json:"conclusion,omitempty"`
}

func (r *Recipe) MapFromType(rt *types.Recipe) {
	r.ID = rt.ID
	r.Title = rt.Title
	r.Subtitle = rt.Subtitle
	r.Likes = rt.Likes
	r.Description = rt.Description
	r.Conclusion = rt.Conclusion
	r.Difficulty = rt.Difficulty
	r.Cost = rt.Cost
	r.PrepTime = rt.PrepTime
	r.CookTime = rt.CookTime
	r.Servings = rt.Servings
	r.ExtraNotes = rt.ExtraNotes

	for _, ingr := range rt.Ingredients {
		var sb strings.Builder
		sb.WriteString(ingr.Quantity.(string))
		sb.WriteString(" ")
		sb.WriteString(ingr.UnitOfMeasure)
		sb.WriteString(" ")
		sb.WriteString(ingr.Name)

		r.Ingredients = append(r.Ingredients, sb.String())
	}

	for _, step := range rt.Steps {
		r.Steps = append(r.Steps, fmt.Sprintf("%s %s", step.Title, step.Description))
	}
}
