package db

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type redisDB struct {
	rdb *redis.Client
}

func NewRedisDB(client *redis.Client) *redisDB {
	return &redisDB{client}
}

func (r *redisDB) IDExists(id string) (bool, error) {
	return false, nil
}

func (r *redisDB) Tags(ctx context.Context, index, field string) ([]string, error) {
	cmd := redis.NewStringSliceCmd(ctx, "ft.tagvals", index, field)
	err := r.rdb.Process(ctx, cmd)
	if err != nil {
		return nil, fmt.Errorf("error collecting tags: %w", err)
	}
	tags, err := cmd.Result()
	if err != nil {
		return nil, fmt.Errorf("error collecting tags: %w", err)
	}
	return tags, nil
}
