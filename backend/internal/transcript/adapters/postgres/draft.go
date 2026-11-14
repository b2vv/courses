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
	var sourceLeaf pgtype.Text
	var updatedAt time.Time

	err := repository.pool.QueryRow(ctx, `
		SELECT asset_id::text, generation, state, language, segments, failure, source_leaf, updated_at
		FROM asr_drafts WHERE lesson_id = $1::text::uuid`,
		lessonID).Scan(&assetID, &generation, &state, &language, &segmentsJSON, &failure, &sourceLeaf, &updatedAt)
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
		SourceLeaf: sourceLeaf.String,
	})
	if err != nil {
		return ports.LessonDraft{}, err
	}

	if err := repository.pool.QueryRow(ctx, readyAudioForLesson, lessonID).Scan(&result.ReadyGeneration, &result.DefaultLeaf); err != nil {
		return ports.LessonDraft{}, persistenceError(err)
	}

	result.Found = true
	result.Draft = draft
	result.UpdatedAt = updatedAt.UTC().Format(time.RFC3339Nano)
	return result, nil
}

func (repository *Repository) PendingRecognitions(ctx context.Context, limit int) ([]ports.PendingRecognition, error) {
	rows, err := repository.pool.Query(ctx, `
		WITH current_assets AS (
			SELECT DISTINCT ON (assets.lesson_id)
			       assets.lesson_id, assets.id, assets.generation, assets.duration_seconds
			FROM video_assets AS assets
			WHERE assets.lesson_id IS NOT NULL
			  AND assets.state = 'ready' AND assets.evicted_at IS NULL
			ORDER BY assets.lesson_id, assets.generation DESC
		)
		SELECT assets.lesson_id::text, assets.id::text, assets.generation, assets.duration_seconds,
		       drafts.updated_at
		FROM current_assets AS assets
		-- Recognition starts only once a default is assigned (D5); its leaf is what must be heard,
		-- so a draft that heard another leaf is re-recognised (multi-audio S6, D-S6-3).
		JOIN audio_tracks AS tracks
		  ON tracks.video_asset_id = assets.id AND tracks.is_default AND tracks.state = 'ready'
		LEFT JOIN asr_drafts AS drafts ON drafts.lesson_id = assets.lesson_id
		WHERE drafts.lesson_id IS NULL OR drafts.generation <> assets.generation
		   OR COALESCE(drafts.source_leaf, 'original') <> tracks.source_leaf
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
		var updatedAt pgtype.Timestamptz
		if err := rows.Scan(&p.LessonID, &p.AssetID, &p.Generation, &duration, &updatedAt); err != nil {
			return nil, persistenceError(err)
		}
		if duration.Valid {
			p.DurationSeconds = &duration.Float64
		}
		if updatedAt.Valid {
			p.TaskMarker = updatedAt.Time.UTC().Format(time.RFC3339Nano)
		}
		pending = append(pending, p)
	}
	return pending, rows.Err()
}

func (repository *Repository) Claim(ctx context.Context, lessonID, assetID string, generation int32, requested bool, now time.Time) (ports.ClaimedDraft, error) {
	var exists bool
	// The asset must have a default Ready audio track: recognition starts only once one is
	// assigned (D5), and its leaf is what the claim records (multi-audio S6, D-S6-2).
	err := repository.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM video_assets AS assets
			JOIN audio_tracks AS tracks
			  ON tracks.video_asset_id = assets.id AND tracks.is_default AND tracks.state = 'ready'
			WHERE assets.lesson_id = $1::text::uuid AND assets.id = $2::text::uuid AND assets.generation = $3
		)`, lessonID, assetID, generation).Scan(&exists)
	if err != nil {
		return ports.ClaimedDraft{}, persistenceError(err)
	}
	if !exists {
		return ports.ClaimedDraft{}, domain.ErrNotFound
	}

	var state, sourceLeaf string
	var updatedAt time.Time
	// The default leaf is read by the same statement that claims, so the recorded leaf is the one
	// the extraction will hear. If the default vanishes in between, nothing is inserted and the
	// claim simply does not happen; the next sweep asks again.
	err = repository.pool.QueryRow(ctx, `
		WITH heard AS (
			SELECT tracks.source_leaf
			FROM audio_tracks AS tracks
			WHERE tracks.video_asset_id = $2::text::uuid AND tracks.is_default AND tracks.state = 'ready'
		)
		INSERT INTO asr_drafts (lesson_id, asset_id, generation, state, segments, source_leaf, created_at, updated_at)
		SELECT $1::text::uuid, $2::text::uuid, $3, 'running', '[]'::jsonb, heard.source_leaf, $4, $4
		FROM heard
		ON CONFLICT (lesson_id) DO UPDATE
			SET asset_id = EXCLUDED.asset_id,
			    generation = EXCLUDED.generation,
			    state = 'running',
			    language = NULL,
			    segments = '[]'::jsonb,
			    failure = NULL,
			    source_leaf = EXCLUDED.source_leaf,
			    updated_at = EXCLUDED.updated_at
			-- Keep the invariant at the write site so it does not depend on queue ordering (Concurrency 1, FIFO).
			-- A same-generation draft is re-claimed only by the author's Regenerate or when it heard
			-- another leaf than the current default; a duplicate automatic task is refused (D-S6-2).
			WHERE asr_drafts.generation < EXCLUDED.generation
			   OR (asr_drafts.generation = EXCLUDED.generation
			       AND ($5 OR COALESCE(asr_drafts.source_leaf, 'original') <> EXCLUDED.source_leaf))
		RETURNING state, source_leaf, updated_at`,
		lessonID, assetID, generation, now, requested).Scan(&state, &sourceLeaf, &updatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return ports.ClaimedDraft{}, domain.ErrNotFound
		}
		if errors.Is(err, pgx.ErrNoRows) {
			return ports.ClaimedDraft{Claimed: false}, nil
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
			SourceLeaf: sourceLeaf,
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

func (repository *Repository) Recover(ctx context.Context, now time.Time, limit int) ([]ports.RecoveredDraft, error) {
	rows, err := repository.pool.Query(ctx, `
		SELECT d.lesson_id::text, d.asset_id::text, d.generation, d.updated_at
		FROM asr_drafts d
		JOIN video_assets va ON va.id = d.asset_id
		WHERE d.state = 'running'
		  AND d.updated_at < $1::timestamptz - (COALESCE(va.duration_seconds, 0) * interval '1 second' + interval '300 seconds')
		ORDER BY d.updated_at
		LIMIT $2`, now, limit)
	if err != nil {
		return nil, persistenceError(err)
	}
	defer rows.Close()

	var drafts []ports.RecoveredDraft
	for rows.Next() {
		var d ports.RecoveredDraft
		var updatedAt time.Time
		if err := rows.Scan(&d.Draft.LessonID, &d.Draft.AssetID, &d.Draft.Generation, &updatedAt); err != nil {
			return nil, persistenceError(err)
		}
		d.Draft.State = domain.DraftRunning
		d.Token = updatedAt.UTC().Format(time.RFC3339Nano)
		drafts = append(drafts, d)
	}
	return drafts, rows.Err()
}

func (repository *Repository) LessonAsset(ctx context.Context, lessonID string) (*ports.PendingRecognition, error) {
	var p ports.PendingRecognition
	var duration pgtype.Float8
	err := repository.pool.QueryRow(ctx, `
		SELECT id::text, generation, duration_seconds
		FROM video_assets
		WHERE lesson_id = $1::text::uuid AND state = 'ready' AND evicted_at IS NULL
		ORDER BY generation DESC
		LIMIT 1`,
		lessonID).Scan(&p.AssetID, &p.Generation, &duration)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, persistenceError(err)
	}
	p.LessonID = lessonID
	if duration.Valid {
		p.DurationSeconds = &duration.Float64
	}
	return &p, nil
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
