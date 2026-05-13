package database

import (
	"os"

	"github.com/redis/go-redis/v9"
)

func NewRedisDB() *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_HOST") + ":6379",
		Password: "",
		DB:       0,
	})

	return rdb
}
