-- Down destroys comments written by people. Unlike generated media, they cannot be recreated;
-- this is a warning for disposable databases only, not a retained-environment migration plan.
-- +goose Up
SET LOCAL lock_timeout = '5s';

-- A comment remains after moderation so a thread can retain its shape. Removal clears `body`
-- and stamps `removed_at`; the application, rather than a DELETE, owns that transition.
CREATE TABLE discussion_comments (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    lesson_id uuid NOT NULL REFERENCES course_lessons(id) ON DELETE CASCADE,
    author_id uuid NOT NULL REFERENCES users(id),
    parent_id uuid REFERENCES discussion_comments(id),
    body text CHECK (body IS NULL OR (char_length(body) BETWEEN 1 AND 2000 AND body = btrim(body))),
    position_ms bigint CHECK (position_ms IS NULL OR position_ms >= 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    edited_at timestamptz,
    removed_at timestamptz,

    CONSTRAINT discussion_comments_body_removal_check
        CHECK ((removed_at IS NULL) = (body IS NOT NULL))
);

-- Root pagination walks newest roots by the complete keyset, while replies walk oldest first.
CREATE INDEX discussion_comments_roots_keyset
    ON discussion_comments (lesson_id, created_at DESC, id DESC)
    WHERE parent_id IS NULL;

CREATE INDEX discussion_comments_replies_keyset
    ON discussion_comments (parent_id, created_at, id);

-- +goose Down
DROP TABLE discussion_comments;
