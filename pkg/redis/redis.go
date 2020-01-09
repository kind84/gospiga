package redis

import (
	redis "github.com/go-redis/redis/v7"
)

func NewClient(host string) (*redis.Client, error) {
	opts := &redis.Options{
		Addr: host,
	}
	client := redis.NewClient(opts)

	return client, client.Ping().Err()
}
