-- Multi-audio S6 (docs/multi-audio/plan-s6.md D-S6-1): a draft records the leaf of the audio it
-- was recognised from — the default track's `audio_tracks.source_leaf` at claim time, in the same
-- shape (embedded → 'original', uploaded → 'audio-<trackId>'). NULL is a draft made before S6,
-- which always read `source/<assetID>/original`, so readers treat it as 'original' and old drafts
-- are not re-recognised wholesale. Down drops the column; test-only, no bytes touched.
-- +goose Up
ALTER TABLE asr_drafts
    ADD COLUMN source_leaf text NULL,
    ADD CONSTRAINT asr_drafts_source_leaf_shape CHECK (
        source_leaf IS NULL
        OR source_leaf = 'original'
        OR source_leaf ~ '^audio-[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$'
    );

-- +goose Down
ALTER TABLE asr_drafts
    DROP CONSTRAINT IF EXISTS asr_drafts_source_leaf_shape,
    DROP COLUMN IF EXISTS source_leaf;
