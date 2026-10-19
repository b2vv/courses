# Review — ratings T03 (commit 18d2589, branch `feat/ratings`)

Reviewer: Courses 2. Date: 2026-10-03. Against: `docs/ratings/spec.md`, `docs/ratings/plan.md`
§3/§4/§5(queue)/§7/§8, brief `work/brief-codex-t03.md`. Scope: migration 00013 + discussion ports +
postgres adapter + integration tests. No application/HTTP/OpenAPI (T05) — correctly respected.

## Verdict

**Approve. 0 blocking, 3 non-blocking** (one is a test gap the brief asked for; two are observations,
one of them not even T03). All five focus points Courses 1 named PASS. The three self-fixed test
defects do **not** weaken what the rows assert — I traced the subtest sequence and the counts.

Evidence I ran myself (not trusted from the commit message):
- `go build ./internal/discussion/... ./internal/contracts/...` → **OK**.
- `go test -tags integration -c ./internal/discussion/adapters/postgres/` (compile the integration
  suite without running) → **OK**.
- `go vet ./internal/discussion/...` → clean; `go test ./internal/discussion/... ./internal/contracts/rating/...`
  (non-integration) → **ok** (http adapter, domain, rating).
- I did **not** run the Docker integration suite myself (reaper/credential-helper hang risk per the
  repo's own lessons); Courses 1 reports 12/12 ratings subtests green, and I verified the test logic
  and counts statically.

## Focus 1 — RemoveReview atomicity (update + log in one tx) — PASS

`review.go:164` — `Begin` → `UPDATE … SET removed_at/removed_by … WHERE removed_at IS NULL RETURNING
course_id` → `INSERT … discussion_moderation_log(… review_id …)` → `Commit`, with `defer Rollback`.
Both or neither. The `action` is `'remove'`, matching the comment path (`application.go:212`), and the
log column `action` has no CHECK, so it is accepted. Rt22 proves atomicity with a `NOT VALID`
force-fail CHECK on `reason`: the log INSERT fails, tx rolls back, and the test asserts the review is
**still not removed** and **0 log rows** — a real atomicity proof, not a review of the happy path.

## Focus 2 — ON CONFLICT mapping (no raw unique error) — PASS

`CreateReview` (`review.go:45`): `INSERT … ON CONFLICT (course_id,author_id) DO NOTHING RETURNING`.
0 rows (conflict) → `scanReview` maps `pgx.ErrNoRows`→`ErrNotFound` → reads the existing row → removed
→ `ErrReviewRemoved`, else `ErrReviewAlreadyExists`. No raw unique-violation ever surfaces. The
concurrency subtest runs two `CreateReview` in goroutines and asserts exactly 1 success + 1
`review_already_exists`, then create-after-removal → `review_removed`, delete-after-removal →
`review_removed`, delete-then-create → success. `InsertReviewReport` is likewise `ON CONFLICT DO
NOTHING` (idempotent) and refuses the author's own report with `report_own_review`.

## Focus 3 — facts SQL — PASS

`reviewFactsQuery` (`review.go:17`):
- **entitled** = `EXISTS (course_grants … state='active')` with **no source filter** → any active grant
  (purchase or admin). ✔
- **total_lessons** = `count(*)` of `course_modules ⋈ course_lessons` for the course, **no publish/preview
  filter** → all lessons of the course (plan §3). ✔
- **has_progress** = `EXISTS lesson_progress` on any lesson of the course, **no is_preview filter** →
  preview-lesson progress counts as progress. Entitlement is a separate fact, so a preview-only watcher
  on a paid course is `Paid=true, Entitled=false, HasProgress=true` → `CanReview` → `review_purchase_required`
  (Rt16, asserted). Role is never a fact here (Rt17 asserts admin without grant → same refusal). ✔
- **completed_lessons** = `count(… completed_at IS NOT NULL)`; owner = `instructor_id`; paid =
  `price_minor IS NOT NULL`. Lesson↦one module, `lesson_progress` unique per (user,lesson), so no
  double counting and `completed ⊆ total`. ✔
- Unknown course → `ErrNotFound`; transient DB error → `persistenceError` (→ `dependency_unavailable`).
- Rt12/Rt13 prove freeze at the DB layer: created 2/4, then +2 lessons all completed, `GetOwnReview`
  still 2/4; `UpdateReview` with freshly-read facts → 6/6 and `edited_at` set, same row id.

## Focus 4 — UNION ALL queue keyset across kinds — PASS

`moderationQueueQuery` (`review.go:261`): CTE unions open comment-report aggregates (as `queueQuery`)
and open review-report aggregates, each tagged `kind`; removed items excluded in both branches. Keyset:
`(open_count, latest_report, kind, id) < ($2,$3,$4,$5)` with `ORDER BY open_count DESC, latest_report
DESC, kind DESC, id DESC` — row-comparison is exactly the strict-after predicate for all-DESC order.
Types line up (`kind` text↔`$4::text`, `id` uuid↔`$5::text::uuid`, projected `id::text` round-trips
through the cursor). `kind DESC` ⇒ 'review' before 'comment' on ties (plan §5's `(kind,id) DESC`).
First page is NULL-guarded (`$2::int IS NULL OR …`). Subtest pages 1+1 and asserts the two items are
different kinds. **`ListQueue` and `moderation.go` are untouched** (0 changes in the commit) → existing
comment queue behaviour preserved, as the brief required.

## Focus 5 — Down migration — PASS

Order is FK/constraint-safe: `DELETE … WHERE review_id IS NOT NULL` (so `comment_id SET NOT NULL`
won't fail on the review rows, which are exactly the ones with NULL comment_id) → drop the one-target
constraint → drop `review_id` → restore `comment_id NOT NULL` → drop `discussion_review_reports` (FK
child) → drop `discussion_reviews`. Up matches plan §3 exactly (both partial indexes, the two CHECKs,
UNIQUE(course_id,author_id)); no summary table / counter / trigger (Rt27 asserts tables=2, 0 rating
columns on `courses`, 0 non-internal triggers on `discussion_reviews`).

## The three self-fixed test defects — none weaken the assertions

1. **delete-after-removal called with the wrong author** → now `DeleteReview(fx.course, fx.concurrent)`,
   i.e. the removed review's real author, so it actually exercises the removed-slot-blocks-delete path
   and asserts `ErrReviewRemoved`. This **strengthens** the test (the old call hit `ErrNotFound` for the
   wrong reason). ✔
2. **queue subtest depended on a later subtest's report** → it now files its own report on review
   `…0005` (fx.reporter's review from the atomic-removal subtest, which survived the force-fail
   rollback and is not removed) instead of borrowing Rt25's. Still asserts both kinds present and
   cross-kind paging via the 4-part cursor. Self-contained, order-independent, no weaker. ✔
3. **Rt26 asserted a fixed count** → now reads `before`/`beforeCount`/`beforeSum`, archives, then asserts
   list length and (count,sum) unchanged, **with a `len(page.Items)==0` guard so it can't pass
   vacuously**. This is the correct before/after pattern and still proves AC10/G3 (archive changes
   nothing). ✔ (Exactly the "scope a count to before/after, don't assert a fixed number" lesson.)

I also verified the subtest counts are internally coherent: Rt12/13 leaves reviewA at 5★; Rt20 revokes
the grant but keeps the review (count still 1); Rt09 adds second(2★)+third(5★, removed) → asserts
count=2, sum=7 (5+2). Consistent across the ordered, non-parallel sequence.

## Non-blocking

1. **Missing CHECK test (brief asked for it).** The brief's "moderation log" bullet wanted a test that
   a log row with **both or neither** target is rejected by `num_nonnulls(comment_id, review_id)=1`.
   There is no such subtest — the one-target CHECK's enforcement is untested (only the one-target happy
   path, via Rt22, is). Low risk (declarative Postgres CHECK) but it is the guard that keeps the shared
   log coherent now that `comment_id` is nullable. Add a 2-line integration assertion (insert both-set
   and both-null → expect error). Could land in T05 if not here.
2. **Already-removed RemoveReview → `ErrNotFound`.** `UPDATE … WHERE removed_at IS NULL` returns no row
   for an already-removed review, so a double-remove is indistinguishable from a missing review
   (`ErrNotFound`). Acceptable for the repo layer, but T05 should decide the HTTP status for a
   re-remove (plan §5 doesn't name it). Not a defect.
3. **Observation, not T03.** The working tree does **not** `go build` right now — but the failures are
   all in `internal/course` (`row.ReviewCount/ReviewStarsSum/LessonCount undefined`), which is
   **uncommitted T04 WIP** (`courses.sql` is modified, sqlc not yet regenerated). The T03 commit touches
   0 files under `internal/course` and builds cleanly in isolation. Flagging only so T04's gate runs
   `make generate` before building; nothing for T03 to fix.

## Rows covered by this slice (plan §12 / brief)

Rt02, Rt09, Rt12, Rt13, Rt16, Rt17, Rt20, Rt22, Rt25, Rt26, Rt27, plus concurrency and the cross-kind
queue — all present as integration subtests, each named after its row, with own fixtures/ids
(`seedRatings`, 71/72/73/74/75 prefixes) and counts scoped to its own course.
