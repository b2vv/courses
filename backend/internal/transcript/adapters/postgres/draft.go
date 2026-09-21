package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"courses/internal/transcript/domain"
	"courses/internal/transcript/ports"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

func (repository *Repository) GetLessonDraft(ctx context.Context, lessonID string) (ports.LessonDraft, error) {
	var result ports.LessonDraft

	var assetID, state string
	var generation int32
	var language pgtype.Text
	var segmentsJSON []byte
	var failure pgtype.Text
	var updatedAt time.Time

	err := repository.pool.QueryRow(ctx, `
		SELECT asset_id::text, generation, state, language, segments, failure, updated_at
		FROM asr_drafts WHERE lesson_id = $1::text::uuid`,
		lessonID).Scan(&assetID, &generation, &state, &language, &segmentsJSON, &failure, &updatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return result, nil
	}
	if err != nil {
		return ports.LessonDraft{}, persistenceError(err)
	}

	segments, err := decodeSegments(segmentsJSON)
	if err != nil {
		return ports.LessonDraft{}, err
	}

	draft, err := domain.NewTranscriptDraft(domain.TranscriptDraft{
		LessonID:   lessonID,
		AssetID:    assetID,
		Generation: generation,
		State:      domain.DraftState(state),
		Language:   domain.Language(language.String),
		Segments:   segments,
		Failure:    failure.String,
	})
	if err != nil {
		return ports.LessonDraft{}, err
	}

	if err := repository.pool.QueryRow(ctx, readyGenerationForLesson, lessonID).Scan(&result.ReadyGeneration); err != nil {
		return ports.LessonDraft{}, persistenceError(err)
	}

	result.Found = true
	result.Draft = draft
	result.UpdatedAt = updatedAt.UTC().Format(time.RFC3339Nano)
	return result, nil
}

func (repository *Repository) PendingRecognitions(ctx context.Context, limit int) ([]ports.PendingRecognition, error) {
	rows, err := repository.pool.Query(ctx, `
		SELECT assets.lesson_id::text, assets.id::text, assets.generation, assets.duration_seconds
		FROM video_assets AS assets
		LEFT JOIN asr_drafts AS drafts ON drafts.lesson_id = assets.lesson_id
		WHERE assets.state = 'ready' AND assets.evicted_at IS NULL
		  AND (drafts.lesson_id IS NULL OR drafts.generation <> assets.generation)
		ORDER BY assets.lesson_id
		LIMIT $1`, limit)
	if err != nil {
		return nil, persistenceError(err)
	}
	defer rows.Close()

	var pending []ports.PendingRecognition
	for rows.Next() {
		var p ports.PendingRecognition
		var duration pgtype.Float8
		if err := rows.Scan(&p.LessonID, &p.AssetID, &p.Generation, &duration); err != nil {
			return nil, persistenceError(err)
		}
		if duration.Valid {
			p.DurationSeconds = &duration.Float64
		}
		pending = append(pending, p)
	}
	return pending, rows.Err()
}

