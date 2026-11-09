package postgres

import (
	"context"
	"errors"
	"time"

	"courses/internal/media/domain"
	"courses/internal/media/ports"
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
declared_bytes,
received_bytes,
original_name,
expires_at,
created_at,
updated_at`

// ownedAudioAssetOwner reads the course instructor behind an asset without
// locking, for read-only authorisation of the track list.
const ownedAudioAssetOwner = `SELECT courses.instructor_id::text
FROM video_assets AS assets
JOIN course_lessons AS lessons ON lessons.id=assets.lesson_id
JOIN course_modules AS modules ON modules.id=lessons.module_id
JOIN courses ON courses.id=modules.course_id
WHERE assets.id=$1::text::uuid AND assets.evicted_at IS NULL`

// authorizeAudioRead refuses a track-list read the actor may not see: a missing or
// evicted asset is ErrVideoNotFound, a non-owner non-admin is ErrCourseForbidden.
func (repository *Repository) authorizeAudioRead(ctx context.Context, actor domain.Actor, assetID string) error {
	var ownerID string
	err := repository.pool.QueryRow(ctx, ownedAudioAssetOwner, assetID).Scan(&ownerID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrVideoNotFound
	}
	if err != nil {
		return persistenceError(err)
	}
	return actor.CanEditLesson(ownerID)
}

// authorizeAudioEdit locks the owning asset row (FOR UPDATE OF assets, the same
// lock order promote/publish take) and checks the actor may edit its lesson's
// course (D-S5L1-3). It returns the locked asset so a caller can read its state
// without a second query.
func authorizeAudioEdit(ctx context.Context, tx pgx.Tx, actor domain.Actor, assetID string) (domain.VideoAsset, error) {
	locked, ownerID, err := scanOwnedAsset(tx.QueryRow(ctx, lockOwnedMediaAsset, assetID))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.VideoAsset{}, domain.ErrVideoNotFound
	}
	if err != nil {
		return domain.VideoAsset{}, persistenceError(err)
	}
	if err := actor.CanEditLesson(ownerID); err != nil {
		return domain.VideoAsset{}, err
	}
	return locked, nil
}

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
	var expiresAt *time.Time
	if !track.ExpiresAt.IsZero() {
		t := storeTime(track.ExpiresAt)
		expiresAt = &t
	}
	stored, err := scanAudioTrack(tx.QueryRow(ctx, `INSERT INTO audio_tracks(
		id,video_asset_id,kind,language,label,is_default,state,source_leaf,
		declared_bytes,received_bytes,original_name,expires_at,created_at,updated_at)
		VALUES($1::text::uuid,$2::text::uuid,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
		RETURNING `+audioTrackColumns,
		track.ID, track.VideoAssetID, string(track.Kind), track.Language, track.Label,
		track.IsDefault, string(track.State), track.SourceLeaf,
		track.DeclaredBytes, track.ReceivedBytes, track.OriginalName, expiresAt,
		created, updated))
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
func (repository *Repository) ListAudioTracks(ctx context.Context, actor domain.Actor, assetID string) ([]domain.AudioTrack, error) {
	if err := repository.authorizeAudioRead(ctx, actor, assetID); err != nil {
		return nil, err
	}
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
func (repository *Repository) SetDefaultAudioTrack(ctx context.Context, actor domain.Actor, assetID, trackID string, now time.Time, rewrite func(ready []domain.AudioTrack) error) error {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return persistenceError(err)
	}
	defer rollback(tx)
	now = storeTime(now)
	if _, err := authorizeAudioEdit(ctx, tx, actor, assetID); err != nil {
		return err
	}
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
	// The default is DEFAULT=YES on its EXT-X-MEDIA line, so the master changes too.
	if err := rewriteReadyTracks(ctx, tx, assetID, rewrite); err != nil {
		return err
	}
	return commit(ctx, tx)
}

// rewriteReadyTracks hands the asset's current Ready set, read inside the caller's
// transaction (so it reflects the caller's uncommitted change), to rewrite while
// the asset lock is still held (D7). A failing rewrite aborts the transaction, so
// the database never commits a track change the master does not reflect.
func rewriteReadyTracks(ctx context.Context, tx pgx.Tx, assetID string, rewrite func(ready []domain.AudioTrack) error) error {
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
	return rewrite(ready)
}

// DeleteAudioTrack removes a track, refusing to drop the last Ready track of a
// Ready asset and re-electing a new default when the deleted track was the
// default (D4). Election is deterministic (first Ready by created_at, id).
func (repository *Repository) DeleteAudioTrack(ctx context.Context, actor domain.Actor, assetID, trackID string, now time.Time, rewrite func(ready []domain.AudioTrack) error) error {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return persistenceError(err)
	}
	defer rollback(tx)
	now = storeTime(now)
	locked, err := authorizeAudioEdit(ctx, tx, actor, assetID)
	if err != nil {
		return err
	}
	tracks, err := lockAssetTracks(ctx, tx, assetID)
	if err != nil {
		return err
	}
	if err := domain.ValidateTrackDeletion(tracks, trackID, locked.State); err != nil {
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
	// L3 (closes the S3 limitation): deleting a track changes the Ready set, so the
	// master must be regenerated under the same lock before commit, exactly as
	// promote does — otherwise the master keeps a stale EXT-X-MEDIA to the removed
	// track's directory.
	if err := rewriteReadyTracks(ctx, tx, assetID, rewrite); err != nil {
		return err
	}
	return commit(ctx, tx)
}

// UpdateAudioTrackMetadata renames a track (language and/or label) after the
// domain has validated the new values. The (language, asset) unique index rejects
// a clash with another non-failed track as ErrDuplicateLanguage (D6). The target
// must exist under the asset.
func (repository *Repository) UpdateAudioTrackMetadata(ctx context.Context, actor domain.Actor, assetID, trackID, language, label string, now time.Time, rewrite func(ready []domain.AudioTrack) error) (domain.AudioTrack, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return domain.AudioTrack{}, persistenceError(err)
	}
	defer rollback(tx)
	now = storeTime(now)
	if _, err := authorizeAudioEdit(ctx, tx, actor, assetID); err != nil {
		return domain.AudioTrack{}, err
	}
	updated, err := scanAudioTrack(tx.QueryRow(ctx, `UPDATE audio_tracks
		SET language=$3, label=$4, updated_at=$5
		WHERE id=$1::text::uuid AND video_asset_id=$2::text::uuid
		RETURNING `+audioTrackColumns, trackID, assetID, language, label, now))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.AudioTrack{}, domain.ErrVideoNotFound
	}
	if err != nil {
		if mapped, ok := audioTrackConflict(err); ok {
			return domain.AudioTrack{}, mapped
		}
		return domain.AudioTrack{}, persistenceError(err)
	}
	// NAME and LANGUAGE are rendered into the track's EXT-X-MEDIA line (D7).
	if err := rewriteReadyTracks(ctx, tx, assetID, rewrite); err != nil {
		return domain.AudioTrack{}, err
	}
	if err := commit(ctx, tx); err != nil {
		return domain.AudioTrack{}, err
	}
	return updated, nil
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
		declaredBytes, receivedBytes                               int64
		originalName                                               string
		expiresAt                                                  pgtype.Timestamptz
		createdAt, updatedAt                                       pgtype.Timestamptz
	)
	if err := row.Scan(&id, &videoAssetID, &kind, &language, &label, &isDefault, &state, &sourceLeaf,
		&declaredBytes, &receivedBytes, &originalName, &expiresAt, &createdAt, &updatedAt); err != nil {
		return domain.AudioTrack{}, err
	}
	return mapAudioTrack(id, videoAssetID, kind, language, label, isDefault, state, sourceLeaf,
		declaredBytes, receivedBytes, originalName, expiresAt, createdAt, updatedAt)
}

func mapAudioTrack(
	id, videoAssetID, kind, language, label string,
	isDefault bool,
	state, sourceLeaf string,
	declaredBytes, receivedBytes int64,
	originalName string,
	expiresAt pgtype.Timestamptz,
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
	track := domain.AudioTrack{
		ID:            id,
		VideoAssetID:  videoAssetID,
		Kind:          trackKind,
		Language:      language,
		Label:         label,
		IsDefault:     isDefault,
		State:         domain.State(state),
		SourceLeaf:    sourceLeaf,
		DeclaredBytes: declaredBytes,
		ReceivedBytes: receivedBytes,
		OriginalName:  originalName,
		CreatedAt:     createdAt.Time,
		UpdatedAt:     updatedAt.Time,
	}
	if expiresAt.Valid {
		track.ExpiresAt = expiresAt.Time
	}
	return track, nil
}

// MintAudioTrack creates an uploaded audio track attached to the given asset,
// verifying ownership (D-S5L1-3). It evicts expired Created/Uploading tracks of
// the same asset before inserting, mirroring video Mint's expired-upload eviction.
func (repository *Repository) MintAudioTrack(
	ctx context.Context, actor domain.Actor, assetID string,
	declaredBytes int64, originalName, language, label string,
	expiresAt, now time.Time,
) (domain.AudioTrack, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return domain.AudioTrack{}, persistenceError(err)
	}
	defer rollback(tx)
	now = storeTime(now)
	expiresAt = storeTime(expiresAt)
	if _, err := authorizeAudioEdit(ctx, tx, actor, assetID); err != nil {
		return domain.AudioTrack{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE audio_tracks
		SET state='failed', updated_at=$2
		WHERE video_asset_id=$1::text::uuid
		  AND kind='uploaded' AND state IN ('created','uploading')
		  AND expires_at <= $2`, assetID, now); err != nil {
		return domain.AudioTrack{}, persistenceError(err)
	}
	stored, err := scanAudioTrack(tx.QueryRow(ctx, `
		WITH new_id AS (SELECT gen_random_uuid() AS id)
		INSERT INTO audio_tracks(
			id,video_asset_id,kind,language,label,is_default,state,source_leaf,
			declared_bytes,received_bytes,original_name,expires_at,created_at,updated_at)
		SELECT id,$1::text::uuid,'uploaded',$2,$3,false,'created',
			'audio-'||id::text,$4,0,$5,$6,$7,$7
		FROM new_id
		RETURNING `+audioTrackColumns,
		assetID, language, label, declaredBytes, originalName, expiresAt, now))
	if err != nil {
		if mapped, ok := audioTrackConflict(err); ok {
			return domain.AudioTrack{}, mapped
		}
		return domain.AudioTrack{}, persistenceError(err)
	}
	if _, err := domain.NewUploadedTrack(stored); err != nil {
		return domain.AudioTrack{}, err
	}
	if err := commit(ctx, tx); err != nil {
		return domain.AudioTrack{}, err
	}
	return stored, nil
}

