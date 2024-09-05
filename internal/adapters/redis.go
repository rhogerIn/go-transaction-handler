package adapters

import (
    "github.com/go-redis/redis/v8"
)

func NewRedisClient(addr string) *redis.Client {
    return redis.NewClient(&redis.Options{
        Addr:     addr,
        Password: "", // No password
        DB:       0,  // Use default DB
    })
}
