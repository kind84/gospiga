package usecase

import (
	"context"

	"github.com/kind84/gospiga/pkg/types"
)

func (a *app) AllTagsImages(ctx context.Context) ([]*types.Tag, error) {
	dtags, err := a.db.AllTagsImages(ctx)
	if err != nil {
		return nil, err
	}

	tags := make([]*types.Tag, 0, len(dtags))
	for _, t := range dtags {
		tags = append(tags, t.ToType())
	}

	return tags, nil
}
