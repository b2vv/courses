package postgres

import (
	"context"
	"errors"
	"time"

	"courses/internal/media/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

const audioTrackColumns = `
id::text,
video_asset_id::text,
kind,
language,
label,
is_default,
state,
source_leaf,
created_at,
updated_at`

// CreateAudioTrack persists a track already validated by the domain constructor.
// The caller mints the id (the uploaded leaf audio-<id> depends on it, D3), so
// no id is generated here. A missing asset maps to ErrVideoNotFound; a second
// default or a broken invariant surfaces as ErrInvalidRequest from the DB guards.
func (repository *Repository) CreateAudioTrack(ctx context.Context, track domain.AudioTrack) (domain.AudioTrack, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return domain.AudioTrack{}, persistenceError(err)
	}
	defer rollback(tx)
	created := storeTime(track.CreatedAt)
	updated := storeTime(track.UpdatedAt)
	stored, err := scanAudioTrack(tx.QueryRow(ctx, `INSERT INTO audio_tracks(
		id,video_asset_id,kind,language,label,is_default,state,source_leaf,created_at,updated_at)
		VALUES($1::text::uuid,$2::text::uuid,$3,$4,$5,$6,$7,$8,$9,$10)
		RETURNING `+audioTrackColumns,
		track.ID, track.VideoAssetID, string(track.Kind), track.Language, track.Label,
		track.IsDefault, string(track.State), track.SourceLeaf, created, updated))
	if err != nil {
		if mapped, ok := audioTrackConflict(err); ok {
			return domain.AudioTrack{}, mapped
		}
		return domain.AudioTrack{}, persistenceError(err)
	}
	if err := commit(ctx, tx); err != nil {
		return domain.AudioTrack{}, err
	}
	return stored, nil
}

// ListAudioTracks returns every track of an asset in the deterministic
// (created_at, id) order the domain's default election relies on.
func (repository *Repository) ListAudioTracks(ctx context.Context, assetID string) ([]domain.AudioTrack, error) {
	rows, err := repository.pool.Query(ctx, `SELECT `+audioTrackColumns+`
		FROM audio_tracks WHERE video_asset_id=$1::text::uuid
		ORDER BY created_at, id`, assetID)
	if err != nil {
		return nil, persistenceError(err)
	}
	defer rows.Close()
	tracks := make([]domain.AudioTrack, 0)
	for rows.Next() {
		track, err := scanAudioTrack(rows)
		if err != nil {
			return nil, err
		}
		tracks = append(tracks, track)
	}
	if err := rows.Err(); err != nil {
		return nil, persistenceError(err)
	}
	return tracks, nil
}

// SetDefaultAudioTrack makes trackID the sole default of its asset in one
// serialized step (D4). The target must exist and be Ready; the swap clears the
// previous default and sets the new one in a single statement, so the
// one-default-per-asset index never sees a transient second default.
func (repository *Repository) SetDefaultAudioTrack(ctx context.Context, assetID, trackID string, now time.Time) error {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return persistenceError(err)
	}
	defer rollback(tx)
	now = storeTime(now)
	if _, err := tx.Exec(ctx, `SELECT 1 FROM audio_tracks
		WHERE video_asset_id=$1::text::uuid FOR UPDATE`, assetID); err != nil {
		return persistenceError(err)
	}
	var state string
	err = tx.QueryRow(ctx, `SELECT state FROM audio_tracks
		WHERE id=$1::text::uuid AND video_asset_id=$2::text::uuid`, trackID, assetID).Scan(&state)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrVideoNotFound
	}
	if err != nil {
		return persistenceError(err)
	}
	if domain.State(state) != domain.Ready {
		return domain.ErrInvalidRequest
	}
	// Clear the current default, then promote the target, as two statements under
	// the FOR UPDATE lock. A single `is_default=(id=$2)` statement could present
	// two defaults mid-scan (the classic rank-swap), which the non-deferrable
	// partial unique index rejects depending on row order; 1→0 then 0→1 never does.
	if _, err := tx.Exec(ctx, `UPDATE audio_tracks
		SET is_default=false, updated_at=$2
		WHERE video_asset_id=$1::text::uuid AND is_default AND id<>$3::text::uuid`, assetID, now, trackID); err != nil {
		return persistenceError(err)
	}
	if _, err := tx.Exec(ctx, `UPDATE audio_tracks
		SET is_default=true, updated_at=$2
		WHERE id=$3::text::uuid AND video_asset_id=$1::text::uuid AND NOT is_default`, assetID, now, trackID); err != nil {
		return persistenceError(err)
	}
	return commit(ctx, tx)
}

