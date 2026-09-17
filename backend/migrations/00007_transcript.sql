-- Down deletes every transcript and the course document library; it is test-only for
-- disposable databases. Retained environments recover forward, because the text was typed by
-- a human and nothing regenerates it.
-- +goose Up
SET LOCAL lock_timeout = '5s';

-- The key carries the language from day one (§24.16 set the precedent for the neighbouring
-- audio-track model: the language of a track is a mandatory field from the first day, not one
-- added later). This phase fills exactly one language; the second needs no migration.
CREATE TABLE transcripts (
    lesson_id uuid NOT NULL REFERENCES course_lessons(id) ON DELETE CASCADE,
    language text NOT NULL CHECK (
        char_length(language) BETWEEN 2 AND 35
        AND language = lower(language)
        AND language ~ '^[a-z][a-z0-9]{1,7}(-[a-z0-9]{1,8})*$'
    ),
    -- Zero means the text was written while the lesson had no ready asset. Staleness is derived
    -- by comparing this with the generation the player would pick, so there is no stored flag
    -- that can drift away from reality.
    authored_against_generation integer NOT NULL DEFAULT 0 CHECK (authored_against_generation >= 0),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (lesson_id, language),
    CHECK (updated_at >= created_at)
);

CREATE TABLE transcript_segments (
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'::uuid),
    lesson_id uuid NOT NULL,
    language text NOT NULL,
    ordinal integer NOT NULL CHECK (ordinal >= 0),
    start_ms bigint NOT NULL CHECK (start_ms >= 0),
    end_ms bigint NOT NULL,
    text text NOT NULL CHECK (char_length(text) BETWEEN 1 AND 1000),
    speaker text CHECK (speaker IS NULL OR char_length(speaker) BETWEEN 1 AND 100),
    confidence double precision CHECK (confidence IS NULL OR confidence BETWEEN 0 AND 1),
    CHECK (end_ms >= start_ms),
    FOREIGN KEY (lesson_id, language) REFERENCES transcripts(lesson_id, language) ON DELETE CASCADE,
    UNIQUE (lesson_id, language, ordinal)
);

CREATE INDEX transcript_segments_ordered ON transcript_segments (lesson_id, language, ordinal);

CREATE TABLE course_documents (
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'::uuid),
    course_id uuid NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    label text NOT NULL CHECK (char_length(label) BETWEEN 1 AND 200 AND label ~ '[^[:space:]]'),
    -- The scheme is enforced here as well as in the domain: a CHECK that only the domain knows
    -- about is a rule one insert path can forget.
    url text NOT NULL CHECK (char_length(url) BETWEEN 1 AND 2000 AND url ~ '^https?://'),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX course_documents_by_course ON course_documents (course_id, id);

-- Deleting a course cascades through both sides. The rule that a document cannot be removed
-- while a segment still references it belongs to the author's action, not to this cascade —
-- otherwise saving a course tree would start failing on a flow that already works.
CREATE TABLE segment_documents (
    segment_id uuid NOT NULL REFERENCES transcript_segments(id) ON DELETE CASCADE,
    document_id uuid NOT NULL REFERENCES course_documents(id) ON DELETE CASCADE,
    position integer NOT NULL CHECK (position >= 0),
    PRIMARY KEY (segment_id, document_id),
    UNIQUE (segment_id, position)
);

CREATE INDEX segment_documents_by_document ON segment_documents (document_id);

-- +goose Down
DROP TABLE segment_documents;
DROP TABLE course_documents;
DROP TABLE transcript_segments;
DROP TABLE transcripts;
