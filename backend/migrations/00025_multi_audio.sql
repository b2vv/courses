-- Down drops the audio-track table and restores the pre-multi-audio video_assets
-- state set; it is test-only for disposable databases. No media bytes are touched.
-- +goose Up
ALTER TABLE video_assets
    DROP CONSTRAINT video_assets_state_check,
    ADD CONSTRAINT video_assets_state_check
        CHECK (state IN ('created','uploading','queued','processing','awaiting_audio','ready','failed'));

CREATE TABLE audio_tracks (
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'::uuid),
    video_asset_id uuid NOT NULL REFERENCES video_assets(id) ON DELETE CASCADE,
    kind text NOT NULL CHECK (kind IN ('embedded','uploaded')),
    language text NOT NULL CHECK (char_length(language) BETWEEN 1 AND 35),
    label text NOT NULL CHECK (char_length(label) BETWEEN 1 AND 64),
    is_default boolean NOT NULL DEFAULT false,
    state text NOT NULL CHECK (state IN ('created','uploading','queued','processing','ready','failed')),
    source_leaf text NOT NULL,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT audio_tracks_updated_not_before_created CHECK (updated_at >= created_at),
    -- An embedded track materialises straight to Ready; it never owns upload states (D3).
    CONSTRAINT audio_tracks_embedded_is_ready CHECK (kind <> 'embedded' OR state = 'ready'),
    -- The source leaf is fixed by kind: embedded reuses the video "original" leaf,
    -- uploaded stores its bytes at audio-<trackId> (D3).
    CONSTRAINT audio_tracks_leaf_matches_kind CHECK (
        (kind = 'embedded' AND source_leaf = 'original')
        OR (kind = 'uploaded' AND source_leaf = 'audio-' || id::text)
    ),
    -- The default is always a Ready track (D4). Combined with the partial unique
    -- index below this gives at most one default per asset, and it is Ready;
    -- "exactly one once the asset is Ready" is enforced transactionally.
    CONSTRAINT audio_tracks_default_is_ready CHECK (is_default = false OR state = 'ready')
);

CREATE UNIQUE INDEX audio_tracks_one_default_per_asset ON audio_tracks (video_asset_id)
WHERE is_default;

CREATE INDEX audio_tracks_by_asset ON audio_tracks (video_asset_id, created_at, id);

-- +goose Down
DROP TABLE audio_tracks;
ALTER TABLE video_assets
    DROP CONSTRAINT video_assets_state_check,
    ADD CONSTRAINT video_assets_state_check
        CHECK (state IN ('created','uploading','queued','processing','ready','failed'));
