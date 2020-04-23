package dgraph

import (
	"encoding/json"

	"github.com/kind84/gospiga/server/domain"
)

// Ingredient represents repository verison of the domain ingredient.
type Ingredient struct {
	domain.Ingredient
	Food    *Food    `json:"food,omitempty"`
	Recipes []Recipe `json:"recipe,omitempty"`
	DType   []string `json:"dgraph.type,omitempty"`
}

// Food used as recipe ingredient.
type Food struct {
	ID          string       `json:"uid,omitempty"`
	Term        string       `json:"term,omitempty"`
	Stem        string       `json:"stem,omitempty"`
	Ingredients []Ingredient `json:"ingredient,omitempty"`
	DType       []string     `json:"dgraph.type,omitempty"`
}

func (i Ingredient) MarshalJSON() ([]byte, error) {
	type Alias Ingredient
	if len(i.DType) == 0 {
		i.DType = []string{"Ingredient"}
	}
	return json.Marshal((Alias)(i))
}
