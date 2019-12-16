package redis

import (
	"github.com/go-redis/redis"
)

func NewClient(host string) (*redis.Client, error) {
	opts := &redis.Options{
		Addr: host,
	}
	client := redis.NewClient(opts)

	return client, client.Ping().Err()
}
