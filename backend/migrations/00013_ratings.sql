-- +goose Up
SET LOCAL lock_timeout = '5s';

CREATE TABLE discussion_reviews (
    id uuid PRIMARY KEY,
    course_id uuid NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    author_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    stars smallint NOT NULL CHECK (stars BETWEEN 1 AND 5),
    body text NOT NULL CHECK (char_length(btrim(body)) BETWEEN 20 AND 2000),
    completed_lessons integer NOT NULL CHECK (completed_lessons >= 0),
    total_lessons integer NOT NULL CHECK (total_lessons >= completed_lessons),
    created_at timestamptz NOT NULL,
    edited_at timestamptz NULL,
    removed_at timestamptz NULL,
    removed_by uuid NULL REFERENCES users(id),
    CHECK (removed_by IS NULL OR removed_at IS NOT NULL),
    UNIQUE (course_id, author_id)
);
CREATE INDEX discussion_reviews_course_newest ON discussion_reviews (course_id, created_at DESC, id DESC);
CREATE INDEX discussion_reviews_counted ON discussion_reviews (course_id) INCLUDE (stars) WHERE removed_at IS NULL;

CREATE TABLE discussion_review_reports (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    review_id uuid NOT NULL REFERENCES discussion_reviews(id) ON DELETE CASCADE,
    reporter_id uuid NOT NULL REFERENCES users(id),
    created_at timestamptz NOT NULL DEFAULT now(),
    dismissed_at timestamptz,
    dismissed_by uuid REFERENCES users(id),
    UNIQUE (review_id, reporter_id)
);
CREATE INDEX discussion_review_reports_open ON discussion_review_reports (review_id) WHERE dismissed_at IS NULL;

ALTER TABLE discussion_moderation_log ALTER COLUMN comment_id DROP NOT NULL;
ALTER TABLE discussion_moderation_log ADD COLUMN review_id uuid NULL;
ALTER TABLE discussion_moderation_log ADD CONSTRAINT discussion_moderation_log_one_target
    CHECK (num_nonnulls(comment_id, review_id) = 1);

-- +goose Down
DELETE FROM discussion_moderation_log WHERE review_id IS NOT NULL;
ALTER TABLE discussion_moderation_log DROP CONSTRAINT discussion_moderation_log_one_target;
ALTER TABLE discussion_moderation_log DROP COLUMN review_id;
ALTER TABLE discussion_moderation_log ALTER COLUMN comment_id SET NOT NULL;
DROP TABLE discussion_review_reports;
DROP TABLE discussion_reviews;
