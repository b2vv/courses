-- Multi-subtitles MS1 (docs/multi-subtitles/plan-ms1.md M-D2/M-D3, phase D2/D8).
--
-- A track's language becomes the key of its subtitles: the player shows the transcript whose
-- language is lower(track.language). So a track language must be a tag a transcript can carry
-- (the 00007 grammar, applied to the lowercased value; case is kept), and two non-failed tracks
-- of one asset may not differ only by case — otherwise one transcript would belong to two tracks.
--
-- Legacy (owner decision O6, 2026-11-14): an unknown-language (`und` family) track of a Ready
-- asset whose lesson already has a primary transcript takes that transcript's language, so the
-- subtitles of existing lessons do not disappear. The manifest on disk is NOT rewritten here — a
-- migration cannot write files; `media-worker regenerate-masters` (the `media-masters` compose
-- service) does it before the worker starts.
--
-- RAISE EXCEPTION texts below carry only ids and counts: cmd/migrate logs P0001 messages
-- verbatim, so no user-authored value may ever be interpolated into one.
-- +goose Up
SET LOCAL lock_timeout = '5s';
LOCK TABLE video_assets, audio_tracks, transcripts IN SHARE ROW EXCLUSIVE MODE;

-- Step 1 — preflight on the ORIGINAL rows, before anything is relabelled, so an invalid or
-- colliding row is refused rather than silently "healed" by the relabel (`und-` would otherwise
-- become a valid language, and one of `und`/`UND` would be renamed away).
-- +goose StatementBegin
DO $$
DECLARE
    bad_count bigint;
    bad_id uuid;
    collision uuid;
BEGIN
    SELECT count(*) INTO bad_count FROM audio_tracks
    WHERE NOT (language ~ '^[A-Za-z0-9-]+$'
               AND char_length(language) BETWEEN 2 AND 35
               AND lower(language) ~ '^[a-z][a-z0-9]{1,7}(-[a-z0-9]{1,8})*$');
    IF bad_count > 0 THEN
        SELECT id INTO bad_id FROM audio_tracks
        WHERE NOT (language ~ '^[A-Za-z0-9-]+$'
                   AND char_length(language) BETWEEN 2 AND 35
                   AND lower(language) ~ '^[a-z][a-z0-9]{1,7}(-[a-z0-9]{1,8})*$')
        ORDER BY id LIMIT 1;
        RAISE EXCEPTION 'audio track language shape: % track(s) carry a language no transcript can use; first track %', bad_count, bad_id;
    END IF;

    SELECT video_asset_id INTO collision FROM audio_tracks
    WHERE state <> 'failed'
    GROUP BY video_asset_id, lower(language)
    HAVING count(*) > 1
    ORDER BY video_asset_id LIMIT 1;
    IF collision IS NOT NULL THEN
        RAISE EXCEPTION 'audio track language shape: asset % has non-failed tracks whose languages differ only by case', collision;
    END IF;
END
$$;
-- +goose StatementEnd

-- Step 2 — legacy relabel (O6). One unknown-language track per Ready, non-evicted asset whose
-- lesson has a primary transcript P and no non-failed track in P already. The elected track:
-- the default first, then the embedded one, then the smallest id. Others stay as they are.
UPDATE audio_tracks AS tracks
SET language = elected.primary_language, updated_at = now()
FROM (
    SELECT DISTINCT ON (candidate.video_asset_id)
           candidate.id, primary_transcript.language AS primary_language
    FROM audio_tracks AS candidate
    JOIN video_assets AS assets ON assets.id = candidate.video_asset_id
    JOIN transcripts AS primary_transcript
      ON primary_transcript.lesson_id = assets.lesson_id AND primary_transcript.is_primary
    WHERE assets.state = 'ready' AND assets.evicted_at IS NULL
      AND candidate.state <> 'failed'
      AND (lower(candidate.language) = 'und' OR lower(candidate.language) LIKE 'und-%')
      AND NOT EXISTS (SELECT 1 FROM audio_tracks AS taken
                      WHERE taken.video_asset_id = candidate.video_asset_id
                        AND taken.state <> 'failed'
                        AND lower(taken.language) = primary_transcript.language)
    ORDER BY candidate.video_asset_id, candidate.is_default DESC,
             (candidate.kind = 'embedded') DESC, candidate.id
) AS elected
WHERE tracks.id = elected.id;

-- Steps 3–4 — the rule, held by the database for every write path from now on.
ALTER TABLE audio_tracks ADD CONSTRAINT audio_tracks_language_shape CHECK (
    language ~ '^[A-Za-z0-9-]+$'
    AND char_length(language) BETWEEN 2 AND 35
    AND lower(language) ~ '^[a-z][a-z0-9]{1,7}(-[a-z0-9]{1,8})*$'
);

-- Same name as 00026's index: the repository maps a violation to ErrDuplicateLanguage by it.
DROP INDEX audio_tracks_language_per_asset;
CREATE UNIQUE INDEX audio_tracks_language_per_asset ON audio_tracks (video_asset_id, lower(language))
WHERE state <> 'failed';

-- +goose Down
SET LOCAL lock_timeout = '5s';

-- The languages step 2 set are NOT restored: down cannot know which tracks were `und`, and
-- restoring them would take subtitles away from lessons that show them.
DROP INDEX audio_tracks_language_per_asset;
CREATE UNIQUE INDEX audio_tracks_language_per_asset ON audio_tracks (video_asset_id, language)
WHERE state <> 'failed';
ALTER TABLE audio_tracks DROP CONSTRAINT audio_tracks_language_shape;
