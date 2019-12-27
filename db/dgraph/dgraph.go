package dgraph

import (
	"context"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"google.golang.org/grpc"

	"github.com/kind84/gospiga/domain"
)

type DB struct {
	*dgo.Dgraph
}

func NewDB(ctx context.Context) (*DB, error) {
	d, err := grpc.Dial("localhost:9080", grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	dgraph := dgo.NewDgraphClient(
		api.NewDgraphClient(d),
	)

	// load schema
	op := &api.Operation{}
	op.Schema = `
		type Recipe {
			id        
			title     
			subtitle  
			mainImage 
			likes     
			difficulty
			cost     
			prepTime  
			cookTime  
			servings  
			extraNotes
			description
			ingredients
			steps      
			conclusion 
		}

		type Ingredient {
			name         
			quantity     
			unitOfMeasure
		}
		
		type Step {
			title      
			description
			image      
		}

		type Image {
			url
		}

		id: string @index(exact) .
		title: string @lang @index(fulltext) .
		subtitle: string @lang @index(fulltext) .
		mainImage: uid .
		likes: int @index(int) .
		difficulty: string .
		cost: string .
		prepTime: int @index(int) .
		cookTime: int @index(int) .
		servings: int .
		extraNotes: string .
		description: string @lang @index(fulltext) .
		ingredients: @count @reverse [uid] .
		steps: @count [uid] .
		conclusion: string .
		name: string @lang @index(term) .
		quantity: string . 
		unitOfMeasure: string .
		image: uid .
		url: string .
	`

	err = dgraph.Alter(ctx, op)
	if err != nil {
		return nil, err
	}

	return &DB{dgraph}, nil
}

func (db *DB) SaveRecipe(ctx context.Context, recipe *domain.Recipe) error {
	return nil
}