const lockAudioTrack = `SELECT ` + audioTrackColumns + `
FROM audio_tracks WHERE id=$1::text::uuid FOR UPDATE`

// AppendAudioChunk mirrors the video AppendChunk for an uploaded audio track.
func (repository *Repository) AppendAudioChunk(
	ctx context.Context, trackID string, offset int64, now time.Time,
	chunkPolicy ports.AudioChunkPolicy, completePolicy ports.AudioCompletePolicy,
) (domain.AudioTrack, error) {
	tx, err := repository.pool.Begin(ctx)
	if err != nil {
		return domain.AudioTrack{}, persistenceError(err)
	}
	defer rollback(tx)
	now = storeTime(now)
	locked, err := scanAudioTrack(tx.QueryRow(ctx, lockAudioTrack, trackID))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.AudioTrack{}, domain.ErrVideoNotFound
	}
	if err != nil {
		return domain.AudioTrack{}, persistenceError(err)
	}
	if locked.Kind != domain.UploadedTrack ||
		(locked.State != domain.Created && locked.State != domain.Uploading) ||
		locked.ExpiresAt.IsZero() || !locked.ExpiresAt.After(now) {
		return domain.AudioTrack{}, domain.ErrVideoNotFound
	}
	if locked.ReceivedBytes != offset {
		return domain.AudioTrack{}, domain.ErrInvalidStateTransition
	}
	newReceived, err := chunkPolicy(locked)
	if err != nil {
		return domain.AudioTrack{}, err
	}
	if newReceived < offset || newReceived > locked.DeclaredBytes {
		return domain.AudioTrack{}, domain.ErrInvalidRequest
	}
	if newReceived == locked.DeclaredBytes {
		completing := locked
		completing.State = domain.Uploading
		completing.ReceivedBytes = newReceived
		if err := completePolicy(completing); err != nil {
			return domain.AudioTrack{}, err
		}
	}
	updated, err := scanAudioTrack(tx.QueryRow(ctx, `UPDATE audio_tracks
		SET received_bytes=$3,
			state=CASE WHEN $3=declared_bytes THEN 'queued' ELSE 'uploading' END,
			updated_at=$4
		WHERE id=$1::text::uuid AND received_bytes=$2
		  AND kind='uploaded' AND state IN ('created','uploading')
		  AND expires_at>$4
		RETURNING `+audioTrackColumns, trackID, offset, newReceived, now))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.AudioTrack{}, domain.ErrInvalidStateTransition
	}
	if err != nil {
		return domain.AudioTrack{}, persistenceError(err)
	}
	if err := commit(ctx, tx); err != nil {
		return domain.AudioTrack{}, err
	}
	return updated, nil
}

