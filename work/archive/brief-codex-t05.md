# Codex brief — ratings T05: reviews application + HTTP + OpenAPI (discussion module)

Repository /Users/strelia/projects/courses, branch `feat/ratings`. Read first (they bind): `AGENTS.md`,
`docs/ratings/spec.md` (AC1, AC4, AC6, AC7, AC9, G1–G10), `docs/ratings/plan.md` (§4 eligibility and codes,
§5 HTTP contract, §7 «Dependency failure», §11 rows Rt01–Rt04, Rt10, Rt11, Rt14–Rt19, Rt21, Rt23–Rt25),
`work/_t05-carryover.md` (review findings to close here). In the tree: domain `discussion/domain/review.go`,
repository `discussion/adapters/postgres/review.go` + ports `ReviewRepository`, `ListModerationQueue`,
`contracts/rating`. Follow the existing comment path as the pattern: `discussion/application/application.go`
(Publish/Update/Remove/ReportComment/Queue/DismissReports/requireModerationAccess) and
`discussion/adapters/http/handler.go` (routes at lines 24–34, `statusFor` ~line 388). Log each step to
`work/codex-t05.log` (first line before the first tool: `NEXT:`). Do **not** commit.

Scope: discussion application + HTTP + OpenAPI + Orval + contract/unit tests, and the minimal frontend
adapter change that keeps `typecheck` green after Orval. **No UI work** (T06). Write tests first.

## 1. Application (`discussion/application`)

