package fulltext

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/RedisLabs/redisearch-go/redisearch"
	log "github.com/sirupsen/logrus"

	"github.com/kind84/gospiga/finder/domain"
)

type redisFT struct {
	ft *redisearch.Client
}

// NewRedisFT returns a new instance of the Full Text Redis client.
func NewRedisFT(addr string) (*redisFT, error) {
	// Create a client. By default a client is schemaless
	// unless a schema is provided when creating the index
	ft := redisearch.NewClient(addr, "recipes")
	if ft == nil {
		return nil, errors.New("cannot initialize redis client")
	}

	// Create a schema
	sc := redisearch.NewSchema(redisearch.DefaultOptions).
		AddField(redisearch.NewTextFieldOptions("id", redisearch.TextFieldOptions{NoIndex: true})).
		AddField(redisearch.NewTextFieldOptions("xid", redisearch.TextFieldOptions{NoIndex: true})).
		AddField(redisearch.NewTextFieldOptions("title", redisearch.TextFieldOptions{Weight: 5.0, Sortable: true})).
		AddField(redisearch.NewTextField("subtitle")).
		AddField(redisearch.NewTextFieldOptions("mainImage", redisearch.TextFieldOptions{NoIndex: true})).
		AddField(redisearch.NewTextField("description")).
		AddField(redisearch.NewNumericFieldOptions("prepTime", redisearch.NumericFieldOptions{NoIndex: true})).
		AddField(redisearch.NewNumericFieldOptions("cookTime", redisearch.NumericFieldOptions{NoIndex: true})).
		AddField(redisearch.NewNumericFieldOptions("time", redisearch.NumericFieldOptions{Sortable: true})).
		AddField(redisearch.NewTextFieldOptions("ingredients", redisearch.TextFieldOptions{Weight: 4.0})).
		AddField(redisearch.NewTextField("steps")).
		AddField(redisearch.NewTextField("conclusion")).
		AddField(redisearch.NewTagField("tags"))

	// Drop an existing index. If the index does not exist an error is returned
	ft.Drop()

	// Create the index with the given schema
	if err := ft.CreateIndex(sc); err != nil {
		return nil, err
	}

	return &redisFT{ft}, nil
}

// IndexRecipe adds a new recipe to the index.
func (r *redisFT) IndexRecipe(recipe *domain.Recipe) error {
	// Create a document with an id and given score
	doc := redisearch.NewDocument(fmt.Sprintf("recipe:%s", recipe.ID), 1.0)

	doc.Set("id", recipe.ID).
		Set("xid", recipe.ExternalID).
		Set("title", recipe.Title).
		Set("subtitle", recipe.Subtitle).
		Set("mainImage", recipe.MainImageURL).
		Set("description", recipe.Description).
		Set("prepTime", recipe.PrepTime).
		Set("cookTime", recipe.CookTime).
		Set("time", recipe.PrepTime+recipe.CookTime).
		Set("ingredients", recipe.Ingredients).
		Set("steps", recipe.Steps).
		Set("conclusion", recipe.Conclusion).
		Set("tags", recipe.Tags)

	// Index the document. The API accepts multiple documents at a time,
	opts := redisearch.DefaultIndexingOptions
	opts.Replace = true // upsert
	if err := r.ft.IndexOptions(opts, doc); err != nil {
		return err
	}
	return nil
}

func (r *redisFT) DeleteRecipe(recipeID string) error {
	return r.ft.Delete(recipeID, true)
}

func (r *redisFT) SearchRecipes(query string) ([]*Recipe, error) {
	docs, tot, err := r.ft.Search(redisearch.NewQuery(query))
	if err != nil {
		log.Error(err)
		return nil, err
	}

	recipes := make([]*Recipe, 0, tot)
	for _, doc := range docs {
		var recipe Recipe
		jr, err := json.Marshal(doc.Properties)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		err = json.Unmarshal(jr, &recipe)
		if err != nil {
			log.Error(err)
			return nil, err
		}
		recipes = append(recipes, &recipe)
	}
	return recipes, nil
}
