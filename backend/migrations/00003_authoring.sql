-- A developer volume with courses but no users cannot be backfilled and fails here.
-- Recover local seed-only data with: docker compose down -v && make up &&
-- make seed-identities && make seed.
-- +goose Up
ALTER TABLE courses
    ADD COLUMN instructor_id uuid REFERENCES users(id) ON DELETE RESTRICT;

UPDATE courses
SET instructor_id = (
    SELECT id
    FROM users
    WHERE role IN ('admin', 'instructor')
        AND disabled_at IS NULL
    ORDER BY created_at, id
    LIMIT 1
)
WHERE instructor_id IS NULL;

ALTER TABLE courses ALTER COLUMN instructor_id SET NOT NULL;

CREATE INDEX courses_by_instructor ON courses (instructor_id, updated_at DESC, id);

CREATE TABLE course_modules (
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'::uuid),
    course_id uuid NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    title text NOT NULL CHECK (char_length(title) BETWEEN 1 AND 200 AND title ~ '[^[:space:]]'),
    position integer NOT NULL CHECK (position >= 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK (updated_at >= created_at)
);

CREATE INDEX course_modules_ordered ON course_modules (course_id, position, id);

CREATE TABLE course_lessons (
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'::uuid),
    module_id uuid NOT NULL REFERENCES course_modules(id) ON DELETE CASCADE,
    title text NOT NULL CHECK (char_length(title) BETWEEN 1 AND 200 AND title ~ '[^[:space:]]'),
    description text NOT NULL CHECK (char_length(description) BETWEEN 0 AND 1000),
    position integer NOT NULL CHECK (position >= 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CHECK (updated_at >= created_at)
);

CREATE INDEX course_lessons_ordered ON course_lessons (module_id, position, id);

-- +goose Down
DROP TABLE course_lessons;
DROP TABLE course_modules;
DROP INDEX courses_by_instructor;
ALTER TABLE courses DROP COLUMN instructor_id;
