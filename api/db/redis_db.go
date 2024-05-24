package db

import (
	"github.com/redis/go-redis/v9"
)

func MustConnectToRedis(redisUrl string) *redis.Client {
	opts, err := redis.ParseURL(redisUrl)
	if err != nil {
		panic(err)
	}

	return redis.NewClient(opts)
}
