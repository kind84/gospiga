package redis

import (
	"context"

	redis "github.com/redis/go-redis/v9"
)

func NewClient(host string) (*redis.Client, error) {
	opts := &redis.Options{
		Addr: host,
	}
	client := redis.NewClient(opts)

	return client, client.Ping(context.Background()).Err()
}