// DeleteAudioTrack removes a track, refusing to drop the last Ready track of a
// Ready asset and re-electing a new default when the deleted track was the
// default (D4). Election is deterministic (first Ready by created_at, id).
func (repository *Repository) DeleteAudioTrack(ctx context.Context, assetID, trackID string, now time.Time) error {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return persistenceError(err)
	}
	defer rollback(tx)
	now = storeTime(now)
	var assetState string
	err = tx.QueryRow(ctx, `SELECT state FROM video_assets
		WHERE id=$1::text::uuid AND evicted_at IS NULL FOR UPDATE`, assetID).Scan(&assetState)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrVideoNotFound
	}
	if err != nil {
		return persistenceError(err)
	}
	tracks, err := lockAssetTracks(ctx, tx, assetID)
	if err != nil {
		return err
	}
	if err := domain.ValidateTrackDeletion(tracks, trackID, domain.State(assetState)); err != nil {
		return err
	}
	wasDefault := false
	remaining := make([]domain.AudioTrack, 0, len(tracks))
	for _, track := range tracks {
		if track.ID == trackID {
			wasDefault = track.IsDefault
			continue
		}
		remaining = append(remaining, track)
	}
	if _, err := tx.Exec(ctx, `DELETE FROM audio_tracks
		WHERE id=$1::text::uuid AND video_asset_id=$2::text::uuid`, trackID, assetID); err != nil {
		return persistenceError(err)
	}
	if wasDefault {
		if elected := domain.ElectDefault(remaining); elected != "" {
			if _, err := tx.Exec(ctx, `UPDATE audio_tracks
				SET is_default=true, updated_at=$2
				WHERE id=$1::text::uuid`, elected, now); err != nil {
				return persistenceError(err)
			}
		}
	}
	return commit(ctx, tx)
}

// RegenerateMaster serialises a master rewrite per asset (D7). It locks the asset
// row, then its tracks, reads the Ready set in deterministic (created_at, id)
// order and hands it to rewrite while still holding the locks, so two concurrent
// audio transcodes can never write the master at once. rewrite performs the
// atomic file write; its error aborts the transaction.
func (repository *Repository) RegenerateMaster(ctx context.Context, assetID string, rewrite func(ready []domain.AudioTrack) error) error {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return persistenceError(err)
	}
	defer rollback(tx)
	var exists int
	err = tx.QueryRow(ctx, `SELECT 1 FROM video_assets
		WHERE id=$1::text::uuid AND evicted_at IS NULL FOR UPDATE`, assetID).Scan(&exists)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrVideoNotFound
	}
	if err != nil {
		return persistenceError(err)
	}
	tracks, err := lockAssetTracks(ctx, tx, assetID)
	if err != nil {
		return err
	}
	ready := make([]domain.AudioTrack, 0, len(tracks))
	for _, track := range tracks {
		if track.State == domain.Ready {
			ready = append(ready, track)
		}
	}
	if err := rewrite(ready); err != nil {
		return err
	}
	return commit(ctx, tx)
}

// LoadAudioTrack returns a single track of an asset, or ErrVideoNotFound.
func (repository *Repository) LoadAudioTrack(ctx context.Context, assetID, trackID string) (domain.AudioTrack, error) {
	track, err := scanAudioTrack(repository.pool.QueryRow(ctx, `SELECT `+audioTrackColumns+`
		FROM audio_tracks WHERE id=$1::text::uuid AND video_asset_id=$2::text::uuid`, trackID, assetID))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.AudioTrack{}, domain.ErrVideoNotFound
	}
	if err != nil {
		return domain.AudioTrack{}, persistenceError(err)
	}
	return track, nil
}

