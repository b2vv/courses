# Codex brief — ratings T03: migration 00013 + discussion review repository

Repository /Users/strelia/projects/courses, branch `feat/ratings`. Read first (they bind): `AGENTS.md`,
`docs/ratings/spec.md`, `docs/ratings/plan.md` (§3 data model and semantics, §4 writes and ON CONFLICT,
§5 teach-queue row, §7 side effects, §8 migration and rollback, §11 rows Rt02, Rt09, Rt12, Rt13, Rt16,
Rt17, Rt20, Rt22, Rt25, Rt26, Rt27), `docs/adr/013-one-access-rule-as-a-contract.md`. T01 is in the
tree: `backend/internal/discussion/domain/review.go` (`ReviewFacts`, `CanReview`, `ReviewContent`,
`ReviewProgress`) and `backend/internal/contracts/rating`. Log each step to `work/codex-t03.log` (first
line before the first tool: `NEXT:`). Do **not** commit.

Scope: migration + ports + postgres adapter + integration tests. **No application service, no HTTP,
no OpenAPI** (T05). Existing behaviour must not change: the current comment queue (`ListQueue`) and
comment removal stay as they are. Write the integration tests first.

## 1. Migration `backend/migrations/00013_ratings.sql`

Goose, same header style as `00012` (`-- +goose Up`, `SET LOCAL lock_timeout = '5s';`), with a Down.

```sql
CREATE TABLE discussion_reviews (
  id uuid PRIMARY KEY,
  course_id uuid NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
  author_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  stars smallint NOT NULL CHECK (stars BETWEEN 1 AND 5),
  body text NOT NULL CHECK (char_length(btrim(body)) BETWEEN 20 AND 2000),
  completed_lessons integer NOT NULL CHECK (completed_lessons >= 0),
  total_lessons integer NOT NULL CHECK (total_lessons >= completed_lessons),
  created_at timestamptz NOT NULL,
  edited_at timestamptz NULL,
  removed_at timestamptz NULL,
  removed_by uuid NULL REFERENCES users(id),
  CHECK (removed_by IS NULL OR removed_at IS NOT NULL),
  UNIQUE (course_id, author_id)
);
CREATE INDEX discussion_reviews_course_newest ON discussion_reviews (course_id, created_at DESC, id DESC);
CREATE INDEX discussion_reviews_counted ON discussion_reviews (course_id) INCLUDE (stars) WHERE removed_at IS NULL;

CREATE TABLE discussion_review_reports ( -- same shape as discussion_comment_reports in 00010
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  review_id uuid NOT NULL REFERENCES discussion_reviews(id) ON DELETE CASCADE,
  reporter_id uuid NOT NULL REFERENCES users(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  dismissed_at timestamptz, dismissed_by uuid REFERENCES users(id),
  UNIQUE (review_id, reporter_id)
);
CREATE INDEX discussion_review_reports_open ON discussion_review_reports (review_id) WHERE dismissed_at IS NULL;

ALTER TABLE discussion_moderation_log ALTER COLUMN comment_id DROP NOT NULL;
ALTER TABLE discussion_moderation_log ADD COLUMN review_id uuid NULL;
ALTER TABLE discussion_moderation_log ADD CONSTRAINT discussion_moderation_log_one_target
  CHECK (num_nonnulls(comment_id, review_id) = 1);
```
Down (test-only, like earlier Downs): delete log rows with `review_id`, drop the constraint and the
column, restore `comment_id SET NOT NULL`, drop the two tables. **No summary table, counter column or
trigger** (R6, Rt27).

## 2. Ports (`backend/internal/discussion/ports/repository.go`)

Add a `ReviewRepository` interface (and embed it in `Repository` the way `ModerationRepository` is),
plus records. Suggested methods — adjust names to the file's style, keep the semantics:
- `ReviewFacts(ctx, courseID, userID) (ReviewFactsRecord, error)` — one SQL like `lessonFactsQuery`
  (`adapters/postgres/repository.go:20-34`): owner id, status, paid (`price_minor IS NOT NULL`),
  entitled (active `course_grants` row for the user, any source), has progress (a `lesson_progress` row
  for the user on any lesson of the course), completed lessons (`completed_at IS NOT NULL`), total
  lessons (all lessons of the course — a lesson has no publish state, `00003_authoring.sql:35`), and
  whether the user already has a review and whether it is removed. Unknown course → `ErrNotFound`.
- `CreateReview(ctx, review) (ReviewRecord, error)` — `INSERT … ON CONFLICT (course_id, author_id) DO
  NOTHING`; when 0 rows, read the existing row and return a distinct error: removed → a
  `review_removed` error, otherwise `review_already_exists` (add both codes to `domain/errors.go`).
  Never surface a raw unique-violation.
- `UpdateReview(ctx, courseID, authorID, content, progress, editedAt) (ReviewRecord, error)` — only a
  non-removed row of that author; removed → `review_removed`; none → `ErrNotFound`.
