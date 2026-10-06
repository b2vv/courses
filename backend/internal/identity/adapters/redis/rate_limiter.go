package redis

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	redisclient "github.com/redis/go-redis/v9"
)

type rateLimiterClient interface {
	Incr(context.Context, string) *redisclient.IntCmd
	Expire(context.Context, string, time.Duration) *redisclient.BoolCmd
	Del(context.Context, ...string) *redisclient.IntCmd
}

type RateLimiter struct {
	client rateLimiterClient
	limit  int64
	window time.Duration
	logger *slog.Logger
}

func NewRateLimiter(client rateLimiterClient, limit int64, window time.Duration, logger *slog.Logger) (*RateLimiter, error) {
	if client == nil || limit < 1 || window <= 0 {
		return nil, fmt.Errorf("invalid rate limiter configuration")
	}
	if logger == nil {
		logger = slog.Default()
	}
	return &RateLimiter{client: client, limit: limit, window: window, logger: logger}, nil
}

func (limiter *RateLimiter) Allow(ctx context.Context, key string) (bool, error) {
	count, err := limiter.client.Incr(ctx, rateLimitKey(key)).Result()
	if err != nil {
		limiter.failOpen("increment", err)
		return true, nil
	}
	if count == 1 {
		if err := limiter.client.Expire(ctx, rateLimitKey(key), limiter.window).Err(); err != nil {
			limiter.failOpen("set expiry", err)
			return true, nil
		}
	}
	return count <= limiter.limit, nil
}

func (limiter *RateLimiter) Reset(ctx context.Context, key string) error {
	if err := limiter.client.Del(ctx, rateLimitKey(key)).Err(); err != nil {
		limiter.failOpen("reset", err)
	}
	return nil
}

func (limiter *RateLimiter) failOpen(operation string, err error) {
	limiter.logger.Error("identity rate limiter unavailable; allowing request", "operation", operation, "error", err)
}

func rateLimitKey(key string) string { return "identity:rate-limit:" + key }
