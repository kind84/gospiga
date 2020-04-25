package dgraph

import (
	"encoding/json"

	"github.com/kind84/gospiga/pkg/stemmer"
	"github.com/kind84/gospiga/server/domain"
)

// Ingredient represents repository verison of the domain ingredient.
type Ingredient struct {
	ID            string      `json:"uid,omitempty"`
	Name          string      `json:"name,omitempty"`
	Quantity      interface{} `json:"quantity,omitempty"`
	UnitOfMeasure string      `json:"unitOfMeasure,omitempty"`
	Food          *Food       `json:"food,omitempty"`
	Recipes       []Recipe    `json:"recipe,omitempty"`
	DType         []string    `json:"dgraph.type,omitempty"`
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

// FromDomain convert a domain ingredient into dgraph ingredient.
func (i *Ingredient) FromDomain(di *domain.Ingredient) error {
	s, err := stemmer.Stem(i.Name, "italian")
	if err != nil {
		return err
	}
	i.Name = di.Name
	i.Quantity = di.Quantity
	i.UnitOfMeasure = di.UnitOfMeasure
	i.Food = &Food{
		Term:  i.Name,
		Stem:  s,
		DType: []string{"Food"},
	}
	i.DType = []string{"Ingredient"}

	return nil
}

// ToDomain convert a dgraph ingredient into domain ingredient.
func (i *Ingredient) ToDomain() *domain.Ingredient {
	return &domain.Ingredient{
		Name:          i.Name,
		Quantity:      i.Quantity,
		UnitOfMeasure: i.UnitOfMeasure,
	}
}
