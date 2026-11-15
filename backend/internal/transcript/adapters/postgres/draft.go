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

// ListLessonDrafts joins current track metadata separately from the claim snapshot: metadata
// changes make output stale, while replacement hides old-asset drafts entirely (MS-D3/MS-D5).
func (repository *Repository) ListLessonDrafts(ctx context.Context, lessonID string) ([]ports.LessonDraft, error) {
	rows, err := repository.pool.Query(ctx, `
  SELECT d.asset_id::text, d.generation, d.audio_track_id::text, d.track_language,
         d.track_multilingual, d.state, d.language, d.segments, d.failure, t.source_leaf,
         d.updated_at, t.language, t.multilingual, t.is_default
  FROM asr_drafts AS d
  JOIN audio_tracks AS t ON t.id = d.audio_track_id
  WHERE d.lesson_id = $1::text::uuid AND d.asset_id = (`+latestReadyAsset+`)
  ORDER BY t.created_at, t.id`, lessonID)
	if err != nil {
		return nil, persistenceError(err)
	}
	defer rows.Close()
	result := make([]ports.LessonDraft, 0)
	for rows.Next() {
		var item ports.LessonDraft
		var raw domain.TranscriptDraft
		var state, currentLanguage string
		var language, failure pgtype.Text
		var segmentsJSON []byte
		var updatedAt time.Time
		raw.LessonID = lessonID
		if err := rows.Scan(&raw.AssetID, &raw.Generation, &raw.TrackID, &raw.TrackLanguage,
			&raw.TrackMultilingual, &state, &language, &segmentsJSON, &failure, &raw.SourceLeaf,
			&updatedAt, &currentLanguage, &item.TrackMultilingual, &item.IsDefault); err != nil {
			return nil, persistenceError(err)
		}
		raw.State, raw.Language, raw.Failure = domain.DraftState(state), domain.Language(language.String), failure.String
		raw.Segments, err = decodeSegments(segmentsJSON)
		if err != nil {
			return nil, err
		}
		item.Draft, err = domain.NewTranscriptDraft(raw)
		if err != nil {
			return nil, err
		}
		item.TrackLanguage, err = domain.NewLanguage(currentLanguage)
		if err != nil {
			return nil, err
		}
		item.UpdatedAt = updatedAt.UTC().Format(time.RFC3339Nano)
		result = append(result, item)
	}
	if err := rows.Err(); err != nil {
		return nil, persistenceError(err)
	}
	return result, nil
}

// latestReadyAsset deliberately filters Ready/non-evicted before ordering: an upload or a failed
// transcode must not displace the playable video. $1 is always the lesson id in these queries.
const latestReadyAsset = `SELECT id FROM video_assets
 WHERE lesson_id = $1::text::uuid AND state = 'ready' AND evicted_at IS NULL
 ORDER BY generation DESC LIMIT 1`

func (repository *Repository) PendingRecognitions(ctx context.Context, limit int) ([]ports.PendingRecognition, error) {
	rows, err := repository.pool.Query(ctx, `
  WITH current_assets AS (
   SELECT DISTINCT ON (lesson_id) lesson_id, id, generation, duration_seconds
   FROM video_assets
   WHERE lesson_id IS NOT NULL AND state = 'ready' AND evicted_at IS NULL
   ORDER BY lesson_id, generation DESC
  )
  SELECT a.lesson_id::text, a.id::text, a.generation, t.id::text, a.duration_seconds
  FROM current_assets AS a
  JOIN audio_tracks AS t ON t.video_asset_id = a.id AND t.state = 'ready'
  LEFT JOIN asr_drafts AS d ON d.audio_track_id = t.id
  -- O7: even a failed or stale draft needs the author's explicit Regenerate. Only absence
  -- triggers automatic recognition, independently for each track (default is irrelevant).
  WHERE d.audio_track_id IS NULL
  ORDER BY a.lesson_id, t.id LIMIT $1`, limit)
	if err != nil {
		return nil, persistenceError(err)
	}
	defer rows.Close()
	var pending []ports.PendingRecognition
	for rows.Next() {
		var p ports.PendingRecognition
		var duration pgtype.Float8
		if err := rows.Scan(&p.LessonID, &p.AssetID, &p.Generation, &p.TrackID, &duration); err != nil {
			return nil, persistenceError(err)
		}
		if duration.Valid {
			p.DurationSeconds = &duration.Float64
		}
		pending = append(pending, p)
	}
	if err := rows.Err(); err != nil {
		return nil, persistenceError(err)
	}
	return pending, nil
}

