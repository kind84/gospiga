package dgraph

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"

	"github.com/kind84/gospiga/server/domain"
)

type DB struct {
	*dgo.Dgraph
}

func NewDB(ctx context.Context) (*DB, error) {
	d, err := grpc.Dial("alpha:9080", grpc.WithInsecure())
	if err != nil {
		log.Warn("failed to connect to dgraph, retrying..")
		for i := 1; i < 4; i++ {
			err = nil
			time.Sleep(time.Second)
			d, err = grpc.Dial("alpha:9080", grpc.WithInsecure())
			if err == nil {
				break
			}
		}
		if err != nil {
			log.Error("failed to connect to dgraph")
			return nil, err
		}
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
		ingredients: [uid] @count @reverse .
		steps: [uid] @count .
		conclusion: string .
		name: string @lang @index(term) .
		quantity: string .
		unitOfMeasure: string .
		image: uid .
		url: string .
	`
	// check if server is ready to go
	res, err := http.Get("http://alpha:8080/health")
	for err != nil || res.StatusCode != http.StatusOK {
		time.Sleep(time.Second)
		res, err = http.Get("http://alpha:8080/health")
	}
	log.Debug("dgraph server ready")

	err = dgraph.Alter(ctx, op)
	if err != nil {
		return nil, err
	}

	return &DB{dgraph}, nil
}

func (db *DB) SaveRecipe(ctx context.Context, recipe *domain.Recipe) error {
	mu := &api.Mutation{
		CommitNow: true,
	}

	dRecipe := Recipe{*recipe, []string{}}
	dRecipe.ID = "_:recipe"

	rb, err := json.Marshal(dRecipe)
	if err != nil {
		return err
	}

	mu.SetJson = rb
	res, err := db.Dgraph.NewTxn().Mutate(ctx, mu)
	if err != nil {
		return err
	}
	ruid := res.Uids["recipe"]
	recipe.ID = ruid
	return nil
}

func (db *DB) IDSaved(ctx context.Context, id string) (bool, error) {
	vars := map[string]string{"$id": id}
	q := `query IDSaved($id: string){
		recipes(func: eq(id, $id)) {
			uid
		}
	}`

	resp, err := db.Dgraph.NewTxn().QueryWithVars(ctx, q, vars)
	if err != nil {
		return false, err
	}

	var root struct {
		Recipes []Recipe `json:"recipes"`
	}
	err = json.Unmarshal(resp.Json, &root)
	if err != nil {
		return false, err
	}
	return len(root.Recipes) > 0, nil
}
