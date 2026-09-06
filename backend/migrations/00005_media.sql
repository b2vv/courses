-- Down drops every media link and is test-only for disposable databases.
-- Retained environments recover forward; media bytes are not touched by this migration.
-- +goose Up
CREATE TABLE video_assets (
    id uuid PRIMARY KEY CHECK (id <> '00000000-0000-0000-0000-000000000000'::uuid),
    lesson_id uuid REFERENCES course_lessons(id) ON DELETE SET NULL,
    generation integer NOT NULL CHECK (generation >= 1),
    state text NOT NULL CHECK (state IN ('created','uploading','queued','processing','ready','failed')),
    claim_token uuid,
    declared_bytes bigint NOT NULL CHECK (declared_bytes BETWEEN 1 AND 2147483648),
    received_bytes bigint NOT NULL DEFAULT 0,
    original_name text NOT NULL CHECK (char_length(original_name) BETWEEN 1 AND 255),
    attempt integer NOT NULL DEFAULT 0 CHECK (attempt >= 0),
    failure_reason text CHECK (failure_reason IS NULL OR char_length(failure_reason) BETWEEN 1 AND 200),
    renditions text[] NOT NULL DEFAULT '{}',
    expires_at timestamptz NOT NULL,
    evicted_at timestamptz,
    created_at timestamptz NOT NULL,
    updated_at timestamptz NOT NULL,
    CONSTRAINT video_assets_received_within_declared CHECK (received_bytes >= 0 AND received_bytes <= declared_bytes),
    CONSTRAINT video_assets_updated_not_before_created CHECK (updated_at >= created_at),
    CONSTRAINT video_assets_evicted_not_before_created CHECK (evicted_at IS NULL OR evicted_at >= created_at),
    CONSTRAINT video_assets_ready_has_renditions CHECK (
        state <> 'ready' OR (
            cardinality(renditions) >= 1
            AND array_position(renditions, NULL) IS NULL
            AND '' <> ALL (renditions)
        )
    ),
    CONSTRAINT video_assets_failed_has_reason CHECK (state <> 'failed' OR failure_reason IS NOT NULL),
    CONSTRAINT video_assets_claim_only_while_processing CHECK (claim_token IS NULL OR state = 'processing')
);

CREATE UNIQUE INDEX video_assets_pending_per_lesson ON video_assets (lesson_id)
WHERE lesson_id IS NOT NULL AND evicted_at IS NULL
    AND state IN ('created','uploading','queued','processing');

CREATE UNIQUE INDEX video_assets_generation ON video_assets (lesson_id, generation);

CREATE INDEX video_assets_by_lesson ON video_assets (lesson_id, generation DESC)
WHERE evicted_at IS NULL;

-- +goose Down
DROP TABLE video_assets;
