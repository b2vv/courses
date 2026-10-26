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
		// Fail CLOSED (SA08/M1): this limiter guards login, 2FA and password-reset. A Redis outage
		// must not silently remove those limits (that amplified H1); deny instead. The caller maps
		// `!allowed` to ErrRateLimited, so auth is refused for as long as Redis is unreachable —
		// a deliberate availability-for-security trade documented in the audit §SA08.
		limiter.unavailable("increment", err)
		return false, nil
	}
	if count == 1 {
		if err := limiter.client.Expire(ctx, rateLimitKey(key), limiter.window).Err(); err != nil {
			// Incr already succeeded, so Redis is up and the count is real and enforced — only the
			// TTL housekeeping failed. Log it, but keep the real decision rather than spuriously
			// deny a legitimate first attempt.
			limiter.unavailable("set expiry", err)
		}
	}
	return count <= limiter.limit, nil
}

func (limiter *RateLimiter) Reset(ctx context.Context, key string) error {
	if err := limiter.client.Del(ctx, rateLimitKey(key)).Err(); err != nil {
		// Leaving the counter in place is the safe (closed) direction — the account stays limited.
		// Do not fail the caller's already-successful authentication over a Redis hiccup.
		limiter.unavailable("reset", err)
	}
	return nil
}

func (limiter *RateLimiter) unavailable(operation string, err error) {
	limiter.logger.Error("identity rate limiter unavailable", "operation", operation, "error", err)
}

func rateLimitKey(key string) string { return "identity:rate-limit:" + key }
