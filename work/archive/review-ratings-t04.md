# Review — ratings T04 (commit f598d0e, branch `feat/ratings`)

Reviewer: Courses 2. Date: 2026-10-03. Against: `docs/ratings/spec.md` (AC3, AC8, AC11, AC12),
`docs/ratings/plan.md` §1/§5/§9, briefs `work/brief-codex-t04.md` + `work/brief-codex-t04-followup.md`.
Scope: catalog/course/learning rating summary + lessonCount + canPurchase; OpenAPI/Orval; R6 harness.

## Verdict

**Approve. 0 blocking, 4 non-blocking.** All five focus points Courses 1 named PASS. The self-fixed
test defects are not weaker — one is stronger. R6's 81.6× ratio and the materialized-view decision are
the owner's, out of this review per Courses 1; I checked only that the harness measures honestly.

Evidence I ran myself (the whole tree builds now that T04 is committed and sqlc is regenerated):
- `go build ./...` → **BUILD_OK**; `go vet ./internal/course/...` → **VET_OK**.
- `go test ./internal/course/domain/… ./internal/course/adapters/http/… ./internal/course/adapters/postgres/ (non-integration) ./internal/contracts/rating/…` → **all ok** (CanPurchaseBy table, learning HTTP contract, mapCourse mapping, rating rounding).
- I did **not** run the Docker integration suite (Courses 1 reports exit 0); I verified the SQL and the
  subtest logic/counts statically.

## Focus 1 — GetPublishedCourse row-per-lesson shape does NOT multiply counts — PASS

`review_count`, `review_stars_sum`, `lesson_count` each come from an independent `LEFT JOIN LATERAL`
(aggregate, no `GROUP BY`, correlated only on `courses.id`) → each returns exactly one row, so the value
is a **constant repeated on every lesson row**, not a per-row sum. The outer `course_modules ⋈
course_lessons` fan-out never reaches the aggregates. `GetPublished` reads the scalars from `rows[0]`
(`repository.go:66-69`) and never accumulates them across the outline rows. Rt07 proves it at the DB:
3 reviews (4,5,5) on a 3-lesson course → `Rating.Count == 3` (**not** 3×3 = 9), `Average == 4.7`,
`LessonCount == 3`. `mapCourse` delegates visibility+rounding entirely to `contracts/rating.Summary`
and guards against negative/overflow counts → `ErrDependencyUnavailable`.

## Focus 2 — canPurchase ignores admin role (G2) and refuses owner — PASS

`LearningCourse.CanPurchaseBy = Status==Published && Paid && !Entitled && actor.UserID != InstructorID`
(`learning.go:107`). It **never references `actor.Role`/`Administrator`** — role is not consulted (G2) —
and refuses the owner by id. Matches plan §5 exactly. Unit table covers all seven cases: admin without
grant → **true** (admin may buy, G2/AC3), admin with grant → false, owner → false, student with grant →
false, student without grant (paid) → true, free → false, draft → false. Wired into
`learningCourseDTO.CanPurchase` via `course.CanPurchaseBy(actor)` (`learning_handler.go:106`); the HTTP
contract test asserts the free fixture → `canPurchase:false` and `lessonCount == len(lessons)`. OpenAPI
documents it: *"Administrators are deliberately not exempt."*

## Focus 3 — the Cleanups really restore the seed — PASS (verified against the primary source)

**Rt07/Rt09 (`repository_integration_test.go`):** the subtest mutates the shared seeded course
`10000000-…-0001` (status, published_at) and adds 4 reviews + 1 user; `t.Cleanup` deletes the 4 reviews
and the user and runs `UPDATE courses SET status='published', published_at='2026-09-03 10:00:00+00'`.
I checked the restored value against **`sql/seed.sql:5`**, not Courses 1's word: the seed's `published_at`
for that course is `'2026-09-03T10:00:00Z'` — identical (Z == +00). `updated_at` is never touched by the
subtest, so it is not corrupted. `t.Cleanup` fires at the end of this subtest, before its siblings, so
the later subtests and the `page.Items[0] == "go-backend-foundations"` ordering (line 525, which runs
*before* Rt07 anyway) see the seed intact.