// CancelAudioUpload fails a Created or Uploading uploaded track.
func (repository *Repository) CancelAudioUpload(ctx context.Context, trackID string, now time.Time) error {
	now = storeTime(now)
	result, err := repository.pool.Exec(ctx, `UPDATE audio_tracks
		SET state='failed', updated_at=$2
		WHERE id=$1::text::uuid AND kind='uploaded'
		  AND state IN ('created','uploading')`, trackID, now)
	if err != nil {
		return persistenceError(err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrVideoNotFound
	}
	return nil
}

// SweepAudioTrackCandidates returns uploaded tracks whose upload expired.
func (repository *Repository) SweepAudioTrackCandidates(ctx context.Context, now time.Time, limit int) ([]domain.AudioTrack, error) {
	if limit <= 0 {
		return nil, nil
	}
	now = storeTime(now)
	rows, err := repository.pool.Query(ctx, `SELECT `+audioTrackColumns+`
		FROM audio_tracks
		WHERE kind='uploaded' AND state IN ('created','uploading')
		  AND expires_at < $1
		ORDER BY expires_at LIMIT $2`, now, limit)
	if err != nil {
		return nil, persistenceError(err)
	}
	defer rows.Close()
	var tracks []domain.AudioTrack
	for rows.Next() {
		track, err := scanAudioTrack(rows)
		if err != nil {
			return nil, persistenceError(err)
		}
		tracks = append(tracks, track)
	}
	return tracks, persistenceError(rows.Err())
}

// DeleteAudioTrackSwept removes a swept audio track row.
func (repository *Repository) DeleteAudioTrackSwept(ctx context.Context, trackID string) error {
	result, err := repository.pool.Exec(ctx, `DELETE FROM audio_tracks WHERE id=$1::text::uuid`, trackID)
	if err != nil {
		return persistenceError(err)
	}
	if result.RowsAffected() == 0 {
		return domain.ErrVideoNotFound
	}
	return nil
}

func audioTrackConflict(err error) (error, bool) {
	var postgresError *pgconn.PgError
	if !errors.As(err, &postgresError) {
		return nil, false
	}
	switch postgresError.Code {
	case "23503": // foreign key: the asset does not exist
		return domain.ErrVideoNotFound, true
	case "23505": // a unique index
		if postgresError.ConstraintName == "audio_tracks_language_per_asset" {
			return domain.ErrDuplicateLanguage, true
		}
		return domain.ErrInvalidRequest, true
	case "23514": // a check constraint
		return domain.ErrInvalidRequest, true
	}
	return nil, false
}
