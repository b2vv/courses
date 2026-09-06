-- name: LoadMediaAsset :one
SELECT
  id::text AS id,
  COALESCE(lesson_id::text, '')::text AS lesson_id,
  generation,
  state,
  COALESCE(replace(claim_token::text, '-', ''), '')::text AS claim_token,
  declared_bytes,
  received_bytes,
  original_name,
  attempt,
  COALESCE(failure_reason, '')::text AS failure_reason,
  renditions,
  expires_at,
  evicted_at,
  created_at,
  updated_at
FROM video_assets
WHERE id = sqlc.arg(asset_id)::text::uuid
  AND evicted_at IS NULL;

-- name: ListMediaCourseVideos :many
SELECT
  courses.instructor_id::text AS instructor_id,
  (assets.id IS NOT NULL)::boolean AS asset_present,
  COALESCE(assets.id::text, '')::text AS id,
  COALESCE(assets.lesson_id::text, '')::text AS lesson_id,
  COALESCE(assets.generation, 0)::integer AS generation,
  COALESCE(assets.state, '')::text AS state,
  COALESCE(replace(assets.claim_token::text, '-', ''), '')::text AS claim_token,
  COALESCE(assets.declared_bytes, 0)::bigint AS declared_bytes,
  COALESCE(assets.received_bytes, 0)::bigint AS received_bytes,
  COALESCE(assets.original_name, '')::text AS original_name,
  COALESCE(assets.attempt, 0)::integer AS attempt,
  COALESCE(assets.failure_reason, '')::text AS failure_reason,
  COALESCE(assets.renditions, '{}'::text[])::text[] AS renditions,
  assets.expires_at,
  assets.evicted_at,
  assets.created_at,
  assets.updated_at
FROM courses
LEFT JOIN course_modules AS modules ON modules.course_id = courses.id
LEFT JOIN course_lessons AS lessons ON lessons.module_id = modules.id
LEFT JOIN video_assets AS assets ON assets.lesson_id = lessons.id AND assets.evicted_at IS NULL
WHERE courses.id = sqlc.arg(course_id)::text::uuid
ORDER BY assets.generation DESC NULLS LAST, assets.id;