**Learning parity (`learning_integration_test.go`):** the subtest adds a 2nd module + lesson (→
`LessonCount` 5) and reviews on two courses; `t.Cleanup` deletes reviews → lesson → module → users in
**FK-safe order**, returning the course to its original 4 lessons — which is what B27 counts. Complete
and correctly ordered (reviews reference users, lesson references module).

## Focus 4 — no assertion weakened — PASS

The commit's listed fixes preserve or strengthen the rows:
- **Rt07 "removed review inserted as active"** → now inserted with `removed_at` set, so Rt09 genuinely
  tests that a removed review is not counted (`Rating.Count` stays 3, then drops to null below threshold).
  This **strengthens** the test (the old insert couldn't exercise Rt09 at all).
- **invalid password hash** → the fixtures now use a well-formed Argon2id hash (gate green).
- **multi-command parameterised Exec** → split to one statement per `Exec` (SQLSTATE 42601). Same data,
  valid execution.
- **before/after + the two Cleanups** → Focus 3 above; the parity subtest additionally asserts
  `*learning.Rating == *catalog.Rating` and equal `LessonCount`, which is new coverage (the follow-up's
  point: the learning response now carries real rating/lessonCount, not placeholders).

## Focus 5 — measurement harness measures the production texts — PASS

`catalog_measure_test.go` (build tag `measure`, opt-in, not in the gate): `withAggregate` is the
`ListPublishedCourses` SQL with both `LEFT JOIN LATERAL`s **and selects `review_count /
review_stars_sum / lesson_count`** — so an unreferenced-LATERAL planner drop cannot silently measure
nothing (the exact bug that was fixed). `VACUUM ANALYZE` runs before measuring (steady state, justified
in-comment). The seed matches plan §9 (100 courses, 20 000 reviews, Zipf top=2 000, page 20, recent,
median of 5 `EXPLAIN ANALYZE` runs). The hard assertion is the owner's 5 ms budget
(`assumptions.md`); the 81.6× ratio is logged for the R6 decision, which is pending with the owner and
outside this review.

## OpenAPI / Orval — consistent

`Course`/`CourseDetail`/`LearningCourse` gain `rating` (object, `nullable:true`, required `[average,
count]`, `average` number/double 1–5) and `lessonCount` (integer ≥0); `LearningCourse` gains
`canPurchase` (boolean). `rating` being both `required` and `nullable` is correct — the key is always
present and the Go `*ratingDTO json:"rating"` serializes `null` below threshold. `check-generated` ran
green, so the regenerated sqlc (`courses.sql.go`, 96 lines) and Orval models match source.

## Non-blocking

1. **`average` as a JSON number.** A whole average (5.0) is emitted by Go as `5`. Valid under the
   `number` schema, but the card/course page (T06) must format to one decimal (`★ 5.0`). This is the
   T01 forward note, now concrete because the field exists. Add a one-decimal format + a test in T06; no
   backend change.
2. **Harness SQL is a hand-transcribed copy** of the generated `ListPublishedCourses`. If that query
   later changes, `withAggregate` can drift silently (it is opt-in and only informational, so low risk).
   A short comment pointing at the source line, or deriving the text, would keep them honest.
3. **Uncommitted working-tree changes** (not part of f598d0e): the 5 ms budget assertion in
   `catalog_measure_test.go`, and `docs/ratings/{spec,plan,assumptions}.md`. Looks like in-flight
   follow-up — commit it so the budget gate and the docs travel with the slice.
4. **`rating.average` float equality in tests.** `repository_mapping_test`/integration compare
   `Average == 4.7`; this works because `Summary` divides an integer tenths value by 10 identically to
   the literal, but it is float equality. Fine as-is; just flagging the pattern for future averages.

## Rows covered (plan §12)

Rt07 (threshold reached, `★ 4.7 (3)`), Rt09 (removed not counted → below threshold → null), the
catalog↔learning rating/lessonCount parity, and AC3's `canPurchase` (unit + HTTP). R6 measurement
harness present and honest; the >2× decision is the owner's.
