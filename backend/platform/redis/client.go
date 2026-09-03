package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/hibiken/asynq"
	redisclient "github.com/redis/go-redis/v9"
)

func Open(ctx context.Context, redisURL string) (*redisclient.Client, error) {
	options, err := redisclient.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid Redis configuration")
	}
	options.DialTimeout = 3 * time.Second
	options.ReadTimeout = 2 * time.Second
	options.WriteTimeout = 2 * time.Second
	options.ContextTimeoutEnabled = true
	options.MaxRetries = -1
	client := redisclient.NewClient(options)
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, fmt.Errorf("Redis unavailable")
	}
	return client, nil
}

func QueueOption(redisURL string) (asynq.RedisConnOpt, error) {
	option, err := asynq.ParseRedisURI(redisURL)
	if err != nil {
		return nil, fmt.Errorf("invalid queue configuration")
	}
	// A concrete client option lets Asynq use bounded I/O for queue operations.
	if client, ok := option.(asynq.RedisClientOpt); ok {
		client.DialTimeout = 3 * time.Second
		client.ReadTimeout = 2 * time.Second
		client.WriteTimeout = 2 * time.Second
		return client, nil
	}
	return nil, fmt.Errorf("queue requires a standalone Redis connection")
}
