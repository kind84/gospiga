package db

import (
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
