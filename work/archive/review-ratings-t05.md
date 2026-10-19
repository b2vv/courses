# Review — ratings T05 (commit aef982e, branch `feat/ratings`)

Reviewer: Courses 2. Date: 2026-10-03. Against: `docs/ratings/spec.md` (AC1, AC4, AC6, AC7, AC9),
`docs/ratings/plan.md` §4/§5/§10, briefs `work/brief-codex-t05*.md`, carry-over `work/_t05-carryover.md`.
Scope: discussion application review methods + HTTP routes + status mapping + OpenAPI/Orval + teach FE
adapter (kind accepted, not rendered).

## Verdict

**Approve. 0 blocking, 4 non-blocking.** All six focus points PASS; both constitution fixes are correct;
and the T03 one-target-CHECK gap I flagged is now closed by an integration test. One non-blocking item
is a genuine AC10 question worth confirming before T06/e2e.

Evidence I ran myself:
- `go build ./...` → **BUILD_OK**; `go vet ./internal/discussion/...` → **VET_OK**.
- `go test ./internal/discussion/… ./internal/contracts/rating/…` → **all ok** — application (incl. the
  new review unit tests), http (incl. `TestT05ReviewRoutesContract` + `TestT05ReviewRoutesGaps`), domain.
- I did **not** run the Docker integration suite (Courses 1 reports exit 0); the one-target-CHECK and
  ListReviews-keyset integration tests are present and I read them statically.

## Focus 1 — AC6 delete without eligibility — PASS

`DeleteReview` (`application.go`) calls `repository.DeleteReview(courseID, actor.UserID)` with **no
`CanReview`** — scoped to the actor's own row. The unit test deletes with facts set to `Archived` (no
access) and it succeeds; a second delete on a removed row → `review_removed`. So the author can delete
at any time, but only their own review (repo `WHERE author_id = …`).

## Focus 2 — G9 admin-only remove + re-remove no-op — PASS

`RemoveReview` gates on `actor.Administrator` (not owner∨admin as comments do) → non-admin, owner
included, gets `review_moderation_forbidden`. It then `GetReview`s and, **if already removed, returns
`nil` before validating the reason or writing** — a true no-op, no second log row (matches comment
removal). Reason is validated 1–500 runes like comments. Contract test proves it over HTTP: owner → 403
`review_moderation_forbidden` (Rt21), admin → 204, admin again → 204 (idempotent).

## Focus 3 — PutReview order + G4 edit-after-losing-access — PASS

Order is exactly CanReview → `NewReviewContent` → `NewReviewProgress` → `facts.Removed` → create/update
(`application.go`). Because `CanReview` is first and independent of whether the review exists, **an edit
by an author who lost access is refused** (G4): the eligibility-order unit test drives all five refusals
through `PutReview`, and the paid-without-grant case returns `review_purchase_required`. Progress is
**recaptured on edit**: the create-then-update test changes facts to 4/6 and asserts the update stores
`{Completed:4, Total:6}`, not the create-time 1/4 (G6). The repo (`UpdateReview WHERE removed_at IS
NULL`, `CreateReview ON CONFLICT`) is the source of truth, so the `facts.Removed` pre-check is not a TOCTOU
hole.

## Focus 4 — public list only for published courses — PASS

`PublicReviews` resolves the slug via `PublishedReviewCourse`, whose SQL is `… WHERE slug=$1 AND
status='published'` → unknown/draft/archived course → `ErrNotFound` → 404 `course_not_found`. Unit test:
archived slug → `ErrNotFound`; contract test: `GET /courses/archived/reviews` → 404. (See non-blocking #1
for the AC10 implication for grant holders.)

## Focus 5 — status codes vs plan §5 — PASS (one prose mismatch, not a code defect)

`statusFor` maps: **403** owner_forbidden / not_available / purchase_required / not_started /
moderation_forbidden / report_own_review; **409** already_exists / removed; **400** review_invalid /
invalid_cursor; **503** dependency_unavailable. Handlers: PUT create **201** / edit **200**; DELETE,
report, remove, dismiss **204**; not-found **404** (`course_not_found` on course routes,
`review_not_found` on review routes). Every review handler sets `Cache-Control: no-store`, `publicReviews`
included (a defect that was fixed). The contract test asserts each of these and runs against the OpenAPI
spec. **Plan §5 prose says remove → "200"**, but the implementation returns **204**, which I verified
matches the established house style (comment `remove` = 204, `dismissReports` = 204); OpenAPI + the
contract test agree at 204. Harmonise the plan wording; the code is consistent.

