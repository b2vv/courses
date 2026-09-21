-- Down drops every machine-made draft. That is safe in a way 00007's Down is not: a draft is
-- regenerated from the asset it names, while a transcript was typed by a human and nothing
-- reproduces it. Retained environments still recover forward.
-- +goose Up
SET LOCAL lock_timeout = '5s';

-- One row per lesson, and the segments stay an opaque document (plan D8-11). The transcript
-- normalises its segments into a table because search and player sync walk into them; nobody
-- walks into a draft - the editor reads it whole - so a mirror of transcript_segments would buy
-- a second index and a second set of queries with no reader.
--
-- `updated_at` is the optimistic-lock token, exactly as in `transcripts`. The plan sketched a
-- separate `token` column, but it appeared only in that sketch and in the migration draft: no
-- port, no design decision and no acceptance row referenced it, while D8-10 sends publication
-- through the existing `SaveLessonTranscript`, which locks on `expectedUpdatedAt`. A second
-- mechanism for the same job is the thing to drop, not to carry.
CREATE TABLE asr_drafts (
    lesson_id uuid PRIMARY KEY REFERENCES course_lessons(id) ON DELETE CASCADE,

    -- Which asset this text was made from, and which generation of it. Staleness is DERIVED by
    -- comparing this generation with the one the player would pick, the same way the transcript
    -- does it (00007, `authored_against_generation`) - no stored flag that can drift.
    --
    -- The cascade is asymmetric with `video_assets.lesson_id ON DELETE SET NULL`, and that is
    -- deliberate on both sides: an asset outlives its lesson because the bytes are still on
    -- disk, while a draft without its lesson is nobody's.
    --
    -- A swept asset takes its draft too, and that is the owner's decision of 2026-09-21
    -- (D8-15), taken with its price named rather than inferred. The price: eviction is soft
    -- (`evicted_at`), but `DeleteSweptMediaAsset` later issues a real DELETE, so an author who
    -- edited a draft, did not publish it and uploaded a new video loses those edits to a
    -- BACKGROUND task - nothing they did, and nothing they can see coming. `SET NULL` would
    -- have kept the text at the cost of a draft pointing at no video. C22 pins this cascade so
    -- that a later reader cannot quietly "fix" it into SET NULL believing they are helping.
    asset_id uuid NOT NULL REFERENCES video_assets(id) ON DELETE CASCADE,
    generation integer NOT NULL CHECK (generation >= 1),

    state text NOT NULL CHECK (state IN ('queued', 'running', 'ready', 'failed', 'no_speech')),

    -- Nullable because a draft that has not finished, or found no speech, has no language to
    -- report. The pattern is the transcript's, with one addition: `auto` is the UI's word for
    -- "detect it", it satisfies the BCP-47 shape, and this table is where it would enter from a
    -- select. D8-8 forbids it reaching the field; the domain refuses it and so does this.
    language text CHECK (
        language IS NULL OR (
            char_length(language) BETWEEN 2 AND 35
            AND language = lower(language)
            AND language ~ '^[a-z][a-z0-9]{1,7}(-[a-z0-9]{1,8})*$'
            AND language <> 'auto'
        )
    ),

    -- The editor's own line shape, so that publication needs no translator (plan P-2). The
    -- count ceiling mirrors `domain.MaximumSegments`; the per-line rune ceiling cannot live
    -- here, because a CHECK cannot walk array elements, so D8-12 keeps it in the domain and
    -- refuses a looped model output at the module boundary rather than an hour later on save.
    segments jsonb NOT NULL DEFAULT '[]'::jsonb CHECK (
        jsonb_typeof(segments) = 'array' AND jsonb_array_length(segments) <= 5000
    ),

    -- Named cause, and only where a cause exists. C13 asks that silence produce a NAMED result
    -- rather than an empty draft that looks like finished work; the two CHECKs below make that
    -- structural instead of hoping the writer remembers.
    failure text CHECK (failure IS NULL OR char_length(failure) BETWEEN 1 AND 200),

    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),

    CHECK (updated_at >= created_at),
    -- A ready draft has text and a language; an empty one would be indistinguishable from work
    -- nobody did.
    CHECK (state <> 'ready' OR (jsonb_array_length(segments) > 0 AND language IS NOT NULL)),
    -- A failure carries its reason, and only a failure does.
    CHECK ((state = 'failed') = (failure IS NOT NULL))
);

-- Serves recovery alone (plan P-5): the startup sweep asks for drafts left `running` by a worker
-- that died. Reading a draft goes through the primary key, so this index has exactly one caller
-- and stays partial.
CREATE INDEX asr_drafts_running ON asr_drafts (updated_at) WHERE state = 'running';

-- +goose Down
DROP TABLE asr_drafts;
