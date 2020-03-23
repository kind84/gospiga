package db

import (
	"fmt"

	redis "github.com/go-redis/redis/v7"
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

func (r *redisDB) Tags(index, field string) ([]string, error) {
	cmd := redis.NewStringSliceCmd("tagvals", index, field)
	err := r.rdb.Process(cmd)
	if err != nil {
		return nil, fmt.Errorf("error collecting tags: %w", err)
	}
	tags, err := cmd.Result()
	if err != nil {
		return nil, fmt.Errorf("error collecting tags: %w", err)
	}
	return tags, nil
}
