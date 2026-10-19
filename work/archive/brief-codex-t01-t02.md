# Codex brief — ratings T01 + T02: rating contract, review domain rules, no admin self-grant

Repository /Users/strelia/projects/courses, branch `feat/ratings`. Read first (they bind): `AGENTS.md`,
`docs/ratings/spec.md` (AC1, AC2, AC4, AC5, decisions G1, G2, G5, G6, G7), `docs/ratings/plan.md`
(§1 placement, §3 progress semantics, §4 eligibility, §10 constitution, §11 rows Rt06–Rt08, Rt10, Rt11,
Rt15, Rt18, Rt24), `docs/adr/013-one-access-rule-as-a-contract.md`. Log each step to
`work/codex-t01-t02.log` (first line before the first tool: `NEXT:`). Do **not** commit.

Scope is domain and the one orders guard. No migration, no repository, no HTTP for reviews yet
(those are T03/T05). Write the tests first.

## T01-a — `backend/internal/contracts/rating` (new package)

Pure standard library; imports no module (the existing architecture test forbids it — keep it green).

```go
const MinimumVisibleCount = 3
// Summary decides whether a rating is shown and the rounded average.
// count = counted reviews, starsSum = sum of their stars (each 1..5).
func Summary(count int, starsSum int) (average float64, visible bool)
```
- `visible` is `count >= 3`. When not visible, `average` is 0.
- `average` = `starsSum / count` rounded half-up to **one** decimal (4, 5, 5 → 4.7; 5, 5, 4, 4 → 4.5;
  3 × 5 → 5.0). Use integer arithmetic to round (no float drift): e.g. `(starsSum*20 + count) / (2*count)`
  tenths, then `/10`. Prove the rounding with a table including x.x5 cases.
- Invalid input (`count < 0`, `starsSum < count`, `starsSum > 5*count`) returns `(0, false)`.
- Unit table in `rating_test.go`: counts 0, 1, 2, 3, large; the rounding cases above.

## T01-b — review rules in `backend/internal/discussion/domain` (new file `review.go`)

Follow the file style of `comment.go` and `access.go` (same `Actor`, same error style in `errors.go`).

1. **Errors** (add to `errors.go`, `errors.New` like the others):
   `review_owner_forbidden`, `review_not_available`, `review_purchase_required`, `review_not_started`,
   `review_invalid`.
2. **Facts and eligibility:**
   ```go
   type ReviewFacts struct {
       OwnerID       string
       Status        Status
       Paid          bool
       Entitled      bool // active grant for this actor, any source
       HasProgress   bool // at least one lesson_progress row for this actor on this course
   }
   func CanReview(facts ReviewFacts, actor Actor) error
   ```
   Checks **in this order**, returning the first that fails, `nil` if all pass:
   1. `actor.UserID == ""` → `ErrForbidden` (existing).
   2. `actor.UserID == facts.OwnerID` → `review_owner_forbidden` (G2: owner never, even with a grant).
   3. not (`Status == Published` or `Entitled`) → `review_not_available` (draft; archived without a grant).
   4. `Paid && !Entitled` → `review_purchase_required` (G1: preview gives no right; G2: admin role access
      gives no right).
   5. `!HasProgress` → `review_not_started` (R1).

   **`actor.Administrator` must not appear in `CanReview` at all** — role does not decide (G2). Do NOT
   route through `access.CanView`: it exempts admins and owners, which is exactly what must not count.
   Unit table must include: owner with grant and progress (refused); admin, paid, no grant, progress
   (purchase_required); admin, free, progress (allowed); admin, paid, grant, progress (allowed);
   preview-only watcher = paid, no grant, progress (purchase_required); archived + grant + progress
   (allowed); archived free without grant (not_available); draft (not_available); free, no progress
   (not_started); anonymous (forbidden).
3. **Format:**
   ```go
   const (MinimumReviewRunes = 20; MaximumReviewRunes = 2000)
   func NewReviewContent(stars int, body string) (ReviewContent, error) // review_invalid
   ```
   Stars integer 1..5. Body: `strings.TrimSpace`, must be valid UTF-8, rune count after trim in
   [20, 2000] (same rune counting as `comment.go:40`). Store the trimmed body. Table: 19/20/2000/2001
   runes after trim, a body that is 20 runes only because of surrounding spaces (refused), stars
   0/1/5/6, invalid UTF-8.
4. **Progress capture (R2, G6):**
   ```go
   type ReviewProgress struct { Completed, Total int }
   func NewReviewProgress(completed, total int) (ReviewProgress, error) // review_invalid if completed<0 || total<completed
   func (p ReviewProgress) Percent() int // round half-up of Completed*100/Total; 0 when Total == 0
   ```
   Table: 0/4 → 0, 1/3 → 33, 2/3 → 67, 1/8 → 13 (12.5 rounds up), 4/4 → 100, 0/0 → 0.

## T02 — no admin self-grant (`backend/internal/orders`)

- New domain error `ErrGrantSelfForbidden Error = "grant_self_forbidden"` in `orders/domain/errors.go`.
- `Service.GrantCourse` (`orders/application/service.go:131`): after the administrator check, if
  `userID == actor.UserID` return `ErrGrantSelfForbidden` before validating the reason or touching the
  store — nothing is written.
- HTTP: map it to **403** in the orders handler next to `ErrGrantForbidden`; add the code wherever the
  OpenAPI spec lists the admin-grant error codes (`backend/openapi/openapi.yaml`), regenerate if the spec
  changes (`make generate`), and keep `check-generated` clean.
- Tests: unit (service with a fake store: self-grant refused, store untouched; grant to another user
  still works), HTTP contract test (403 + code), and one integration subtest asserting no grant row is
  created for a self-grant. In `repository_integration_test.go` give new subtests their own id source
  and scope counts to your own rows (the shared id slice there has starved other subtests before).
  Users in fixtures use the existing Argon2id fixture hash.

## Verify
`DOCKER_CONFIG=/tmp/claude-501/dockercfg make format-check lint typecheck test` (and `integration` +
`check-generated` if Docker is available to you). If Docker is denied in your sandbox, say so —
Courses 1 runs the gate. Never claim a gate you did not run.

## Channels
If a contract above is missing something you need, stop and say what. If a rule here contradicts the
spec or the code in a way that would produce a bug, argue it with the code path instead of silently
diverging. Final answer: changed files, tests added (names), what you ran with results, what you could not.
