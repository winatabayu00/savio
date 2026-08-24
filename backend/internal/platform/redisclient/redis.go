package redisclient

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Connect opens a Redis client. The client is lazy; failures surface on first use.
func Connect(url string) (*redis.Client, error) {
	opts, err := redis.ParseURL(url)
	if err != nil {
		return nil, err
	}
	c := redis.NewClient(opts)
	return c, nil
}

// Ping verifies Redis liveness with a short timeout.
func Ping(ctx context.Context, c *redis.Client) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return c.Ping(ctx).Err()
}
