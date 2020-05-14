package dgraph

import (
	"context"
	"encoding/json"

	"gospiga/pkg/stemmer"
	"gospiga/server/domain"
)

// tag represents repository version of the domain tag.
type Tag struct {
	ID      string   `json:"uid,omitempty"`
	TagName string   `json:"tagName,omitempty"`
	TagStem string   `json:"tagStem,omitempty"`
	Recipes []Recipe `json:"recipes,omitempty"`
	DType   []string `json:"dgraph.type,omitempty"`
}

func (t Tag) MarshalJSON() ([]byte, error) {
	type Alias Tag
	if len(t.DType) == 0 {
		t.DType = []string{"Tag"}
	}
	return json.Marshal((Alias)(t))
}

// ToDomain converts a dgraph tag into a domain tag.
func (t *Tag) ToDomain() *domain.Tag {
	dt := &domain.Tag{TagName: t.TagName}
	dt.Recipes = make([]*domain.Recipe, 0, len(t.Recipes))
	for _, r := range t.Recipes {
		dt.Recipes = append(dt.Recipes, r.ToDomain()) // /!\ recursive
	}
	return dt
}

// FromDomain converts a domain tag into a dgraph tag.
func (t *Tag) FromDomain(dt *domain.Tag) error {
	t.TagName = dt.TagName
	s, err := stemmer.Stem(t.TagName, "italian")
	if err != nil {
		return err
	}
	t.TagStem = s
	t.DType = []string{"Tag"}

	return nil
}

// AllTagsImages returns one recipe image for each tag stored on db.
func (db *DB) AllTagsImages(ctx context.Context) ([]*domain.Tag, error) {
	q := `
		query Tags {
			tags(func: has(tagName)) {
				tagName
				recipes: ~tags (first: 1){
					uid
					xid
					mainImage
				}
			}
		}
	`

	resp, err := db.Dgraph.NewReadOnlyTxn().Query(ctx, q)
	if err != nil {
		return nil, err
	}

	var root struct {
		Tags []Tag `json:"tags"`
	}
	err = json.Unmarshal(resp.Json, &root)
	if err != nil {
		return nil, err
	}
	if len(root.Tags) == 0 {
		return nil, nil
	}

	tags := make([]*domain.Tag, 0, len(root.Tags))
	for _, t := range root.Tags {
		tags = append(tags, &domain.Tag{
			TagName: t.TagName,
			Recipes: []*domain.Recipe{t.Recipes[0].ToDomain()},
		})
	}
	return tags, nil
}
