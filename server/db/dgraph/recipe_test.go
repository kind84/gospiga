// +build integration

package dgraph

import (
	"context"
	"testing"

	"github.com/dgraph-io/dgo/v2/protos/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kind84/gospiga/server/domain"
)

var db *DB

func init() {
	var err error
	ctx := context.Background()
	db, err = NewDB(ctx)
	if err != nil {
		panic(err)
	}

	// flush db
	op := &api.Operation{DropAll: true}
	if err := db.Dgraph.Alter(ctx, op); err != nil {
		panic(err)
	}
	loadRecipeSchema()
	op = loadRecipeSchema()

	err = db.Dgraph.Alter(ctx, op)
	if err != nil {
		panic(err)
	}
}

func TestDgraphSaveRecipe(t *testing.T) {
	recipe := getTestRecipe()
	recipe2 := recipe
	recipe2.Title = "upsert"

	tests := []struct {
		name   string
		recipe *domain.Recipe
	}{
		{
			name:   "save new recipe",
			recipe: &recipe,
		},
		{
			name:   "update recipe",
			recipe: &recipe2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require := require.New(t)

			err := db.SaveRecipe(context.Background(), tt.recipe)

			require.NoError(err)
			r, err := db.GetRecipeByID(context.Background(), tt.recipe.ExternalID)
			require.NoError(err)
			require.NotNil(r)
			n, err := db.CountRecipes(context.Background())
			require.NoError(err)
			assert.Equal(t, r.ExternalID, tt.recipe.ExternalID)
			assert.Equal(t, r.Title, tt.recipe.Title)
			assert.Equal(t, n, 1)
			err = db.DeleteRecipe(context.Background(), tt.recipe.ExternalID)
			require.NoError(err)
		})
	}
}

func TestDgraphDeleteRecipe(t *testing.T) {
	recipe := getTestRecipe()

	err := db.SaveRecipe(context.Background(), &recipe)
	require.NoError(t, err)

	err = db.DeleteRecipe(context.Background(), recipe.ExternalID)

	require.NoError(t, err)
	r, err := db.GetRecipeByID(context.Background(), recipe.ExternalID)
	require.NoError(t, err)
	require.Nil(t, r)
}

func getTestRecipe() domain.Recipe {
	return domain.Recipe{
		ExternalID:  "externalID",
		Title:       "title",
		Subtitle:    "subtitle",
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
