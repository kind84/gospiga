package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/RedisLabs/redisearch-go/redisearch"
)

func main() {
	file, err := os.Open("../../../include/stopwords-it")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var sw []string
	for scanner.Scan() {
		sw = append(sw, scanner.Text())
	}

	opts := redisearch.DefaultOptions
	opts.Stopwords = sw
	fmt.Printf("%+v", opts)

	// Create a schema
	sc := redisearch.NewSchema(opts).
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

	fmt.Printf("%+v", sc)
}
