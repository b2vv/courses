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

-- name: GetPublishedCourse :many
SELECT
  courses.id,
  courses.slug,
  courses.title,
  courses.summary,
  courses.description,
  courses.status,
  courses.created_at,
  courses.updated_at,
  courses.published_at,
  COALESCE(modules.id::text, '')::text AS module_id,
  COALESCE(modules.title, '')::text AS module_title,
  COALESCE(lessons.title, '')::text AS lesson_title,
  (lessons.id IS NOT NULL)::boolean AS lesson_present
FROM courses
LEFT JOIN course_modules AS modules ON modules.course_id = courses.id
LEFT JOIN course_lessons AS lessons ON lessons.module_id = modules.id
WHERE courses.slug = $1 AND courses.status = 'published'
ORDER BY modules.position, modules.id, lessons.position, lessons.id;

-- name: ListAuthoringCourses :many
SELECT
  id::text AS id,
  slug,
  title,
  status,
  instructor_id::text AS instructor_id,
  updated_at
FROM courses
WHERE (sqlc.arg(is_admin)::boolean OR instructor_id = sqlc.arg(actor_id)::text::uuid)
  AND (sqlc.arg(status)::text = '' OR status = sqlc.arg(status)::text)
ORDER BY updated_at DESC, id ASC
LIMIT sqlc.arg(page_limit)::integer OFFSET sqlc.arg(page_offset)::integer;

-- name: CountAuthoringCourses :one
SELECT count(*)
FROM courses
WHERE (sqlc.arg(is_admin)::boolean OR instructor_id = sqlc.arg(actor_id)::text::uuid)
  AND (sqlc.arg(status)::text = '' OR status = sqlc.arg(status)::text);

-- name: GetAuthoringCourse :one
SELECT
  id::text AS id,
  slug,
  title,
  summary,
  description,
  status,
  instructor_id::text AS instructor_id,
  created_at,
  updated_at,
  published_at
FROM courses
WHERE id = sqlc.arg(course_id)::text::uuid;

-- name: LockAuthoringCourse :one
SELECT
  id::text AS id,
  slug,
  title,
  summary,
  description,
  status,
  instructor_id::text AS instructor_id,
  created_at,
  updated_at,
  published_at
FROM courses
WHERE id = sqlc.arg(course_id)::text::uuid
FOR UPDATE;

-- name: ListAuthoringModules :many
SELECT id::text AS id, title, position
FROM course_modules
WHERE course_id = sqlc.arg(course_id)::text::uuid
ORDER BY position, id;

-- name: ListAuthoringLessons :many
SELECT lessons.id::text AS id, lessons.module_id::text AS module_id,
  lessons.title, lessons.description, lessons.position
FROM course_lessons AS lessons
JOIN course_modules AS modules ON modules.id = lessons.module_id
WHERE modules.course_id = sqlc.arg(course_id)::text::uuid
ORDER BY modules.position, modules.id, lessons.position, lessons.id;

-- name: LockEligibleAuthoringOwner :one
SELECT id::text AS id
FROM users
WHERE id = sqlc.arg(owner_id)::text::uuid
  AND disabled_at IS NULL
  AND role IN ('instructor', 'admin')
FOR SHARE;

-- name: CreateAuthoringCourse :one
INSERT INTO courses (
  id, instructor_id, slug, title, summary, description, status,
  created_at, updated_at, published_at
)
VALUES (
  gen_random_uuid(),
  sqlc.arg(owner_id)::text::uuid,
  sqlc.arg(slug)::text,
  sqlc.arg(title)::text,
  sqlc.arg(summary)::text,
  sqlc.arg(description)::text,
  'draft',
  sqlc.arg(created_at)::timestamptz,
  sqlc.arg(created_at)::timestamptz,
  NULL
)
RETURNING id::text AS id;