New service methods (names in the file's style):
- `PublicReviews(ctx, slug, cursor)` — anyone (no actor). Course resolved by slug; **only a published
  course** (an archived or draft course → `ErrNotFound`, like the catalog's course page). Returns
  `rating` from `rating.Summary(ReviewSummary(...))` (nil below 3) and the page from `ListReviews` (20,
  newest first, keyset cursor).
- `OwnReview(ctx, actor, courseID)` — signed in. Returns own review (or nil) and `eligibility`: `"ok"` or the
  code from `domain.CanReview(facts, actor)` for the current facts. Unknown course → `ErrNotFound`.
- `PutReview(ctx, actor, courseID, stars, body)` — signed in. `CanReview` first (codes §4); then
  `NewReviewContent`; then progress `NewReviewProgress(completed, total)` from the **same facts read**; if the
  actor has no review → create (repository maps duplicates to `review_already_exists` / `review_removed`);
  if they have a non-removed one → update (recapture progress, G6); removed → `review_removed`.
  Return the stored review and whether it was created (201) or updated (200).
- `DeleteReview(ctx, actor, courseID)` — author only, **no eligibility check** (AC6: delete at any time,
  even after losing access); removed → `review_removed`; none → `ErrNotFound`.
- `ReportReview(ctx, actor, reviewID)` — signed in; own review → `report_own_review`; idempotent.
- `RemoveReview(ctx, actor, reviewID, reason)` — **administrator only** (G9): anyone else, the course owner
  included → `review_moderation_forbidden`. An already-removed review → **no-op success**, matching comment
  removal (`application.go:203-204`). To tell "already removed" from "missing" add a repository read of a
  review by id including removed ones (the current `RemoveReview` returns `ErrNotFound` for both — T03
  review finding №2). Missing → `ErrNotFound`.
- `DismissReviewReports(ctx, actor, reviewID, reason)` — owner ∨ admin, via the existing
  `requireModerationAccess` on the review's course.
- `Queue` (teach) — switch to `ListModerationQueue`: items carry `kind` (`comment` | `review`); review items
  carry stars and body and no lesson. Keep the owner ∨ admin guard.

Unit tests with a fake repository for every method: each refusal code in §4 order; create vs update;
progress recaptured on update; delete without eligibility; owner cannot remove (Rt21); admin removes;
re-remove no-op; report own refused.

## 2. HTTP + OpenAPI (`discussion/adapters/http`, `backend/openapi/openapi.yaml`)

| Method · path | Auth | Success | Notes |
|---|---|---|---|
| `GET /api/v1/courses/{slug}/reviews?cursor=` | **none** (public, no `authenticate`) | 200 `{rating: {average, count} \| null, items: [...], nextCursor}` | item: `id, authorDisplayName, stars, body, completedLessons, totalLessons, createdAt, editedAt` |
| `GET /api/v1/learn/courses/{courseId}/review` | authenticate | 200 `{review: … \| null, eligibility}` | |
| `PUT /api/v1/learn/courses/{courseId}/review` | sameOrigin + authenticate | 201 created / 200 updated | body `{stars, body}` |
| `DELETE /api/v1/learn/courses/{courseId}/review` | sameOrigin + authenticate | 204 | |
| `POST /api/v1/learn/reviews/{reviewId}/reports` | sameOrigin + authenticate | 204 (also on repeat) | |
| `POST /api/v1/teach/reviews/{reviewId}/remove` | sameOrigin + authenticate | 204 (also when already removed) | admin only |
| `POST /api/v1/teach/reviews/{reviewId}/dismiss-report` | sameOrigin + authenticate | 204 | owner ∨ admin |
| `GET /api/v1/teach/courses/{courseId}/moderation-queue` (existing) | authenticate | 200 | items gain `kind` + review fields; lesson fields nullable |

Status mapping (extend `statusFor`): **403** `review_owner_forbidden`, `review_not_available`,
`review_purchase_required`, `review_not_started`, `review_moderation_forbidden`, `report_own_review`;
**409** `review_already_exists`, `review_removed`; **400** `review_invalid` (and malformed JSON / wrong
types as the comment handlers do); **404** unknown course / review; **503** `dependency_unavailable`. Add
every new code to the OpenAPI error-code enum. Average is a JSON number. Regenerate Orval.

Contract tests (`handler_test.go`, the existing `assertContract` style) for each route: the success shape
and each status above at least once, including **Rt10** (19/20/2000/2001 runes after trim → 400/201/201/400),
**Rt11** (stars 0/1/5/6), **Rt15** owner 403, **Rt18** not started 403, **Rt19** second create → the PUT
updates instead (one review per user): assert that a PUT by an author who already has a review returns
200 and does not create a second row; **Rt21** owner remove → 403; **Rt23** create after removal and delete
after removal → 409 `review_removed`; **Rt24**-style admin path unaffected; **Rt14** public list empty and
paged (25 reviews → 20 + cursor → 5); public list of an archived course → 404; an anonymous public request
succeeds without a session cookie. Carry-over: assert `rating.average` is a number in the public response.

## 3. Carry-over integration test (T03 review №1)

In `discussion/adapters/postgres/repository_integration_test.go`: a moderation-log row with **both**
`comment_id` and `review_id`, and one with **neither**, is rejected by the one-target CHECK. Own ids; no
shared state changes left behind.

## 4. Frontend — only to keep it compiling

After Orval, fix the teach-queue adapter/mapping so `npm run typecheck` and the existing frontend tests pass
with the new `kind` field (review items may be mapped but not rendered yet — T06 renders them). No new UI.

## Verify
`DOCKER_CONFIG=/tmp/claude-501/dockercfg make generate format-check lint typecheck test integration
check-generated`. If Docker is denied in your sandbox, say so — Courses 1 runs the gate; still run
`go build ./...`, `go vet ./...`, the non-Docker Go tests and the frontend `typecheck` + tests. Integration
fixtures: users need the existing Argon2id fixture hash (a plain string fails the users CHECK); one
statement per parameterised `Exec`; never mutate shared seeded rows without restoring them in
`t.Cleanup`; scope counts to your own rows. Never claim a gate you did not run.

## Channels
If a contract above is missing something you need, stop and say what. If a rule here contradicts the
spec, the plan or the code in a way that would produce a bug, argue it with the code path instead of
silently diverging. Final answer: changed files, tests added (names), what you ran with results, what
you could not.
