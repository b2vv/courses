-- +goose Up
SET LOCAL lock_timeout = '5s';

ALTER TABLE discussion_comments
    ADD COLUMN removed_by uuid NULL REFERENCES users(id);

ALTER TABLE discussion_comments
    ADD CONSTRAINT discussion_comments_removed_by_check
        CHECK (removed_by IS NULL OR removed_at IS NOT NULL);

CREATE TABLE discussion_comment_reports (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    comment_id uuid NOT NULL REFERENCES discussion_comments(id) ON DELETE CASCADE,
    reporter_id uuid NOT NULL REFERENCES users(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    dismissed_at timestamptz,
    dismissed_by uuid REFERENCES users(id),
    UNIQUE (comment_id, reporter_id)
);

CREATE TABLE discussion_moderation_log (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id uuid NOT NULL,
    comment_id uuid NOT NULL,
    actor_id uuid NOT NULL,
    actor_role text NOT NULL,
    action text NOT NULL,
    reason text,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX discussion_moderation_log_course_keyset
    ON discussion_moderation_log (course_id, created_at DESC, id DESC);

CREATE INDEX discussion_comment_reports_open
    ON discussion_comment_reports (comment_id)
    WHERE dismissed_at IS NULL;

-- The course is reached through course_lessons -> course_modules -> courses;
-- the query's global ordering is therefore indexed by its complete keyset and
-- applies the course join/filter while walking the ordered comment stream.
CREATE INDEX discussion_comments_course_keyset
    ON discussion_comments (created_at DESC, id DESC);

-- +goose Down
DROP TABLE discussion_moderation_log;
DROP TABLE discussion_comment_reports;
ALTER TABLE discussion_comments DROP CONSTRAINT discussion_comments_removed_by_check;
ALTER TABLE discussion_comments DROP COLUMN removed_by;
