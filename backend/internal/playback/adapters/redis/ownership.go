package redis

import (
	"context"
	"encoding/json"
	"errors"

	"courses/internal/playback/domain"
	redisclient "github.com/redis/go-redis/v9"
)

type stringGetter interface {
	Get(context.Context, string) *redisclient.StringCmd
}

type OwnershipReader struct{ client stringGetter }

func NewOwnershipReader(client stringGetter) *OwnershipReader {
	return &OwnershipReader{client: client}
}

type storedOwnership struct {
	PlaybackID string `json:"playbackId"`
	VideoID    string `json:"videoId"`
	LessonID   string `json:"lessonId"`
}

func (reader *OwnershipReader) Load(ctx context.Context, userID string) (domain.Ownership, error) {
	value, err := reader.client.Get(ctx, "playback:owner:"+userID).Result()
	if errors.Is(err, redisclient.Nil) {
		return domain.Ownership{}, domain.ErrForbidden
	}
	if err != nil {
		return domain.Ownership{}, domain.ErrDependencyUnavailable
	}
	var stored storedOwnership
	if err := json.Unmarshal([]byte(value), &stored); err != nil {
		return domain.Ownership{}, domain.ErrDependencyUnavailable
	}
	if stored.PlaybackID == "" || stored.VideoID == "" || stored.LessonID == "" {
		return domain.Ownership{}, domain.ErrDependencyUnavailable
	}
	return domain.Ownership{PlaybackID: stored.PlaybackID, VideoID: stored.VideoID, LessonID: stored.LessonID}, nil
}
