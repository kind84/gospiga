package dgraph

import (
	"context"
	"encoding/json"

	"github.com/kind84/gospiga/server/domain"
)

// tag represents repository version of the domain tag.
type Tag struct {
	ID      string   `json:"uid,omitempty"`
	TagName string   `json:"tagName,omitempty"`
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

// AllTagsImages returns one recipe image for each tag stored on db.
func (db *DB) AllTagsImages(ctx context.Context) ([]*domain.Tag, error) {
	q := `
		query Tags {
			tags(func: has(tagName)) {
				tagName
				recipes: ~tags (first: 1){
					uid
					xid
					mainImage {
						url
					}
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
