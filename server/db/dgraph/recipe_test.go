// +build integration

package dgraph

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kind84/gospiga/server/domain"
)

var db *DB

func init() {
	var err error
	db, err = NewDB(context.Background())
	if err != nil {
		panic(err)
	}
}

func TestDgraphDeleteRecipe(t *testing.T) {
	recipe := getTestRecipe()

	err := db.SaveRecipe(context.Background(), recipe)
	require.NoError(t, err)

	err = db.DeleteRecipe(context.Background(), recipe.ExternalID)

	require.NoError(t, err)
	r, err := db.GetRecipeByID(context.Background(), recipe.ExternalID)
	require.NoError(t, err)
	require.Nil(t, r)
}

func getTestRecipe() *domain.Recipe {
	return &domain.Recipe{
		ExternalID:  "externalID",
		Title:       "title",
		Subtitle:    "Subtitle",
		Description: "description",
		Conclusion:  "conclusion",
		MainImage: domain.Image{
			Url: "url",
		},
		Difficulty: domain.DifficultyEasy,
		Cost:       domain.CostLow,
		Servings:   1,
		PrepTime:   1,
		CookTime:   1,
		Ingredients: []*domain.Ingredient{
			{
				Name:          "ingredient",
				Quantity:      1,
				UnitOfMeasure: "unitOfMeasure",
			},
		},
		Steps: []*domain.Step{
			{
				Title:       "title",
				Description: "description",
				Image: domain.Image{
					Url: "url",
				},
			},
		},
	}
}
