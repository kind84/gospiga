package db

import (
	"github.com/go-redis/redis"
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
