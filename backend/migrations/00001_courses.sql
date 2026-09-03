-- +goose Up
CREATE TABLE courses (
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'::uuid),
    slug text NOT NULL UNIQUE CHECK (char_length(slug) <= 120 AND slug ~ '^[a-z0-9]+(-[a-z0-9]+)*$'),
    title text NOT NULL CHECK (char_length(title) BETWEEN 1 AND 200 AND title ~ '[^[:space:]]'),
    summary text NOT NULL CHECK (char_length(summary) BETWEEN 1 AND 500 AND summary ~ '[^[:space:]]'),
    description text NOT NULL CHECK (char_length(description) BETWEEN 1 AND 20000 AND description ~ '[^[:space:]]'),
    status text NOT NULL CHECK (status IN ('draft', 'published', 'archived')),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    published_at timestamptz,
    CHECK (updated_at >= created_at),
    CHECK (published_at IS NULL OR published_at BETWEEN created_at AND updated_at),
    CHECK (status <> 'published' OR published_at IS NOT NULL),
    CHECK (status <> 'draft' OR published_at IS NULL)
);

CREATE INDEX courses_published_recent ON courses (published_at DESC, id ASC) WHERE status = 'published';
CREATE INDEX courses_published_title ON courses (title COLLATE "C" ASC, id ASC) WHERE status = 'published';

-- +goose Down
DROP TABLE courses;
