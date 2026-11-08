-- One audio track per (asset, language), so a lesson's language list has no
-- duplicates (D6). Failed tracks are excluded, so a failed upload frees its
-- language for a retry. The embedded track's "und" counts, so an uploaded track
-- may not also claim "und". Down drops the index; test-only, no bytes touched.
-- +goose Up
CREATE UNIQUE INDEX audio_tracks_language_per_asset ON audio_tracks (video_asset_id, language)
WHERE state <> 'failed';

-- +goose Down
DROP INDEX audio_tracks_language_per_asset;
