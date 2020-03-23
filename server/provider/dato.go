package provider

import (
	"context"
	"fmt"

	"github.com/jaylane/graphql"

	"github.com/kind84/gospiga/pkg/types"
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
				tags
				conclusion
			}
		}
	`)
	req.Var("key", recipeID)

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.token))

	var r struct {
		Recipe struct {
			types.Recipe
		}
	}
	err := p.client.Run(ctx, req, &r)
	if err != nil {
		return nil, err
	}

	// map to the domain recipe.
	return domain.FromType(&r.Recipe.Recipe), nil
}
