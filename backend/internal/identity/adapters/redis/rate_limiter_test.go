package redis

import (
	"context"
	"errors"
	"testing"
	"time"

	redisclient "github.com/redis/go-redis/v9"
)

// fakeRateLimiterClient drives the three Redis calls the limiter makes, letting a test force an
// error on Incr (a Redis outage) or Del (a failed reset) without a real Redis.
type fakeRateLimiterClient struct {
	incrErr error
	delErr  error
	count   int64
}

func (fake *fakeRateLimiterClient) Incr(ctx context.Context, _ string) *redisclient.IntCmd {
	cmd := redisclient.NewIntCmd(ctx)
	if fake.incrErr != nil {
		cmd.SetErr(fake.incrErr)
		return cmd
	}
	fake.count++
	cmd.SetVal(fake.count)
	return cmd
}

func (fake *fakeRateLimiterClient) Expire(ctx context.Context, _ string, _ time.Duration) *redisclient.BoolCmd {
	cmd := redisclient.NewBoolCmd(ctx)
	cmd.SetVal(true)
	return cmd
}

func (fake *fakeRateLimiterClient) Del(ctx context.Context, _ ...string) *redisclient.IntCmd {
	cmd := redisclient.NewIntCmd(ctx)
	if fake.delErr != nil {
		cmd.SetErr(fake.delErr)
		return cmd
	}
	cmd.SetVal(1)
	return cmd
}

// SA08/M1: a Redis outage must not remove the auth limit. Incr failing means Redis is unreachable,
// so Allow must fail CLOSED (deny) — the caller maps a false to ErrRateLimited. The pre-fix code
// returned (true, nil) here, letting an outage open login/2FA/reset to brute force.
func TestAllowFailsClosedWhenRedisIncrementErrors(t *testing.T) {
	limiter, err := NewRateLimiter(&fakeRateLimiterClient{incrErr: errors.New("redis down")}, 5, 15*time.Minute, nil)
	if err != nil {
		t.Fatalf("construct limiter: %v", err)
	}
	allowed, err := limiter.Allow(context.Background(), "login:student@example.test:203.0.113.5")
	if err != nil {
		t.Fatalf("Allow should not surface the Redis error, got %v", err)
	}
	if allowed {
		t.Fatal("Allow must fail closed (deny) when Redis is unreachable")
	}
}

// A healthy Redis still allows under the limit and denies over it — the fail-closed change must not
// break the normal decision.
func TestAllowEnforcesLimitWhenRedisHealthy(t *testing.T) {
	limiter, err := NewRateLimiter(&fakeRateLimiterClient{}, 2, 15*time.Minute, nil)
	if err != nil {
		t.Fatalf("construct limiter: %v", err)
	}
	ctx := context.Background()
	key := "login:student@example.test:203.0.113.5"
	for request := 1; request <= 2; request++ {
		allowed, err := limiter.Allow(ctx, key)
		if err != nil || !allowed {
			t.Fatalf("request %d under the limit: allowed=%v err=%v", request, allowed, err)
		}
	}
	allowed, err := limiter.Allow(ctx, key)
	if err != nil {
		t.Fatalf("over the limit should not error, got %v", err)
	}
	if allowed {
		t.Fatal("Allow must deny once the count exceeds the limit")
	}
}

// Reset clears the counter on a successful auth. If Redis cannot delete the key, leaving the counter
// is the safe direction (the account stays limited); Reset must not fail the caller's already
// successful authentication, so it returns nil.
func TestResetLeavesCounterAndSucceedsWhenRedisErrors(t *testing.T) {
	limiter, err := NewRateLimiter(&fakeRateLimiterClient{delErr: errors.New("redis down")}, 5, 15*time.Minute, nil)
	if err != nil {
		t.Fatalf("construct limiter: %v", err)
	}
	if err := limiter.Reset(context.Background(), "login:student@example.test:203.0.113.5"); err != nil {
		t.Fatalf("Reset must not fail a successful auth over a Redis error, got %v", err)
	}
}
