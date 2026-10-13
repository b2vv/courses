-- +goose Up
SET LOCAL lock_timeout = '5s';

ALTER TABLE course_lessons DROP COLUMN description;

-- +goose Down
SET LOCAL lock_timeout = '5s';

-- Named loss: downgrade re-adds the column empty; original per-lesson text is not
-- reconstructed from lesson_documents.
ALTER TABLE course_lessons ADD COLUMN description text NOT NULL DEFAULT '' CHECK (char_length(description) BETWEEN 0 AND 1000);
