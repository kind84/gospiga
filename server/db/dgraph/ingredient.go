package dgraph

import (
	"encoding/json"

	"github.com/kind84/gospiga/server/domain"
)

// ingredient represents repository verison of the domain ingredient.
type Ingredient struct {
	domain.Ingredient
	Recipes []Recipe `json:"recipe,omitempty"`
	DType   []string `json:"dgraph.type,omitempty"`
}

func (i Ingredient) MarshalJSON() ([]byte, error) {
	type Alias Ingredient
	if len(i.DType) == 0 {
		i.DType = []string{"Ingredient"}
	}
	return json.Marshal((Alias)(i))
}
