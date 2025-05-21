package provider

import (
	"context"
	"fmt"

	"github.com/jaylane/graphql"

	"gospiga/pkg/log"
	"gospiga/pkg/types"
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

func (p *provider) GetRecipe(ctx context.Context, recipeID string) (*types.Recipe, error) {
	log.Debugf("Asking dato for recipe ID %s\n", recipeID)
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
					heading: title
					body: description
					image {
						url
					}
				}
				tags
				conclusion
				slug
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

	return &r.Recipe.Recipe, nil
}

func (p *provider) GetAllRecipeIDs(ctx context.Context) ([]string, error) {
	log.Debugf("Asking dato for all recipe IDs")

	req := graphql.NewRequest(`
		query MyQuery {
			recipesCount: _allRecipesMeta {
				count
			}
		}
	`)

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.token))

	var c struct {
		RecipesCount struct {
			Count int
		}
	}
	err := p.client.Run(ctx, req, &c)
	if err != nil {
		return nil, err
	}

	req = graphql.NewRequest(`
		query MyQuery($first: IntType!){
			recipes: allRecipes(first: $first) {
				id
			}
		}
	`)

	req.Var("first", c.RecipesCount.Count)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", p.token))

	var r struct {
		Recipes []struct {
			ID string
		}
	}
	err = p.client.Run(ctx, req, &r)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(r.Recipes))
	for _, r := range r.Recipes {
		ids = append(ids, r.ID)
	}
	return ids, nil
}
