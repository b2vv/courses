package redis

import (
	"context"
	"encoding/json"
	"time"

	"courses/internal/playback/domain"
	redisclient "github.com/redis/go-redis/v9"
)

const renewOwnershipScript = `local v = redis.call('GET', KEYS[1])
if not v then return -1 end
local ok, stored = pcall(cjson.decode, v)
if not ok or type(stored) ~= 'table' or type(stored.playbackId) ~= 'string' or stored.playbackId == '' then
  return redis.error_reply('invalid ownership value')
end
if stored.playbackId ~= ARGV[1] then return 0 end
redis.call('EXPIRE', KEYS[1], ARGV[2])
return 1`

type ownershipWriterClient interface {
	Set(context.Context, string, any, time.Duration) *redisclient.StatusCmd
	Eval(context.Context, string, []string, ...any) *redisclient.Cmd
}

type OwnershipWriter struct{ client ownershipWriterClient }

func NewOwnershipWriter(client ownershipWriterClient) *OwnershipWriter {
	return &OwnershipWriter{client: client}
}

func (writer *OwnershipWriter) Take(ctx context.Context, userID string, ownership domain.Ownership, ttl time.Duration) error {
	encoded, err := json.Marshal(storedOwnership{
		PlaybackID: ownership.PlaybackID,
		VideoID:    ownership.VideoID,
		LessonID:   ownership.LessonID,
	})
	if err != nil {
		return domain.ErrDependencyUnavailable
	}
	if err := writer.client.Set(ctx, ownershipKey(userID), encoded, ttl).Err(); err != nil {
		return domain.ErrDependencyUnavailable
	}
	return nil
}

func (writer *OwnershipWriter) Renew(ctx context.Context, userID, playbackID string, ttl time.Duration) (domain.OwnershipRenewal, error) {
	result, err := writer.client.Eval(
		ctx,
		renewOwnershipScript,
		[]string{ownershipKey(userID)},
		playbackID,
		int64(ttl/time.Second),
	).Int64()
	if err != nil {
		return 0, domain.ErrDependencyUnavailable
	}
	switch result {
	case -1:
		return domain.OwnershipMissing, nil
	case 0:
		return domain.OwnershipTakenOver, nil
	case 1:
		return domain.OwnershipRenewed, nil
	default:
		return 0, domain.ErrDependencyUnavailable
	}
}