func (repository *Repository) Claim(ctx context.Context, lessonID, assetID string, generation int32, now time.Time) (ports.ClaimedDraft, error) {
	var exists bool
	err := repository.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM video_assets
			WHERE lesson_id = $1::text::uuid AND id = $2::text::uuid AND generation = $3
		)`, lessonID, assetID, generation).Scan(&exists)
	if err != nil {
		return ports.ClaimedDraft{}, persistenceError(err)
	}
	if !exists {
		return ports.ClaimedDraft{}, domain.ErrNotFound
	}

	var state string
	var updatedAt time.Time
	err = repository.pool.QueryRow(ctx, `
		INSERT INTO asr_drafts (lesson_id, asset_id, generation, state, segments, created_at, updated_at)
		VALUES ($1::text::uuid, $2::text::uuid, $3, 'running', '[]'::jsonb, $4, $4)
		ON CONFLICT (lesson_id) DO UPDATE
			SET asset_id = EXCLUDED.asset_id,
			    generation = EXCLUDED.generation,
			    state = 'running',
			    language = NULL,
			    segments = '[]'::jsonb,
			    failure = NULL,
			    updated_at = EXCLUDED.updated_at
		RETURNING state, updated_at`,
		lessonID, assetID, generation, now).Scan(&state, &updatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return ports.ClaimedDraft{}, domain.ErrNotFound
		}
		return ports.ClaimedDraft{}, persistenceError(err)
	}

	return ports.ClaimedDraft{
		Claimed: true,
		Draft: domain.TranscriptDraft{
			LessonID:   lessonID,
			AssetID:    assetID,
			Generation: generation,
			State:      domain.DraftRunning,
		},
		Token: updatedAt.UTC().Format(time.RFC3339Nano),
	}, nil
}

func (repository *Repository) SaveDraft(ctx context.Context, draft domain.TranscriptDraft, expectedUpdatedAt string, now time.Time) (string, error) {
	segmentsJSON, err := encodeSegments(draft.Segments)
	if err != nil {
		return "", err
	}

	var language *string
	if draft.Language != "" {
		s := string(draft.Language)
		language = &s
	}

	var updatedAt time.Time
	err = repository.pool.QueryRow(ctx, `
		UPDATE asr_drafts
		SET language = $2, segments = $3, updated_at = $4
		WHERE lesson_id = $1::text::uuid AND updated_at = $5
		RETURNING updated_at`,
		draft.LessonID, language, segmentsJSON, now, expectedUpdatedAt).Scan(&updatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", domain.ErrTranscriptModified
	}
	if err != nil {
		return "", persistenceError(err)
	}
	return updatedAt.UTC().Format(time.RFC3339Nano), nil
}

func (repository *Repository) Complete(ctx context.Context, draft domain.TranscriptDraft, claimToken string, now time.Time) error {
	segmentsJSON, err := encodeSegments(draft.Segments)
	if err != nil {
		return err
	}

	var language *string
	if draft.Language != "" {
		s := string(draft.Language)
		language = &s
	}
	var failure *string
	if draft.Failure != "" {
		failure = &draft.Failure
	}

	tag, err := repository.pool.Exec(ctx, `
		UPDATE asr_drafts
		SET state = $2, language = $3, segments = $4, failure = $5, updated_at = $6
		WHERE lesson_id = $1::text::uuid AND updated_at = $7`,
		draft.LessonID, string(draft.State), language, segmentsJSON, failure, now, claimToken)
	if err != nil {
		return persistenceError(err)
	}
	if tag.RowsAffected() == 0 {
		var exists bool
		if err := repository.pool.QueryRow(ctx,
			`SELECT EXISTS (SELECT 1 FROM asr_drafts WHERE lesson_id = $1::text::uuid)`,
			draft.LessonID).Scan(&exists); err != nil {
			return persistenceError(err)
		}
		if !exists {
			return domain.ErrNotFound
		}
		return domain.ErrDraftSuperseded
	}
	return nil
}

func (repository *Repository) Recover(ctx context.Context, olderThan time.Time, limit int) ([]domain.TranscriptDraft, error) {
	rows, err := repository.pool.Query(ctx, `
		SELECT lesson_id::text, asset_id::text, generation
		FROM asr_drafts
		WHERE state = 'running' AND updated_at < $1
		ORDER BY updated_at
		LIMIT $2`, olderThan, limit)
	if err != nil {
		return nil, persistenceError(err)
	}
	defer rows.Close()

	var drafts []domain.TranscriptDraft
	for rows.Next() {
		var d domain.TranscriptDraft
		if err := rows.Scan(&d.LessonID, &d.AssetID, &d.Generation); err != nil {
			return nil, persistenceError(err)
		}
		d.State = domain.DraftRunning
		drafts = append(drafts, d)
	}
	return drafts, rows.Err()
}

type storedSegment struct {
	ID         string   `json:"id"`
	StartMs    int64    `json:"startMs"`
	EndMs      int64    `json:"endMs"`
	Text       string   `json:"text"`
	Speaker    string   `json:"speaker,omitempty"`
	Confidence *float64 `json:"confidence,omitempty"`
}

func decodeSegments(data []byte) ([]domain.Segment, error) {
	var stored []storedSegment
	if err := json.Unmarshal(data, &stored); err != nil {
		return nil, domain.ErrDependencyUnavailable
	}
	segments := make([]domain.Segment, len(stored))
	for i, s := range stored {
		segments[i] = domain.Segment{
			ID: s.ID, StartMs: s.StartMs, EndMs: s.EndMs,
			Text: s.Text, Speaker: s.Speaker, Confidence: s.Confidence,
		}
	}
	return segments, nil
}

func encodeSegments(segments []domain.Segment) ([]byte, error) {
	stored := make([]storedSegment, len(segments))
	for i, s := range segments {
		stored[i] = storedSegment{
			ID: s.ID, StartMs: s.StartMs, EndMs: s.EndMs,
			Text: s.Text, Speaker: s.Speaker, Confidence: s.Confidence,
		}
	}
	return json.Marshal(stored)
}
