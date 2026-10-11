-- +goose Up
SET LOCAL lock_timeout = '5s';

-- Full-text search for the public catalog, replacing the ILIKE substring match.
-- 'simple' config: lowercasing + whitespace/punctuation tokenization, no stemming
-- dictionary — Postgres ships no Ukrainian dictionary, and 'simple' tokenizes
-- Cyrillic correctly (measured: to_tsvector('simple','фундамент') -> 'фундамент':1).
ALTER TABLE courses ADD COLUMN search_vector tsvector
    GENERATED ALWAYS AS (to_tsvector('simple', coalesce(title, '') || ' ' || coalesce(summary, ''))) STORED;
CREATE INDEX courses_search_vector_idx ON courses USING GIN (search_vector);

-- +goose Down
DROP INDEX courses_search_vector_idx;
ALTER TABLE courses DROP COLUMN search_vector;
