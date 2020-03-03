package provider

import (
	"context"
	"encoding/json"
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

	// ping-pong to get the domain recipe.
	var mrecipe map[string]interface{}
	bs, err := json.Marshal(r.Recipe)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bs, &mrecipe)
	if err != nil {
		return nil, err
	}
	var recipe domain.Recipe
	bs, err = json.Marshal(mrecipe)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(bs, &recipe)
	if err != nil {
		return nil, err
	}
	return &recipe, nil
}
