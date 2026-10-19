# Codex brief — ratings T05b: learning-scoped review list + two T05 review notes

Repository /Users/strelia/projects/courses, branch `feat/ratings`. Read first: `AGENTS.md`,
`docs/ratings/assumptions.md` (last entry, 2026-10-03 — the decision), `docs/ratings/plan.md` §5 (new row
`GET /api/v1/learn/courses/{courseId}/reviews`) and §11 row Rt30, `work/review-ratings-t05.md` (notes 2, 3).
Pattern: the T05 code in `discussion/application/application.go` and `discussion/adapters/http/handler.go`.
Log each step to `work/codex-t05b.log` (first line before the first tool: `NEXT:`). Do **not** commit.
Write the tests first, and write ALL tests listed below before you stop — do not hand back without them.

## 1. Learning-scoped list (Rt30, AC10)

- Application `LearningReviews(ctx, actor, courseID, cursor)`: signed in; the course must be viewable by
  the actor through the existing rule (`domain.Course.CanBeViewedBy` — grant holder, owner, admin; an
  archived course is viewable for a grant holder). Not viewable → `ErrForbidden` (`course_forbidden`, 403);
  unknown course → not found (404). Returns the same shape as `PublicReviews`: `rating` via
  `domain.NewRatingSummary`, items from `ListReviews` (20, newest first, keyset cursor).
  The course facts you need are owner, status, paid, entitled — read them with one repository method (reuse
  `ReviewFacts` if it already carries them; otherwise add a small read). No new port.
- Route `GET /api/v1/learn/courses/{courseId}/reviews?cursor=` — `authenticate`, `Cache-Control: no-store`,
  same response schema as the public list (reuse the OpenAPI schema). OpenAPI + Orval regenerate.

## 2. Review notes from T05

- `DELETE /api/v1/learn/courses/{courseId}/review` when the author has no review → **404 with code
  `review_not_found`** (today it says `course_not_found`). Add the code to the OpenAPI enum if missing.
- Unit: an author whose review exists and who has **lost access** (facts: paid, not entitled) → `PutReview`
  refused with `review_purchase_required` and the stored review unchanged (G4 on the edit path).
- Unit: `RemoveReview` by an admin twice → the repository remove is called **once** (re-remove is a no-op).

## Tests (all required)
- application unit: LearningReviews — grant holder on an archived course → items + rating; no right → 403
  error; unknown → not found; the two T05 notes above.
- HTTP contract (`handler_test.go`, `assertContract`): learning list 200 shape for an archived course with a
  grant; 403 `course_forbidden` without a right; 401 anonymous; DELETE without a review → 404
  `review_not_found`.
- integration is not required here (reads already covered); do not change fixtures.

## Verify
`go vet ./internal/discussion/...`, `go test ./internal/discussion/...`, frontend `npm run typecheck` after
Orval. Docker steps (`make generate`, integration, check-generated) are run by Courses 1 — say if denied.
Final answer: tests added (names), what you ran, what you could not.
