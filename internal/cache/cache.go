package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(ctx context.Context, redisURL string) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr: redisURL,
		DB:   0,
	})

	// Test the connection
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("unable to ping redis: %w", err)
	}

	return client, nil
}
