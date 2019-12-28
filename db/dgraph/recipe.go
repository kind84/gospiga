package dgraph

import (
	"encoding/json"

	"github.com/kind84/gospiga/domain"
)

type Recipe struct {
	domain.Recipe
	DType []string `json:"dgraph.type,omitempty"`
}

func (r Recipe) MarshalJSON() ([]byte, error) {
	type Alias Recipe
	if len(r.DType) == 0 {
		r.DType = []string{"Recipe"}
	}
	return json.Marshal((Alias)(r))
}
