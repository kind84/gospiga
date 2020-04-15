package dgraph

import (
	"context"
	"encoding/json"

	"github.com/kind84/gospiga/server/domain"
)

type TagImage struct {
	TagName string   `json:"tagName,omitempty"`
	Recipes []recipe `json:"recipe,omitempty"`
}

func (db *DB) AllTagsImages(ctx context.Context) ([]*domain.Tag, error) {
	q := `
		query Tags {
			tags(func: has(tagName)) {
				tagName
				recipe: ~tags (orderdesc: likes, first: 1){
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
		Tags []TagImage `json:"tags"`
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
		recipes := make([]*domain.Recipe, 0, len(t.Recipes))
		for _, r := range t.Recipes {
			recipes = append(recipes, &r.Recipe)
		}

		tags = append(tags, &domain.Tag{
			TagName: t.TagName,
			Recipes: recipes,
		})
	}
	return tags, nil
}
