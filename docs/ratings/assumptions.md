# Ratings phase — assumptions and decisions

Append-only by meaning: a correction is a new entry referencing the earlier one.

## 2026-10-03 — GATE 1 passed (owner, `grilling`)

**Decision (owner, 2026-10-03):** the ratings spec is agreed. Ten decisions G1–G10 were taken in the
gate, each with the code fact it rests on; they live in `spec.md` §«GATE 1 decisions», and the
acceptance criteria AC1–AC12 follow from them. **Alternatives closed:** listed per decision in that
table (e.g. preview counts as a right, owner moderates reviews, pre-moderation, optional text).
**Scope it adds beyond reviews:** G5 changes `GrantCourse` in the closed payment module (no
self-grant); G2 shows the buy path to an admin. **How to check:** each AC maps to an acceptance row
in `plan.md` and to a test; GATE 2 refuses a plan that leaves an AC unmapped.

## 2026-10-03 — No course duration on the card in this phase (Courses 1, GATE 2 critique finding 2)

**Decision (Courses 1, 2026-10-03, plan §5):** the catalog `Course` gains `rating` and `lessonCount`
but no duration. **Alternative closed:** a nullable total duration. **Mechanism:** a true total
needs the current ready asset of every lesson — a per-lesson `LATERAL` (as in
`discussion/adapters/postgres/repository.go:28-34`) — inside `ListPublishedCourses`, the hottest
public read, and any lesson without a ready video makes the total wrong; AC11 already makes the
duration conditional on real data. **How to check:** the OpenAPI `Course` schema has no duration
field; reopen when a duration is wanted and measure the query with the `LATERAL` under §9's method.

## 2026-10-03 — R6 trigger made absolute: catalog page median ≤ 5 ms (owner; supersedes the ">2×" trigger)

**Decision (owner, 2026-10-03):** the R6 revisit trigger is an absolute budget — the production
`ListPublishedCourses` page (defaults: empty search, sort `recent`, page 20) under the plan §9 method
(100 courses, 20 000 reviews Zipf top 2 000, `VACUUM ANALYZE`, median of 5) must stay **≤ 5 ms**; over it,
the materialized view is the next step. **Alternative closed:** the ratio trigger ">2× the same query
without the aggregate", and building the materialized view now. **Why (magnitude):** measured with the
harness `catalog_measure_test.go` (tag `measure`): 1.306 ms with the aggregate vs 16 µs without, 81.6×.
The baseline is a 2-buffer index scan, so any join exceeds 2× — the ratio fires for every design and
measures nothing; 1.3 ms is within the budget. Before `VACUUM ANALYZE` the same run gave 2.63 ms / 175×
(heap fetches on an empty visibility map, guessed row counts) — not the steady state. **How to check:**
run the harness; it prints the median and fails when it exceeds 5 ms.
