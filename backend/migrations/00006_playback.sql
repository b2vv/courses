-- Down deletes progress and persisted media durations and is test-only for disposable databases.
-- Retained environments recover forward; lost durations require re-uploading each video.
-- +goose Up
SET LOCAL lock_timeout = '5s';

CREATE TABLE lesson_progress (
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    lesson_id uuid NOT NULL REFERENCES course_lessons(id) ON DELETE CASCADE,
    position_seconds numeric NOT NULL,
    completed_at timestamptz,
    client_seq bigint NOT NULL,
    updated_at timestamptz NOT NULL,
    PRIMARY KEY (user_id, lesson_id)
);

ALTER TABLE video_assets ADD COLUMN duration_seconds numeric NULL;

-- +goose Down
DROP TABLE lesson_progress;
ALTER TABLE video_assets DROP COLUMN duration_seconds;