- `DeleteReview(ctx, courseID, authorID) error` — hard `DELETE` of the author's non-removed row (reports
  cascade); removed → `review_removed`; none → `ErrNotFound`.
- `GetOwnReview(ctx, courseID, authorID) (*ReviewRecord, error)` — nil when none.
- `ListReviews(ctx, courseID, cursor, limit) (ReviewPage, error)` — non-removed only, newest first,
  keyset on `(created_at DESC, id DESC)`, author `display_name` joined like comments
  (`repository.go:38`, deleted users handled as comments do).
- `ReviewSummary(ctx, courseID) (count, starsSum int, err)` — non-removed rows only.
- `InsertReviewReport(ctx, reviewID, reporterID, at)` — idempotent on `(review_id, reporter_id)`;
  reporting your own review → `report_own_review` error (new code).
- `DismissReviewReports(ctx, reviewID, actorID, at)`.
- `RemoveReview(ctx, reviewID, actorID, actorRole, reason, at) error` — **one transaction**: set
  `removed_at/removed_by` on a non-removed row AND insert the moderation-log row with `review_id`
  (`comment_id` NULL, `action` like the comment removal action). Either both or neither.
- `ListModerationQueue(ctx, courseID, cursor, limit) (ModerationQueuePage, error)` — **new method; do
  not change `ListQueue`**. One SQL: `UNION ALL` of open comment-report aggregates (as `queueQuery`,
  `adapters/postgres/moderation.go:32-49`) and open review-report aggregates, each row with
  `kind` ('comment' | 'review'); order `open_count DESC, latest_report DESC, kind DESC, id DESC`; cursor
  carries all four; review rows have null lesson fields. Removed items excluded, as today.

## 3. Adapter (`backend/internal/discussion/adapters/postgres/`)

Implement in a new `review.go` (and touch `moderation.go` only to add the new queue). Map rows to
records explicitly; no generated types leak out. Map a transient DB failure to
`domain.ErrDependencyUnavailable` the way the existing methods do.

## 4. Integration tests (`repository_integration_test.go`, real Postgres)

Give the new subtests their **own** fixtures and id source (no shared id slice), and scope every count
to your own course/users. Users use the existing Argon2id fixture hash (`seedDiscussion`, line ~546).
Cover, naming each subtest after its row:
- **Rt02** facts for a buyer with an active grant and progress on a paid course → entitled, has progress.
- **Rt16** preview-only watcher (progress on an `is_preview` lesson, no grant) → paid, not entitled, has
  progress; `domain.CanReview` on those facts → `review_purchase_required`.
- **Rt17** admin with progress, no grant, paid → same refusal (role is not a fact here).
- **Rt12** review written at 2 of 4 completed lessons; then add 2 lessons and complete all → the stored
  review still has 2/4.
- **Rt13** then `UpdateReview` with freshly read facts → 6/6 stored, `edited_at` set.
- **Rt09** three reviews, one removed via `RemoveReview` → `ReviewSummary` = 2 and the right sum;
  `ListReviews` omits the removed one.
- **Rt22** `RemoveReview` writes exactly one log row with `review_id` and the actor. Atomicity: in a
  serial (non-parallel) subtest, add a temporary constraint
  `ALTER TABLE discussion_moderation_log ADD CONSTRAINT test_force_fail CHECK (reason IS DISTINCT FROM 'force-fail') NOT VALID`,
  drop it in `t.Cleanup`, call `RemoveReview` with reason `'force-fail'` → error, and assert the review
  is still not removed and no log row was written.
- **Rt20** grant revoked after the review (`UPDATE course_grants SET state='revoked'` in the fixture) →
  the review stays in `ListReviews` and `ReviewSummary`; facts now show not entitled.
- **Rt25** own review report refused; two reports by the same user → one row.
- **Rt26** course archived → `ListReviews`/`ReviewSummary` unchanged; facts for a grant holder still
  return entitled.
- **Rt27** schema: no table/column/trigger beyond the two tables (query `information_schema` /
  `pg_trigger` for `discussion_reviews`).
- concurrency: two concurrent `CreateReview` for the same author/course → one success, one
  `review_already_exists`; create after removal → `review_removed`; delete after removal →
  `review_removed`; delete then create again → success.
- moderation log: a comment entry and a review entry both insert; a row with both or neither target
  is rejected by the CHECK.
- `ListModerationQueue` returns both kinds in the specified order and pages correctly across kinds
  with the four-part cursor.

## Verify
`DOCKER_CONFIG=/tmp/claude-501/dockercfg make format-check lint typecheck test integration
check-generated`. If Docker is denied in your sandbox, say so — Courses 1 runs the gate; still run
`go build ./...`, `go vet ./...` and the non-Docker tests. Never claim a gate you did not run.

## Channels
If a contract above is missing something you need, stop and say what. If a rule here contradicts the
spec, the plan or the code in a way that would produce a bug, argue it with the code path instead of
silently diverging. Final answer: changed files, tests added (names), what you ran with results, what
you could not.
