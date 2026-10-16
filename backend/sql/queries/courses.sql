-- name: ListPublishedCourses :many
SELECT courses.id, courses.slug, courses.title, courses.summary, courses.description, courses.status, courses.created_at, courses.updated_at, courses.published_at,
  courses.price_minor, courses.currency,
  COALESCE(reviews.review_count, 0)::bigint AS review_count,
  COALESCE(reviews.review_stars_sum, 0)::bigint AS review_stars_sum,
  COALESCE(lessons.lesson_count, 0)::bigint AS lesson_count
FROM courses
LEFT JOIN LATERAL (
  SELECT count(*) AS review_count, COALESCE(sum(stars), 0)::bigint AS review_stars_sum
  FROM discussion_reviews
  WHERE course_id = courses.id AND removed_at IS NULL
) AS reviews ON true
LEFT JOIN LATERAL (
  SELECT count(*) AS lesson_count
  FROM course_lessons
  JOIN course_modules ON course_modules.id = course_lessons.module_id
  WHERE course_modules.course_id = courses.id
) AS lessons ON true
WHERE status = 'published'
  AND (sqlc.arg(search_pattern)::text = ''
       OR search_vector @@ websearch_to_tsquery('simple', sqlc.arg(search_pattern)::text))
ORDER BY
  CASE WHEN sqlc.arg(sort_by)::text = 'title' THEN title COLLATE "C" END ASC,
  CASE WHEN sqlc.arg(sort_by)::text = 'recent' THEN published_at END DESC,
  id ASC
LIMIT sqlc.arg(page_limit)::integer OFFSET sqlc.arg(page_offset)::integer;

-- name: CountPublishedCourses :one
SELECT count(*) FROM courses
WHERE status = 'published'
  AND (sqlc.arg(search_pattern)::text = ''
       OR search_vector @@ websearch_to_tsquery('simple', sqlc.arg(search_pattern)::text));

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
  courses.price_minor,
  courses.currency,
  COALESCE(reviews.review_count, 0)::bigint AS review_count,
  COALESCE(reviews.review_stars_sum, 0)::bigint AS review_stars_sum,
  COALESCE(lesson_counts.lesson_count, 0)::bigint AS lesson_count,
  COALESCE(modules.id::text, '')::text AS module_id,
  COALESCE(modules.title, '')::text AS module_title,
  COALESCE(lessons.title, '')::text AS lesson_title,
  (lessons.id IS NOT NULL)::boolean AS lesson_present
FROM courses
LEFT JOIN LATERAL (
  SELECT count(*) AS review_count, COALESCE(sum(stars), 0)::bigint AS review_stars_sum
  FROM discussion_reviews
  WHERE course_id = courses.id AND removed_at IS NULL
) AS reviews ON true
LEFT JOIN LATERAL (
  SELECT count(*) AS lesson_count
  FROM course_lessons AS counted_lessons
  JOIN course_modules AS counted_modules ON counted_modules.id = counted_lessons.module_id
  WHERE counted_modules.course_id = courses.id
) AS lesson_counts ON true
LEFT JOIN course_modules AS modules ON modules.course_id = courses.id
LEFT JOIN course_lessons AS lessons ON lessons.module_id = modules.id
WHERE courses.slug = $1 AND courses.status = 'published'
ORDER BY modules.position, modules.id, lessons.position, lessons.id;

-- name: ListAuthoringCourses :many
SELECT
  courses.id::text AS id,
  courses.slug,
  courses.title,
  courses.status,
  courses.instructor_id::text AS instructor_id,
  courses.updated_at,
  courses.price_minor,
  courses.currency,
  COALESCE(lessons.lesson_count, 0)::bigint AS lesson_count,
  COALESCE(enrollments.student_count, 0)::bigint AS student_count
FROM courses
LEFT JOIN LATERAL (
  SELECT count(*) AS lesson_count
  FROM course_lessons
  JOIN course_modules ON course_modules.id = course_lessons.module_id
  WHERE course_modules.course_id = courses.id
) AS lessons ON true
LEFT JOIN LATERAL (
  SELECT count(*) AS student_count
  FROM course_grants
  WHERE course_grants.course_id = courses.id AND course_grants.state = 'active'
) AS enrollments ON true
WHERE (sqlc.arg(is_admin)::boolean OR courses.instructor_id = sqlc.arg(actor_id)::text::uuid)
  AND (sqlc.arg(status)::text = '' OR courses.status = sqlc.arg(status)::text)
  AND (sqlc.arg(search_pattern)::text = ''
       OR courses.search_vector @@ websearch_to_tsquery('simple', sqlc.arg(search_pattern)::text))
ORDER BY courses.updated_at DESC, courses.id ASC
LIMIT sqlc.arg(page_limit)::integer OFFSET sqlc.arg(page_offset)::integer;

-- name: CountAuthoringCourses :one
SELECT count(*)
FROM courses
WHERE (sqlc.arg(is_admin)::boolean OR instructor_id = sqlc.arg(actor_id)::text::uuid)
  AND (sqlc.arg(status)::text = '' OR status = sqlc.arg(status)::text)
  AND (sqlc.arg(search_pattern)::text = ''
       OR search_vector @@ websearch_to_tsquery('simple', sqlc.arg(search_pattern)::text));

-- name: CountAuthoringCoursesByStatus :many
SELECT status, count(*)::bigint AS count
FROM courses
WHERE (sqlc.arg(is_admin)::boolean OR instructor_id = sqlc.arg(actor_id)::text::uuid)
GROUP BY status;

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
  , price_minor
  , currency
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
  , price_minor
  , currency
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
  lessons.title, lessons.position, lessons.is_preview
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
  created_at, updated_at, published_at, price_minor, currency
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
  NULL,
  sqlc.narg(price_minor)::bigint,
  sqlc.narg(currency)::text
)
RETURNING id::text AS id;
