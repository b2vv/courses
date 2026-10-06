package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"courses/internal/identity/ports"
	redisclient "github.com/redis/go-redis/v9"
)

type pendingTwoFactorClient interface {
	Set(context.Context, string, interface{}, time.Duration) *redisclient.StatusCmd
	GetDel(context.Context, string) *redisclient.StringCmd
}

// PendingTwoFactorStore holds no authentication material: only the user ID and
// remember-me choice between password and second-factor verification.
type PendingTwoFactorStore struct{ client pendingTwoFactorClient }

func NewPendingTwoFactorStore(client pendingTwoFactorClient) (*PendingTwoFactorStore, error) {
	if client == nil {
		return nil, fmt.Errorf("nil pending two-factor Redis client")
	}
	return &PendingTwoFactorStore{client: client}, nil
}

func (store *PendingTwoFactorStore) Put(ctx context.Context, pendingID string, data ports.PendingTwoFactor, ttl time.Duration) error {
	if pendingID == "" || data.UserID == "" || ttl <= 0 {
		return fmt.Errorf("invalid pending two-factor value")
	}
	payload, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return store.client.Set(ctx, pendingTwoFactorKey(pendingID), payload, ttl).Err()
}

func (store *PendingTwoFactorStore) Take(ctx context.Context, pendingID string) (ports.PendingTwoFactor, bool, error) {
	if pendingID == "" {
		return ports.PendingTwoFactor{}, false, nil
	}
	payload, err := store.client.GetDel(ctx, pendingTwoFactorKey(pendingID)).Bytes()
	if err == redisclient.Nil {
		return ports.PendingTwoFactor{}, false, nil
	}
	if err != nil {
		return ports.PendingTwoFactor{}, false, err
	}
	var data ports.PendingTwoFactor
	if err := json.Unmarshal(payload, &data); err != nil {
		return ports.PendingTwoFactor{}, false, fmt.Errorf("decode pending two-factor value: %w", err)
	}
	if data.UserID == "" {
		return ports.PendingTwoFactor{}, false, fmt.Errorf("decode pending two-factor value: missing user ID")
	}
	return data, true, nil
}

func pendingTwoFactorKey(id string) string { return "identity:pending-two-factor:" + id }
