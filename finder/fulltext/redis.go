package fulltext

import (
	"fmt"
	"log"

	"github.com/RedisLabs/redisearch-go/redisearch"
	"github.com/kind84/gospiga/finder/domain"
)

type redisFT struct {
	client *redisearch.Client
}

func NewRedisFT(addr string) *redisFT {
	// Create a client. By default a client is schemaless
	// unless a schema is provided when creating the index
	c := redisearch.NewClient(addr, "recipes")
	if c == nil {
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
	c.Drop()

	// Create the index with the given schema
	if err := c.CreateIndex(sc); err != nil {
		log.Fatal(err)
	}

	return &redisFT{c}
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
	if err := r.client.IndexOptions(redisearch.DefaultIndexingOptions, doc); err != nil {
		return err
	}
	return nil
}

func (r *redisFT) SearchRecipe(query string) ([]string, error) {
	// Searching with limit and sorting
	docs, tot, err := r.client.Search(redisearch.NewQuery("hello world").
		Limit(0, 2).
		SetReturnFields("title"))
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, tot)
	for _, doc := range docs {
		ids = append(ids, doc.Id)
	}
	return ids, nil
}
