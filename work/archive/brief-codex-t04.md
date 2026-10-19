# Codex brief — ratings T04: catalog rating summary, lessonCount, canPurchase, R6 measurement

Repository /Users/strelia/projects/courses, branch `feat/ratings`. Read first (they bind): `AGENTS.md`,
`docs/ratings/spec.md` (AC3, AC8, AC10, AC11, AC12, G2, G3, R6), `docs/ratings/plan.md` (§1 placement
rows for the catalog, the threshold and the admin buy entry; §5 rows «catalog `Course`» and «learning
course»; §7; §9 R6 measurement; §11 rows Rt07, Rt09), `docs/ratings/assumptions.md` (no duration),
`docs/adr/013-one-access-rule-as-a-contract.md`, `docs/adr/014-payment-facts-in-the-access-contract.md`.
In the tree already: `backend/internal/contracts/rating` (`Summary`) and migration
`backend/migrations/00013_ratings.sql` (`discussion_reviews`). Log each step to `work/codex-t04.log`
(first line before the first tool: `NEXT:`). Do **not** commit.

Scope: the `course` module only, end to end (SQL → domain → HTTP → OpenAPI → Orval), plus the
measurement. **Do not touch the `discussion` module.** The `course` module reads the reviews table in
its own SQL, exactly as modules read `course_grants` today (ADR 013 rule B) — no Go import of
`discussion`. Write tests first.

## 1. Catalog summary and lesson count

- `backend/sql/queries/courses.sql`: `ListPublishedCourses` and `GetPublishedCourse` also return, per
  course, `review_count` and `review_stars_sum` over `discussion_reviews` rows with
  `removed_at IS NULL`, and `lesson_count` (all lessons of the course — a lesson has no publish state).
  Use aggregate subqueries / `LEFT JOIN LATERAL` so a course with no reviews or no lessons gets 0, and so
  the `GetPublishedCourse` row-per-lesson shape does not multiply the counts. Keep the existing search,
  sort and paging semantics unchanged. Regenerate sqlc (`make generate`); generated types stay in the
  adapter and are mapped explicitly.
- Domain: the catalog course gains `Rating *{Average float64; Count int}` (nil when not visible) and
  `LessonCount int`. Visibility and rounding come **only** from `rating.Summary(count, sum)`; no
  threshold or rounding logic in `course` itself.
- HTTP + OpenAPI `Course` schema (`backend/openapi/openapi.yaml` ~line 1632) and the course-page
  response: add `rating` (nullable object `{average: number, count: integer}`) and `lessonCount`
  (integer). Required + nullable per the file's style (`additionalProperties: false` is in force). **No
  duration field** (assumptions 2026-10-03). Regenerate Orval; `check-generated` must stay clean.
- Tests: unit for the mapping (count 2 → `rating: null`; 3 reviews 4,5,5 → `{4.7, 3}`); HTTP contract
  test against OpenAPI for both responses; **integration** (course repository tests, real Postgres):
  - **Rt07** course with 3 counted reviews (4, 5, 5) → count 3, sum 14, and the mapped `{4.7, 3}`;
  - **Rt09** add a removed review → still count 3; remove one of the three → count 2 → `rating: null`;
  - a course with 0 reviews and a course with lessons in two modules → `lessonCount` right and
    `GetPublishedCourse` counts not multiplied by the lesson rows;
  - an archived course is absent from the catalog as today (Rt26's catalog side: re-publishing shows the
    rating again).
  Insert review rows with plain SQL in the fixture (the discussion repository is not importable here).
  Give new subtests their own ids and scope counts to your own courses.

## 2. `canPurchase` on the learning course (G2, AC3)

- The learning course response (`LearningCourse` schema ~line 1608; `course/adapters/http/learning_handler.go`;
  domain `course/domain/learning.go`) gains `canPurchase: boolean` = published ∧ paid ∧ **no active
  grant** ∧ **actor is not the owner**. The administrator role is **ignored** here on purpose: an admin
  without a grant gets `canPurchase: true` (today `PurchaseRequired` is false for them, so they never see
  a buy entry). Put it next to the existing facts; it is a small pure function in `course/domain`, unit
  tested: admin without grant → true; admin with grant → false; owner → false; student with grant → false;
  student without grant on a paid course → true; free course → false; draft → false.
- OpenAPI + Orval + contract test.

## 3. R6 measurement (plan §9) — a separate, opt-in test

- A Go test behind its own build tag `measure` (so `make test`/`make integration` never run it), e.g.
  `backend/internal/course/adapters/postgres/catalog_measure_test.go`: starts Postgres via testcontainers
  like the integration tests, applies migrations, seeds **100 published courses** with 3 lessons each and
  **20 000 reviews** spread Zipf-like (top course 2 000; deterministic seed), then runs
  `EXPLAIN (ANALYZE, BUFFERS, FORMAT TEXT)` on the catalog list page (the catalog's default page size and
  sort) **with** the new aggregate and **without** it (the previous query text, kept in the test), 5 runs
  each after one warm-up, and prints both medians, the ratio, and one full plan of each.
- Document the exact command in the file header. You will likely not be able to run it (Docker); say so.

## Verify
`DOCKER_CONFIG=/tmp/claude-501/dockercfg make generate format-check lint typecheck test integration
check-generated`. If Docker is denied in your sandbox, say so — Courses 1 runs the gate and the
measurement; still run `go build ./...`, `go vet ./...` and the non-Docker tests, and the frontend
`typecheck` after Orval. Never claim a gate you did not run.

## Channels
If a contract above is missing something you need, stop and say what. If a rule here contradicts the
spec, the plan or the code in a way that would produce a bug, argue it with the code path instead of
silently diverging. Final answer: changed files, tests added (names), what you ran with results, what
you could not.
