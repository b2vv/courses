-- MS1 retained-volume diagnostic (read-only): every non-failed und-family track of a Ready,
-- non-evicted asset, with the lesson's primary transcript language and why 00030 step 2 would /
-- would not relabel it. Mirrors the migration's candidate predicate AND its election (default
-- first, then embedded, then smallest id): only one candidate per asset is 'elected'; the rest are
-- 'candidate, not elected' and keep und.
WITH rows AS (
  SELECT t.id AS track, a.id AS asset, t.language, t.is_default, t.kind,
         p.language AS primary_language,
         CASE
           WHEN p.language IS NULL THEN 'no primary transcript'
           WHEN EXISTS (SELECT 1 FROM audio_tracks o WHERE o.video_asset_id = a.id AND o.state <> 'failed'
                        AND lower(o.language) = p.language) THEN 'asset already has a track in the primary language'
           ELSE 'candidate'
         END AS status
  FROM audio_tracks t
  JOIN video_assets a ON a.id = t.video_asset_id
  LEFT JOIN transcripts p ON p.lesson_id = a.lesson_id AND p.is_primary
  WHERE a.state = 'ready' AND a.evicted_at IS NULL AND t.state <> 'failed'
    AND (lower(t.language) = 'und' OR lower(t.language) LIKE 'und-%')
)
SELECT track, asset, language, is_default, kind, primary_language,
       CASE WHEN status <> 'candidate' THEN status
            WHEN row_number() OVER (PARTITION BY asset, status = 'candidate'
                                    ORDER BY is_default DESC, (kind = 'embedded') DESC, track) = 1
              THEN 'eligible (elected)'
            ELSE 'candidate, not elected'
       END AS reason
FROM rows
ORDER BY reason, asset, track;
