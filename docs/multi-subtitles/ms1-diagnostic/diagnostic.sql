-- MS1 retained-volume diagnostic (read-only): every non-failed und-family track of a Ready,
-- non-evicted asset, with the lesson's primary transcript language and why it would / would not
-- take that language under 00030 step 2.
SELECT t.id AS track, a.id AS asset, t.language, t.is_default, t.kind,
       p.language AS primary_language,
       CASE
         WHEN p.language IS NULL THEN 'no primary transcript'
         WHEN EXISTS (SELECT 1 FROM audio_tracks o WHERE o.video_asset_id = a.id AND o.state <> 'failed'
                      AND lower(o.language) = p.language) THEN 'asset already has a track in the primary language'
         ELSE 'eligible'
       END AS reason
FROM audio_tracks t
JOIN video_assets a ON a.id = t.video_asset_id
LEFT JOIN transcripts p ON p.lesson_id = a.lesson_id AND p.is_primary
WHERE a.state = 'ready' AND a.evicted_at IS NULL AND t.state <> 'failed'
  AND (lower(t.language) = 'und' OR lower(t.language) LIKE 'und-%')
ORDER BY reason, asset, t.id;
