-- Multi-audio S-backfill (docs/multi-audio/plan-sbackfill.md). Before S3 no row in
-- audio_tracks existed for a Ready asset, yet every such asset carries the sound of its
-- source: pre-S3 AcceptSource refused a source without AAC/MP3 audio, and the transcoder
-- packaged it as audio/index.m3u8. Since S3 a Ready asset always holds at least one Ready
-- track (Publish inserts the embedded track, a video-only asset turns Ready only with its
-- first Ready upload, the last Ready track cannot be deleted). So a Ready, non-evicted
-- asset without a Ready track is a pre-S3 asset: it gets the embedded track Publish would
-- have created (D-B1), keyed by a deterministic id so Down removes only these rows (D-B2).
-- An old asset whose uploaded track already turned Ready is indistinguishable from a
-- video-only asset and is left alone (D-B6). Data only: no media bytes, no master.m3u8.
-- +goose Up
SET LOCAL lock_timeout = '5s';
-- Serializes with every writer of either table (mint, promote, delete, publish, evict):
-- SHARE ROW EXCLUSIVE conflicts with the ROW EXCLUSIVE their writes take (D-B1).
LOCK TABLE video_assets, audio_tracks IN SHARE ROW EXCLUSIVE MODE;

-- +goose StatementBegin
DO $$
DECLARE
    blocked uuid;
BEGIN
    SELECT assets.id INTO blocked
    FROM video_assets AS assets
    WHERE assets.state = 'ready' AND assets.evicted_at IS NULL
      AND NOT EXISTS (SELECT 1 FROM audio_tracks AS tracks
                      WHERE tracks.video_asset_id = assets.id AND tracks.state = 'ready')
      AND EXISTS (SELECT 1 FROM audio_tracks AS tracks
                  WHERE tracks.video_asset_id = assets.id AND tracks.state <> 'failed'
                    AND tracks.language = 'und')
      AND EXISTS (SELECT 1 FROM audio_tracks AS tracks
                  WHERE tracks.video_asset_id = assets.id AND tracks.state <> 'failed'
                    AND tracks.language = 'und-x-original')
    ORDER BY assets.id
    LIMIT 1;
    IF blocked IS NOT NULL THEN
        RAISE EXCEPTION 'multi-audio backfill: asset % already uses both und and und-x-original', blocked;
    END IF;
END
$$;
-- +goose StatementEnd

-- The language index (00026) also counts in-flight uploads, which cannot all be
-- cancelled, so a taken "und" falls back to the private-use tag (D-B1).
INSERT INTO audio_tracks (
    id, video_asset_id, kind, language, label, is_default, state, source_leaf, created_at, updated_at)
SELECT md5('multi-audio-backfill:' || assets.id::text)::uuid,
       assets.id,
       'embedded',
       CASE WHEN EXISTS (SELECT 1 FROM audio_tracks AS tracks
                         WHERE tracks.video_asset_id = assets.id AND tracks.state <> 'failed'
                           AND tracks.language = 'und')
            THEN 'und-x-original' ELSE 'und' END,
       'Original',
       true,
       'ready',
       'original',
       now(),
       now()
FROM video_assets AS assets
WHERE assets.state = 'ready' AND assets.evicted_at IS NULL
  AND NOT EXISTS (SELECT 1 FROM audio_tracks AS tracks
                  WHERE tracks.video_asset_id = assets.id AND tracks.state = 'ready')
ON CONFLICT (id) DO NOTHING;

-- The Ready gate for existing data (D-B3): every Ready, non-evicted asset now has at
-- least one Ready track and exactly one default (a default is always Ready, 00025).
-- +goose StatementBegin
DO $$
DECLARE
    broken uuid;
BEGIN
    SELECT assets.id INTO broken
    FROM video_assets AS assets
    WHERE assets.state = 'ready' AND assets.evicted_at IS NULL
      AND ((SELECT count(*) FROM audio_tracks AS tracks
            WHERE tracks.video_asset_id = assets.id AND tracks.state = 'ready') = 0
        OR (SELECT count(*) FROM audio_tracks AS tracks
            WHERE tracks.video_asset_id = assets.id AND tracks.is_default) <> 1)
    ORDER BY assets.id
    LIMIT 1;
    IF broken IS NOT NULL THEN
        RAISE EXCEPTION 'multi-audio backfill: ready asset % has no ready default audio track', broken;
    END IF;
END
$$;
-- +goose StatementEnd

-- +goose Down
-- Test-only, for disposable databases: removes exactly the rows Up created, leaving
-- embedded rows written by Publish and every uploaded track in place.
DELETE FROM audio_tracks
WHERE kind = 'embedded'
  AND id = md5('multi-audio-backfill:' || video_asset_id::text)::uuid;
