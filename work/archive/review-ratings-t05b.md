# Review — ratings T05b (commit ca31898, branch `feat/ratings`)

Reviewer: Courses 2. Date: 2026-10-03. Closes my T05 notes №1 (AC10 archived-list), №2 (DELETE→404
review_not_found), №3 (unit: edit-after-losing-access exists path; admin re-remove calls repo once).
Owner decision recorded in `docs/ratings/assumptions.md`; plan §5 row + Rt30.

## Verdict

**Approve. 0 blocking, 0 non-blocking.** All three focus points PASS, and all three of my T05 notes are
closed exactly as described, each with a test. Clean slice.

Evidence I ran myself: `go build ./...` → **OK**; `go vet ./internal/discussion/...` → **OK**;
`go test ./internal/discussion/...` (application + http + domain) → **ok**. (Integration not run here;
no reaper involved in these packages' unit/contract tests, and ryuk is now disabled in the gate.)

## Focus 1 — the viewability check uses the existing rule, no admin special-casing — PASS

`LearningReviews` (`application.go:308`) builds a `domain.Course` from the review facts and gates on
`course.CanBeViewedBy(actor)` — the **existing** domain method, which wraps the shared
`contracts/access.CanView`:

```
CanView = accessExempt || (Published && (!Paid || Preview))
accessExempt = Administrator || (UserID == OwnerID) || Entitled
```

The admin exemption lives **inside** the one shared access contract (ADR 013), not in a separate
`if actor.Administrator` branch in `LearningReviews`. No special-casing beyond the rule. `Preview` is
left `false` when constructing the Course — correct and deliberate: preview is a per-lesson concept, and
the review **list** is a course-level read, so it must not enter course-level viewability (I checked the
rule: `Preview` only affects `PurchaseRequired`/the lesson lock, never `CanView` for a published-free or
exempt case). No right → `ErrForbidden` → 403 `course_forbidden` (verified `ErrForbidden = "course_forbidden"`
and `statusFor` → 403).

## Focus 2 — archived course with a grant → 200 — PASS

`ReviewFacts` reads the course with no status filter, so facts come back for an archived course. With
`Status=Archived` (`Published=false`) and `Entitled=true`, `accessExempt` is true via `Entitled` →
`CanView` true → the list is returned. Covered at two levels: application `TestLearningReviews/grant
holder can list archived course reviews`, and HTTP `learning list permits an archived course grant
holder`. The refusal path (`Archived, Paid, not entitled, not owner, not admin`) → `ErrForbidden` is
tested too (`without a viewing right is forbidden` / HTTP `refuses a viewer without a right`), plus
anonymous → 401 (`anonymous learning list is refused`).

## Focus 3 — public list unchanged (published only) — PASS

`PublicReviews` (`application.go:302`) still resolves the slug via `PublishedReviewCourse` (published-only)
first, then delegates to the new shared `listReviews`. The extraction of `listReviews` is a pure refactor
(summary + list + rating) shared by both endpoints; `PublicReviews`'s published-only gate and shape are
byte-for-byte the same behaviour as T05. The T05 contract test for the public route stays green.

## My T05 notes — all closed

- **№1 (AC10).** New `GET /api/v1/learn/courses/{courseId}/reviews`, `authenticate`-gated, same
  `publicReviewsDTO` shape, viewable by grant holder / owner / admin incl. archived — exactly the
  owner's decision. The archived-course review list now reaches a grant holder, which the public
  (published-only) endpoint could not. The singular `…/review` (own) and plural `…/reviews` (list) are
  distinct ServeMux patterns — no routing conflict.
- **№2.** `deleteReview` now passes `review_not_found` as the not-found token → DELETE with no own review
  → 404 `review_not_found` (HTTP test `delete without a review is review not found`). (A DELETE against a
  genuinely unknown course also reports `review_not_found` now; fine — the route is review-scoped.)
- **№3.** `TestPutReviewRefusesEditAfterAccessIsLost` sets `facts.Exists = true` → the **edit** path with
  lost access is refused (G4), no longer only the create path. `TestRemoveReviewIsNoopWhenAlreadyRemoved`
  asserts `f.removed == 1` → admin re-remove is a true no-op that hits the repository exactly once.

## Ryuk follow-up (FYI, not part of this review)

The commit disables ryuk in `integration.sh` (82f99df) per my research, and records Codex's read-only
consult (`work/_wip-ryuk-race.md`): a detached context only lowers the #3918 probability and a
same-process retry cannot recover from the orphan, so keeping `TESTCONTAINERS_RYUK_DISABLED=true` is the
right call. Consistent with my findings in `work/research-ryuk.md` (recommendation #1). I did not
re-audit the wrapper consult; the conclusion matches the upstream mechanism.

## Rows (plan §5 / Rt30)

Rt30 (archived course, grant holder, learning list → 200 with reviews; no right → 403) covered at
application + contract. OpenAPI + Orval for the new route regenerated; the contract test validates HTTP
against the spec and `check-generated` is green.
