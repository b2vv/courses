-- Multi-subtitles MS-D2/MS-D10: recognition belongs to an audio track, not to the lesson's
-- current default. Track identity lets two tracks keep independent attempts and tokens; the
-- language and multilingual snapshots explain staleness after metadata changes without
-- automatically spending another recognition run (owner O7).
-- +goose Up
SET LOCAL lock_timeout = '5s';

ALTER TABLE audio_tracks ADD COLUMN multilingual boolean NOT NULL DEFAULT false;
-- The backfill below and every per-track read resolve a draft's audio by (asset, leaf). Writers
-- already keep it unique (one embedded 'original', uploads at 'audio-<id>'); the index makes an
-- ambiguous match impossible rather than an arbitrary pick (MS2a review w1). An environment that
-- somehow holds two tracks on one leaf fails here, loudly, before any draft is rewritten.
CREATE UNIQUE INDEX audio_tracks_leaf_per_asset ON audio_tracks (video_asset_id, source_leaf);
ALTER TABLE asr_drafts
    ADD COLUMN audio_track_id uuid REFERENCES audio_tracks(id) ON DELETE CASCADE,
    ADD COLUMN track_language text CHECK (
        char_length(track_language) BETWEEN 2 AND 35
        AND track_language = lower(track_language)
        AND track_language ~ '^[a-z][a-z0-9]{1,7}(-[a-z0-9]{1,8})*$'
    ),
    ADD COLUMN track_multilingual boolean NOT NULL DEFAULT false;

-- The draft's own asset and heard leaf identify its source, even if the default has since
-- changed. Pre-S6 NULL leaves heard the embedded original. Unmatched machine output is
-- reproducible and is dropped rather than falsely attributed to another track (pre-launch).
UPDATE asr_drafts AS drafts
SET audio_track_id = tracks.id, track_language = lower(tracks.language)
FROM audio_tracks AS tracks
WHERE tracks.video_asset_id = drafts.asset_id
  AND tracks.source_leaf = COALESCE(drafts.source_leaf, 'original')
  AND tracks.state <> 'failed';
DELETE FROM asr_drafts WHERE audio_track_id IS NULL;

ALTER TABLE asr_drafts
    ALTER COLUMN audio_track_id SET NOT NULL,
    ALTER COLUMN track_language SET NOT NULL,
    DROP CONSTRAINT asr_drafts_pkey,
    ADD PRIMARY KEY (audio_track_id),
    DROP CONSTRAINT asr_drafts_source_leaf_shape,
    DROP COLUMN source_leaf;
-- Lesson and asset cascades remain: track deletion adds a third reason a draft has no owner.
-- Recovery still uses asr_drafts_running; lesson reads now need their own lookup index.
CREATE INDEX asr_drafts_lesson ON asr_drafts (lesson_id);

-- +goose Down
SET LOCAL lock_timeout = '5s';

-- Test-only rollback cannot represent several drafts of one lesson. Prefer its default
-- track's draft, otherwise the newest attempt; the track id makes tied timestamps stable.
DELETE FROM asr_drafts AS drafts
USING (
    SELECT drafts.audio_track_id,
           row_number() OVER (
               PARTITION BY drafts.lesson_id
               ORDER BY tracks.is_default DESC, drafts.updated_at DESC, drafts.audio_track_id
           ) AS position
    FROM asr_drafts AS drafts
    JOIN audio_tracks AS tracks ON tracks.id = drafts.audio_track_id
) AS ranked
WHERE drafts.audio_track_id = ranked.audio_track_id AND ranked.position > 1;

ALTER TABLE asr_drafts ADD COLUMN source_leaf text;
UPDATE asr_drafts AS drafts SET source_leaf = tracks.source_leaf
FROM audio_tracks AS tracks WHERE tracks.id = drafts.audio_track_id;
ALTER TABLE asr_drafts ADD CONSTRAINT asr_drafts_source_leaf_shape CHECK (
    source_leaf IS NULL OR source_leaf = 'original'
    OR source_leaf ~ '^audio-[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
);
DROP INDEX asr_drafts_lesson;
ALTER TABLE asr_drafts
    DROP CONSTRAINT asr_drafts_pkey,
    ADD PRIMARY KEY (lesson_id),
    DROP COLUMN audio_track_id,
    DROP COLUMN track_language,
    DROP COLUMN track_multilingual;
DROP INDEX audio_tracks_leaf_per_asset;
ALTER TABLE audio_tracks DROP COLUMN multilingual;
