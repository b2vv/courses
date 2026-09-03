-- name: ListPublishedCourses :many
SELECT id, slug, title, summary, description, status, created_at, updated_at, published_at
FROM courses
WHERE status = 'published'
  AND (title ILIKE '%' || sqlc.arg(search_pattern)::text || '%' ESCAPE '!'
       OR summary ILIKE '%' || sqlc.arg(search_pattern)::text || '%' ESCAPE '!')
ORDER BY
  CASE WHEN sqlc.arg(sort_by)::text = 'title' THEN title COLLATE "C" END ASC,
  CASE WHEN sqlc.arg(sort_by)::text = 'recent' THEN published_at END DESC,
  id ASC
LIMIT sqlc.arg(page_limit)::integer OFFSET sqlc.arg(page_offset)::integer;

-- name: CountPublishedCourses :one
SELECT count(*) FROM courses
WHERE status = 'published'
  AND (title ILIKE '%' || sqlc.arg(search_pattern)::text || '%' ESCAPE '!'
       OR summary ILIKE '%' || sqlc.arg(search_pattern)::text || '%' ESCAPE '!');

-- name: GetPublishedCourse :one
SELECT id, slug, title, summary, description, status, created_at, updated_at, published_at
FROM courses WHERE slug = $1 AND status = 'published';