## Focus 6 — Rt19 = repeat PUT updates — PASS

`created := !facts.Exists`: an existing non-removed review takes the `UpdateReview` branch → **200**, not
409. `review_already_exists` (409) is therefore reachable only via the `CreateReview` ON CONFLICT race
(two concurrent creates). Contract test: a second PUT with a 2000-rune body returns **200** ("should
update"). Matches the updated Rt19 contract (commit 55860a1).

## Constitution fixes — both correct

1. **`domain.NewRatingSummary`** (`domain/review.go`) wraps `contracts/rating.Summary`; the application
   now calls the domain, not the contract. A domain **may** import a contract (ADR 013), and application
   importing a contract was the violation — so the arch test is green. Returns `*RatingSummary` (nil below
   threshold); `PublicReviews` maps nil → no rating.
2. **Adapter id** `COALESCE(NULLIF($1,'')::uuid, gen_random_uuid())`: `PutReview` passes an empty `ID`, so
   the DB generates it (no more `github.com/google/uuid` in the application); a non-empty id is still
   honoured, so T03's explicit-id integration tests keep working. Same `NULLIF` generalisation applied to
   `WriteModerationLog` for `comment_id`/`review_id` → the one-target CHECK holds for both kinds and the
   existing comment-removal/dismiss path is unchanged.

## Follow-through I want to credit

The **T03 non-blocking gap I flagged** — no test that the moderation-log one-target CHECK rejects a
both-set or neither-set row — is **now closed**: integration subtest "moderation log rejects both targets
and neither target" asserts SQLSTATE 23514 on `discussion_moderation_log_one_target` for both inserts.
A ListReviews keyset-paging integration test was also added.

## Auth / CSRF wiring — correct

Public GET open; `GET …/review` is `authenticate` only (anonymous → 401 `authentication_required`, Gaps
test); every write (PUT/DELETE/report/remove/dismiss) is `sameOrigin(authenticate(...))`. Cross-origin
refusal is a pass-through in the handler harness → left to e2e (the carry-over's known gap; the wrapper is
applied to all writes, so the e2e has something real to exercise).

## Non-blocking

1. **AC10 vs plan §5 — archived-course review list for grant holders (confirm before T06/e2e).** The
   review **list** is served only by the public endpoint, which 404s for an archived course. A grant
   holder on an archived course can still read/edit *their own* review (`OwnReview`/`PutReview` work —
   `CanReview` passes on `Archived + Entitled`) and see the rating summary (learning, T04), but cannot
   fetch the **list** of others' reviews. AC10 says reviews stay "visible to those who can see the
   course," and Rt26's e2e is "grant holder opens the course, reviews visible." T05 faithfully implements
   plan §5 (list routed only through the public/published endpoint), so this is a plan/spec question, not
   a T05 code defect — but T06/e2e will hit it. Confirm how an archived course's review list reaches a
   grant holder (a learning-scoped list, or relaxing the public endpoint for entitled viewers).
2. **`deleteReview` with no review to delete → 404 `course_not_found`.** `DeleteReview` returns
   `ErrNotFound` both for a missing course and for "no review of mine to delete"; `writeReviewError` maps
   it to `course_not_found`, which is misleading when the course exists but the review does not
   (`review_not_found` would be truer). Plan §5 doesn't define this case; minor labelling.
3. **Two thin spots in the unit tests.** G4 (edit after losing access) is proven only implicitly — the
   eligibility-order test uses the create path (`facts.Exists=false`); an explicit `Exists=true` +
   lost-access edit case would nail it (the code path is identical, so this is coverage, not a bug). And
   the re-remove no-op test asserts both calls return `nil` but not `f.removed == 1`.
4. **(carry) cross-origin refusal** is untestable in the handler harness → e2e. Acknowledged.

## Rows advanced (plan §12)

Contract level of Rt01–Rt04, Rt10, Rt11, Rt14 (public page/cursor), Rt15–Rt19, Rt21, Rt23–Rt25; the
one-target CHECK and ListReviews keyset paging at integration. UI parts and the archived-list e2e stay for
T06/T07.
