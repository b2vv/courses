-- +goose Up
SET LOCAL lock_timeout = '5s';

ALTER TABLE course_grants
    DROP CONSTRAINT course_grants_source_check,
    ADD CONSTRAINT course_grants_source_check
        CHECK (source IN ('purchase', 'admin', 'grandfather', 'self')),
    DROP CONSTRAINT course_grants_state_check,
    ADD CONSTRAINT course_grants_state_check
        CHECK (state IN ('active', 'blocked', 'revoked'));

DROP INDEX course_grants_one_active_per_user_course;
ALTER TABLE course_grants
    ADD CONSTRAINT course_grants_one_per_user_course UNIQUE (user_id, course_id);

CREATE TABLE participant_reports (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id uuid NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    reporter_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    target_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    kind text NOT NULL CHECK (kind IN ('on_lecturer', 'on_student')),
    reason text NOT NULL,
    state text NOT NULL DEFAULT 'open' CHECK (state IN ('open', 'resolved')),
    created_at timestamptz NOT NULL DEFAULT now(),
    CHECK (reporter_id <> target_id)
);

CREATE UNIQUE INDEX participant_reports_one_open_per_pair
    ON participant_reports (course_id, reporter_id, target_id)
    WHERE state = 'open';

INSERT INTO course_grants (id, user_id, course_id, source, order_id, state, granted_at, granted_by)
SELECT gen_random_uuid(), activity.user_id, activity.course_id, 'self', NULL, 'active', now(), NULL
FROM (
    SELECT progress.user_id, modules.course_id
    FROM lesson_progress AS progress
    JOIN course_lessons AS lessons ON lessons.id = progress.lesson_id
    JOIN course_modules AS modules ON modules.id = lessons.module_id

    UNION

    SELECT comments.author_id, modules.course_id
    FROM discussion_comments AS comments
    JOIN course_lessons AS lessons ON lessons.id = comments.lesson_id
    JOIN course_modules AS modules ON modules.id = lessons.module_id

    UNION

    SELECT reviews.author_id, reviews.course_id
    FROM discussion_reviews AS reviews
) AS activity
JOIN courses ON courses.id = activity.course_id
WHERE courses.price_minor IS NULL
  AND courses.status IN ('published', 'archived')
ON CONFLICT (user_id, course_id) DO NOTHING;

-- +goose Down
DROP TABLE participant_reports;

DELETE FROM course_grants WHERE source = 'self';
UPDATE course_grants SET state = 'revoked' WHERE state = 'blocked';

ALTER TABLE course_grants
    DROP CONSTRAINT course_grants_one_per_user_course;
CREATE UNIQUE INDEX course_grants_one_active_per_user_course
    ON course_grants (user_id, course_id)
    WHERE state = 'active';

ALTER TABLE course_grants
    DROP CONSTRAINT course_grants_source_check,
    ADD CONSTRAINT course_grants_source_check
        CHECK (source IN ('purchase', 'admin', 'grandfather')),
    DROP CONSTRAINT course_grants_state_check,
    ADD CONSTRAINT course_grants_state_check
        CHECK (state IN ('active', 'revoked'));
