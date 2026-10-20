-- B-21: a read with no language asked picked the first language by normalised code, so a second
-- transcript added to a lesson (one sorting earlier) would silently become the canonical read —
-- a lesson with uk gains en and every reader flips to en, with no signal to the author. This
-- marks the canonical transcript explicitly. The read prefers it, and adding another language no
-- longer displaces it. 00007 noted "the second needs no migration" for the key; the key still
-- needs none — this adds which of several is primary, not the language itself.
-- +goose Up
SET LOCAL lock_timeout = '5s';

ALTER TABLE transcripts ADD COLUMN is_primary boolean NOT NULL DEFAULT false;

-- Exactly one primary per lesson. This partial unique index is the sole keeper of that fact:
-- the write path marks the first transcript of a lesson primary and leaves later languages
-- non-primary, and the index refuses a second primary rather than trusting the write path alone.
CREATE UNIQUE INDEX transcripts_one_primary ON transcripts (lesson_id) WHERE is_primary;

-- Backfill: mark the first transcript by normalised code per lesson primary — the same row the
-- old alphabetical read already returned, so no existing read changes at migration time.
-- Production carries exactly one language per lesson today, so this marks that one.
UPDATE transcripts AS t SET is_primary = true
WHERE t.language = (
    SELECT pick.language FROM transcripts AS pick
    WHERE pick.lesson_id = t.lesson_id
    ORDER BY pick.language
    LIMIT 1
);

-- +goose Down
SET LOCAL lock_timeout = '5s';

DROP INDEX transcripts_one_primary;
ALTER TABLE transcripts DROP COLUMN is_primary;
