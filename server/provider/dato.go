package provider

import (
	"context"
	"fmt"

	"github.com/jaylane/graphql"

	"github.com/kind84/gospiga/server/domain"
)

type provider struct {
	client *graphql.Client
	token  string
}

func NewDatoProvider(token string) (*provider, error) {
	client := graphql.NewClient("https://graphql.datocms.com/preview")
	return &provider{
		client: client,
		token:  token,
	}, nil
}

func (p *provider) GetRecipe(ctx context.Context, recipeID string) (*domain.Recipe, error) {
	fmt.Printf("Asking dato for recipe ID %s\n", recipeID)
	req := graphql.NewRequest(`
		query MyQuery($key: ItemId!) {
			recipe (filter: {id: {eq: $key}}) {
				id
				title
				subtitle
				mainImage {
					url
				}
				likes
				difficulty
				cost
				prepTime
				cookTime
				servings
				extraNotes
				description
				ingredients {
					name
					quantity
					unitOfMeasure
				}
				steps {
					title
					description
					image {
						url
					}
				}
				conclusion
			}
		}
	`)
	req.Var("key", recipeID)

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.token))

	var r struct {
		Recipe struct {
			domain.Recipe
			DatoID string `json:"id"`
		}
	}
	err := p.client.Run(ctx, req, &r)
	if err != nil {
		return nil, err
	}
	r.Recipe.ExternalID = r.Recipe.DatoID

	// map to the domain recipe.
	recipe := domain.Recipe{
		ID:         r.Recipe.ID,
		ExternalID: r.Recipe.ExternalID,
		Title:      r.Recipe.Title,
		Subtitle:   r.Recipe.Subtitle,
		MainImage: &domain.Image{
			URL: r.Recipe.MainImage.URL,
		},
		Likes:       r.Recipe.Likes,
		Difficulty:  r.Recipe.Difficulty,
		Cost:        r.Recipe.Cost,
		PrepTime:    r.Recipe.PrepTime,
		CookTime:    r.Recipe.CookTime,
		Servings:    r.Recipe.Servings,
		ExtraNotes:  r.Recipe.ExtraNotes,
		Description: r.Recipe.Description,
		Ingredients: r.Recipe.Ingredients,
		Steps:       r.Recipe.Steps,
		Conclusion:  r.Recipe.Conclusion,
	}
	return &recipe, nil
}
