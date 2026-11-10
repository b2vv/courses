-- +goose Up
ALTER TABLE audio_tracks
    ADD COLUMN declared_bytes bigint NOT NULL DEFAULT 0,
    ADD COLUMN received_bytes bigint NOT NULL DEFAULT 0,
    ADD COLUMN original_name text NOT NULL DEFAULT '',
    ADD COLUMN expires_at     timestamptz;

-- Uploaded tracks in flight must carry byte metadata; embedded tracks keep defaults.
ALTER TABLE audio_tracks
    ADD CONSTRAINT audio_tracks_uploaded_has_bytes
        CHECK (kind <> 'uploaded' OR state IN ('ready', 'failed') OR declared_bytes > 0);

-- +goose Down
ALTER TABLE audio_tracks DROP CONSTRAINT IF EXISTS audio_tracks_uploaded_has_bytes;
ALTER TABLE audio_tracks
    DROP COLUMN IF EXISTS expires_at,
    DROP COLUMN IF EXISTS original_name,
    DROP COLUMN IF EXISTS received_bytes,
    DROP COLUMN IF EXISTS declared_bytes;