// PromoteAudioTrack marks an uploaded track Ready, elects it the default when the
// asset has no default yet, lifts the asset from AwaitingAudio to Ready (T11), and
// regenerates the master from the resulting Ready set — all under the asset +
// track locks in one transaction (D4/D7).
func (repository *Repository) PromoteAudioTrack(ctx context.Context, assetID, trackID string, now time.Time, rewrite func(ready []domain.AudioTrack) error) (evicted *domain.VideoAsset, err error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return nil, persistenceError(err)
	}
	defer rollback(tx)
	now = storeTime(now)
	var assetState, lessonID string
	err = tx.QueryRow(ctx, `SELECT state, lesson_id FROM video_assets
		WHERE id=$1::text::uuid AND evicted_at IS NULL FOR UPDATE`, assetID).Scan(&assetState, &lessonID)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrVideoNotFound
	}
	if err != nil {
		return nil, persistenceError(err)
	}
	tracks, err := lockAssetTracks(ctx, tx, assetID)
	if err != nil {
		return nil, err
	}
	index := -1
	hasDefault := false
	for position, track := range tracks {
		if track.ID == trackID {
			index = position
		}
		if track.IsDefault {
			hasDefault = true
		}
	}
	if index < 0 {
		return nil, domain.ErrVideoNotFound
	}
	if _, err := tx.Exec(ctx, `UPDATE audio_tracks SET state='ready', updated_at=$2
		WHERE id=$1::text::uuid AND video_asset_id=$3::text::uuid`, trackID, now, assetID); err != nil {
		if mapped, ok := audioTrackConflict(err); ok {
			return nil, mapped
		}
		return nil, persistenceError(err)
	}
	tracks[index].State = domain.Ready
	// A video-only asset gains its first default here; an asset with an embedded
	// (already default) track keeps it.
	if !hasDefault {
		if elected := domain.ElectDefault(tracks); elected != "" {
			if _, err := tx.Exec(ctx, `UPDATE audio_tracks SET is_default=true, updated_at=$2
				WHERE id=$1::text::uuid`, elected, now); err != nil {
				if mapped, ok := audioTrackConflict(err); ok {
					return nil, mapped
				}
				return nil, persistenceError(err)
			}
			for position := range tracks {
				tracks[position].IsDefault = tracks[position].ID == elected
			}
		}
	}
	if domain.State(assetState) == domain.AwaitingAudio {
		if _, err := tx.Exec(ctx, `UPDATE video_assets SET state='ready', updated_at=$2
			WHERE id=$1::text::uuid AND state='awaiting_audio' AND evicted_at IS NULL`, assetID, now); err != nil {
			return nil, persistenceError(err)
		}
		// The asset only now joins its lesson's Ready set, so the per-lesson cap can
		// be exceeded here exactly as it can in Publish. Mirror Publish's eviction
		// (D4): when the set is over the cap, evict the oldest generation so a
		// re-uploaded video-only lesson does not accumulate Ready bytes.
		//
		// awaiting_audio is NOT in the pending-per-lesson unique index (00005), so a
		// stale awaiting_audio generation can linger while newer generations are
		// uploaded and published Ready; when its audio finally lands it is the oldest,
		// and ORDER BY generation can select the just-promoted asset itself. That
		// self-eviction is cap-consistent (keeps the newest two) and cannot strand or
		// corrupt, but it wastes this promotion's master rewrite. The audio-upload
		// path that makes this reachable is not wired yet (S5-L1); whether a
		// superseded generation should be promoted at all is decided when it is.
		var readyCount int
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM video_assets
			WHERE lesson_id=$1::text::uuid AND state='ready' AND evicted_at IS NULL`, lessonID).Scan(&readyCount); err != nil {
			return nil, persistenceError(err)
		}
		if readyCount > 2 {
			victim, err := scanAsset(tx.QueryRow(ctx, `UPDATE video_assets AS assets
				SET evicted_at=$2, updated_at=$2
				WHERE id=(
					SELECT id FROM video_assets
					WHERE lesson_id=$1::text::uuid AND state='ready' AND evicted_at IS NULL
					ORDER BY generation
					LIMIT 1
				)
				RETURNING `+mediaAssetColumns, lessonID, now))
			if err != nil {
				return nil, persistenceError(err)
			}
			evicted = &victim
		}
	}
	ready := make([]domain.AudioTrack, 0, len(tracks))
	for _, track := range tracks {
		if track.State == domain.Ready {
			ready = append(ready, track)
		}
	}
	if err := rewrite(ready); err != nil {
		return nil, err
	}
	return evicted, commit(ctx, tx)
}

// FailAudioTrack marks the track Failed. When the asset is AwaitingAudio and is
// then left with no Ready track and nothing still in flight, the asset fails too
// (T12): a video-only lesson whose only audio never arrives cannot stay pending.
func (repository *Repository) FailAudioTrack(ctx context.Context, assetID, trackID string, now time.Time) error {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return persistenceError(err)
	}
	defer rollback(tx)
	now = storeTime(now)
	var assetState string
	err = tx.QueryRow(ctx, `SELECT state FROM video_assets
		WHERE id=$1::text::uuid AND evicted_at IS NULL FOR UPDATE`, assetID).Scan(&assetState)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrVideoNotFound
	}
	if err != nil {
		return persistenceError(err)
	}
	tracks, err := lockAssetTracks(ctx, tx, assetID)
	if err != nil {
		return err
	}
	found := false
	for _, track := range tracks {
		if track.ID == trackID {
			found = true
		}
	}
	if !found {
		return domain.ErrVideoNotFound
	}
	if _, err := tx.Exec(ctx, `UPDATE audio_tracks SET state='failed', is_default=false, updated_at=$2
		WHERE id=$1::text::uuid AND video_asset_id=$3::text::uuid`, trackID, now, assetID); err != nil {
		return persistenceError(err)
	}
	anyReady, anyPending := false, false
	for _, track := range tracks {
		if track.ID == trackID {
			continue
		}
		switch track.State {
		case domain.Ready:
			anyReady = true
		case domain.Failed:
		default:
			anyPending = true
		}
	}
	if domain.State(assetState) == domain.AwaitingAudio && !anyReady && !anyPending {
		if _, err := tx.Exec(ctx, `UPDATE video_assets SET state='failed', failure_reason='render_failed', updated_at=$2
			WHERE id=$1::text::uuid AND state='awaiting_audio' AND evicted_at IS NULL`, assetID, now); err != nil {
			return persistenceError(err)
		}
	}
	return commit(ctx, tx)
}

func lockAssetTracks(ctx context.Context, tx pgx.Tx, assetID string) ([]domain.AudioTrack, error) {
	rows, err := tx.Query(ctx, `SELECT `+audioTrackColumns+`
		FROM audio_tracks WHERE video_asset_id=$1::text::uuid
		ORDER BY created_at, id FOR UPDATE`, assetID)
	if err != nil {
		return nil, persistenceError(err)
	}
	defer rows.Close()
	tracks := make([]domain.AudioTrack, 0)
	for rows.Next() {
		track, err := scanAudioTrack(rows)
		if err != nil {
			return nil, err
		}
		tracks = append(tracks, track)
	}
	if err := rows.Err(); err != nil {
		return nil, persistenceError(err)
	}
	return tracks, nil
}

func scanAudioTrack(row pgx.Row) (domain.AudioTrack, error) {
	var (
		id, videoAssetID, kind, language, label, state, sourceLeaf string
		isDefault                                                  bool
		createdAt, updatedAt                                       pgtype.Timestamptz
	)
	if err := row.Scan(&id, &videoAssetID, &kind, &language, &label, &isDefault, &state, &sourceLeaf, &createdAt, &updatedAt); err != nil {
		return domain.AudioTrack{}, err
	}
	return mapAudioTrack(id, videoAssetID, kind, language, label, isDefault, state, sourceLeaf, createdAt, updatedAt)
}

func mapAudioTrack(
	id, videoAssetID, kind, language, label string,
	isDefault bool,
	state, sourceLeaf string,
	createdAt, updatedAt pgtype.Timestamptz,
) (domain.AudioTrack, error) {
	trackKind := domain.AudioTrackKind(kind)
	if trackKind != domain.EmbeddedTrack && trackKind != domain.UploadedTrack {
		return domain.AudioTrack{}, domain.ErrDependencyUnavailable
	}
	if !createdAt.Valid || !updatedAt.Valid || updatedAt.Time.Before(createdAt.Time) {
		return domain.AudioTrack{}, domain.ErrDependencyUnavailable
	}
	if err := domain.ValidateAudioTrackState(trackKind, domain.State(state)); err != nil {
		return domain.AudioTrack{}, domain.ErrDependencyUnavailable
	}
	return domain.AudioTrack{
		ID:           id,
		VideoAssetID: videoAssetID,
		Kind:         trackKind,
		Language:     language,
		Label:        label,
		IsDefault:    isDefault,
		State:        domain.State(state),
		SourceLeaf:   sourceLeaf,
		CreatedAt:    createdAt.Time,
		UpdatedAt:    updatedAt.Time,
	}, nil
}

func audioTrackConflict(err error) (error, bool) {
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) {
		return nil, false
	}
	switch postgresError.Code {
	case "23503": // foreign key: the asset does not exist
		return domain.ErrVideoNotFound, true
	case "23505", "23514": // unique default index, or a check constraint
		return domain.ErrInvalidRequest, true
	}
	return nil, false
}
