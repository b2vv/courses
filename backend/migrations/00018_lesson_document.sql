-- +goose Up
SET LOCAL lock_timeout = '5s';

CREATE TABLE lesson_documents (
    lesson_id uuid NOT NULL,
    locale text NOT NULL,
    draft_body jsonb NOT NULL,
    published_body jsonb NOT NULL,
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (lesson_id, locale),
    FOREIGN KEY (lesson_id) REFERENCES course_lessons(id) ON DELETE CASCADE,
    CONSTRAINT lesson_documents_locale_ck CHECK (locale IN ('uk', 'en')),
    CONSTRAINT lesson_documents_draft_object_ck CHECK (jsonb_typeof(draft_body) = 'object'),
    CONSTRAINT lesson_documents_published_object_ck CHECK (jsonb_typeof(published_body) = 'object'),
    CONSTRAINT lesson_documents_draft_size_ck CHECK (pg_column_size(draft_body) <= 1048576),
    CONSTRAINT lesson_documents_published_size_ck CHECK (pg_column_size(published_body) <= 1048576)
);

CREATE TABLE lesson_images (
    id uuid PRIMARY KEY,
    lesson_id uuid NOT NULL REFERENCES course_lessons(id) ON DELETE CASCADE,
    content_type text NOT NULL,
    width integer NOT NULL,
    height integer NOT NULL,
    byte_size bigint NOT NULL,
    created_by uuid REFERENCES users(id),
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX lesson_images_lesson_id_idx ON lesson_images (lesson_id);

INSERT INTO lesson_documents (lesson_id, locale, draft_body, published_body, updated_at)
SELECT built.id, 'uk', built.doc, built.doc, now()
FROM (
    SELECT cl.id,
        jsonb_build_object(
            'schemaVersion', 1,
            'locale', 'uk',
            'sections', jsonb_build_array(
                jsonb_build_object(
                    'id', cl.id::text || '-s1',
                    'type', 'text',
                    'content', jsonb_build_object(
                        'blocks', jsonb_build_array(
                            jsonb_build_object(
                                'type', 'paragraph',
                                'spans', jsonb_build_array(jsonb_build_object('text', cl.description))
                            )
                        )
                    )
                )
            )
        ) AS doc
    FROM course_lessons AS cl
    WHERE cl.description <> ''
) AS built;

-- +goose Down
DROP TABLE IF EXISTS lesson_images;
DROP TABLE IF EXISTS lesson_documents;