func (repository *Repository) Claim(ctx context.Context, lessonID, assetID string, generation int32, trackID string, requested bool, now time.Time) (ports.ClaimedDraft, error) {
	var language, sourceLeaf string
	var multilingual bool
	var updatedAt time.Time
	// The source pairing, latest-asset guard and metadata snapshot belong to the write itself.
	// A preflight SELECT would let a replacement or metadata edit slip between proof and INSERT.
	// Only requested jobs replace rows: rename/mode changes never bypass the O7 author decision.
	err := repository.pool.QueryRow(ctx, `
  INSERT INTO asr_drafts (lesson_id, asset_id, generation, audio_track_id, track_language,
                         track_multilingual, state, segments, created_at, updated_at)
  SELECT a.lesson_id, a.id, a.generation, t.id, lower(t.language), t.multilingual,
         'running', '[]'::jsonb, $6, $6
  FROM audio_tracks AS t JOIN video_assets AS a ON a.id = t.video_asset_id
  WHERE a.lesson_id = $1::text::uuid AND a.id = $2::text::uuid AND a.generation = $3
    AND t.id = $4::text::uuid AND t.state = 'ready'
    AND a.id = (`+latestReadyAsset+`)
  ON CONFLICT (audio_track_id) DO UPDATE
   SET state = 'running', language = NULL, segments = '[]'::jsonb, failure = NULL,
       track_language = EXCLUDED.track_language,
       track_multilingual = EXCLUDED.track_multilingual, updated_at = EXCLUDED.updated_at
   WHERE $5
  RETURNING track_language, track_multilingual,
            (SELECT source_leaf FROM audio_tracks WHERE id = asr_drafts.audio_track_id), updated_at`,
		lessonID, assetID, generation, trackID, requested, now).Scan(&language, &multilingual, &sourceLeaf, &updatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return ports.ClaimedDraft{}, domain.ErrNotFound
		}
		if !errors.Is(err, pgx.ErrNoRows) {
			return ports.ClaimedDraft{}, persistenceError(err)
		}
		// An empty RETURNING can mean a refused duplicate or an invalid source. Diagnose only
		// after the atomic write, never use this read as authorisation for a later INSERT (C21).
		var current bool
		err = repository.pool.QueryRow(ctx, `
   SELECT COALESCE(a.id = (`+latestReadyAsset+`), false)
   FROM audio_tracks AS t JOIN video_assets AS a ON a.id = t.video_asset_id
   WHERE a.lesson_id = $1::text::uuid AND a.id = $2::text::uuid AND a.generation = $3
     AND t.id = $4::text::uuid AND t.state = 'ready'`, lessonID, assetID, generation, trackID).Scan(&current)
		if errors.Is(err, pgx.ErrNoRows) {
			return ports.ClaimedDraft{}, domain.ErrNotFound
		}
		if err != nil {
			return ports.ClaimedDraft{}, persistenceError(err)
		}
		if !current {
			return ports.ClaimedDraft{}, domain.ErrDraftSuperseded
		}
		return ports.ClaimedDraft{Claimed: false}, nil
	}
	return ports.ClaimedDraft{
		Claimed: true,
		Draft: domain.TranscriptDraft{LessonID: lessonID, AssetID: assetID, Generation: generation,
			TrackID: trackID, TrackLanguage: domain.Language(language), TrackMultilingual: multilingual,
			State: domain.DraftRunning, SourceLeaf: sourceLeaf},
		Token: updatedAt.UTC().Format(time.RFC3339Nano),
	}, nil
}

func (repository *Repository) Complete(ctx context.Context, draft domain.TranscriptDraft, claimToken string, now time.Time) error {
	segmentsJSON, err := encodeSegments(draft.Segments)
	if err != nil {
		return err
	}
	var language, failure *string
	if draft.Language != "" {
		s := string(draft.Language)
		language = &s
	}
	if draft.Failure != "" {
		failure = &draft.Failure
	}
	// The attempt token fences recovery/re-claim races on this track. Another track's attempt
	// cannot invalidate it, and a slow recovered worker cannot replace the new attempt's text.
	tag, err := repository.pool.Exec(ctx, `
  UPDATE asr_drafts SET state=$2, language=$3, segments=$4, failure=$5, updated_at=$6
  WHERE audio_track_id=$1::text::uuid AND updated_at=$7`,
		draft.TrackID, string(draft.State), language, segmentsJSON, failure, now, claimToken)
	if err != nil {
		return persistenceError(err)
	}
	if tag.RowsAffected() == 0 {
		var exists bool
		if err := repository.pool.QueryRow(ctx,
			`SELECT EXISTS (SELECT 1 FROM asr_drafts WHERE audio_track_id=$1::text::uuid)`, draft.TrackID).Scan(&exists); err != nil {
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
  SELECT d.lesson_id::text, d.asset_id::text, d.generation, d.audio_track_id::text,
         d.track_language, d.track_multilingual, t.source_leaf, d.updated_at
  FROM asr_drafts AS d
  JOIN video_assets AS a ON a.id = d.asset_id
  JOIN audio_tracks AS t ON t.id = d.audio_track_id
  WHERE d.state='running'
    AND d.updated_at < $1::timestamptz - (COALESCE(a.duration_seconds,0) * interval '1 second' + interval '300 seconds')
  ORDER BY d.updated_at, d.audio_track_id LIMIT $2`, now, limit)
	if err != nil {
		return nil, persistenceError(err)
	}
	defer rows.Close()
	var drafts []ports.RecoveredDraft
	for rows.Next() {
		var d ports.RecoveredDraft
		var updatedAt time.Time
		if err := rows.Scan(&d.Draft.LessonID, &d.Draft.AssetID, &d.Draft.Generation, &d.Draft.TrackID,
			&d.Draft.TrackLanguage, &d.Draft.TrackMultilingual, &d.Draft.SourceLeaf, &updatedAt); err != nil {
			return nil, persistenceError(err)
		}
		d.Draft.State = domain.DraftRunning
		d.Token = updatedAt.UTC().Format(time.RFC3339Nano)
		drafts = append(drafts, d)
	}
	if err := rows.Err(); err != nil {
		return nil, persistenceError(err)
	}
	return drafts, nil
}

func (repository *Repository) LessonTrack(ctx context.Context, lessonID, trackID string) (*ports.PendingRecognition, error) {
	var p ports.PendingRecognition
	var duration pgtype.Float8
	err := repository.pool.QueryRow(ctx, `
  SELECT a.id::text, a.generation, t.id::text, a.duration_seconds
  FROM video_assets AS a JOIN audio_tracks AS t ON t.video_asset_id = a.id
  WHERE a.id = (`+latestReadyAsset+`) AND t.state = 'ready'
    AND (($2 = '' AND t.is_default) OR t.id = NULLIF($2, '')::uuid)`,
		lessonID, trackID).Scan(&p.AssetID, &p.Generation, &p.TrackID, &duration)
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
