package fulltext

import (
	"fmt"
	"strings"

	"github.com/RedisLabs/redisearch-go/redisearch"
	log "github.com/sirupsen/logrus"

	"github.com/kind84/gospiga/finder/domain"
)

type redisFT struct {
	ft *redisearch.Client
}

func NewRedisFT(addr string) *redisFT {
	// Create a client. By default a client is schemaless
	// unless a schema is provided when creating the index
	ft := redisearch.NewClient(addr, "recipes")
	if ft == nil {
		return nil
	}

	// Create a schema
	sc := redisearch.NewSchema(redisearch.DefaultOptions).
		AddField(redisearch.NewTextFieldOptions("title", redisearch.TextFieldOptions{Weight: 5.0, Sortable: true})).
		AddField(redisearch.NewTextField("subtitle")).
		AddField(redisearch.NewTextField("description")).
		AddField(redisearch.NewTextFieldOptions("ingredients", redisearch.TextFieldOptions{Weight: 4.0})).
		AddField(redisearch.NewTextField("steps")).
		AddField(redisearch.NewTextField("conclusion"))

	// Drop an existing index. If the index does not exist an error is returned
	ft.Drop()

	// Create the index with the given schema
	if err := ft.CreateIndex(sc); err != nil {
		log.Fatal(err)
	}

	return &redisFT{ft}
}

func (r *redisFT) IndexRecipe(recipe *domain.Recipe) error {
	// Create a document with an id and given score
	doc := redisearch.NewDocument(fmt.Sprintf("recipe-%s", recipe.ID), 1.0)

	doc.Set("title", recipe.Title).
		Set("subtitle", recipe.Subtitle).
		Set("description", recipe.Description).
		Set("ingredients", recipe.Ingredients).
		Set("steps", recipe.Steps).
		Set("conclusion", recipe.Conclusion)

	// Index the document. The API accepts multiple documents at a time,
	if err := r.ft.IndexOptions(redisearch.DefaultIndexingOptions, doc); err != nil {
		return err
	}
	return nil
}

func (r *redisFT) SearchRecipes(query string) ([]string, error) {
	docs, tot, err := r.ft.Search(redisearch.NewQuery(query))
	if err != nil {
		log.Error(err)
		return nil, err
	}

	ids := make([]string, 0, tot)
	for _, doc := range docs {
		id := strings.Split(doc.Id, "recipe-")[1]
		ids = append(ids, id)
	}
	return ids, nil
}
