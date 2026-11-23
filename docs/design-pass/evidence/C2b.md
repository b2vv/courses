# C2b — a course with no lessons leaves the catalog after 30 days

| | |
|---|---|
| Date | 2026-11-23 |
| Branch | `design-pass-c2`, commit `37bb1009` on top of C2 (`699b4032`) |
| Contract | `plan.md` §15 (C2b-1…C2b-5); owner answer to C2-audit #6 |
| Build checked | API image `702c3e9a…` built from `37bb1009` on Kiri |

## Result against acceptance

| Item | Proof | Status |
|---|---|---|
| C2b-1 | Integration `C2b an empty course leaves the list…`: an empty course published 1 s before the window is absent from the list and the total; one published exactly at the window edge is listed. Live: before reseeding, `GET /api/v1/courses` returned `total 0` (the old seed courses have no lessons and were published in September). | PASS |
| C2b-2 | Same test: a course with a lesson published in February is listed; an empty course created in 2025 but published inside the window is listed (the count starts at publication). | PASS |
| C2b-3 | Same test: `total` equals the list length in both the windowed and the unwindowed case; one predicate in `ListPublishedCourses` and `CountPublishedCourses`. | PASS |
| C2b-4 | After `make seed`, total 3, published 1/2/3 days before the seed run, «спочатку нові» order unchanged. The migration-preservation tests that run the seed on a downgraded schema pass. | PASS |
| C2b-5 | e2e `catalog.spec.ts` «a course with no lessons leaves the catalog 30 days after publication (C2b)»: at 31 days, API `total 0` and «Курсів не знайдено», the course page answers 200; at 29 days, one card with «Скоро буде» and «1 курс». The fixture is removed afterwards. | PASS |

The application unit test pins the window: with the clock at 2026-11-23 12:00 UTC the list gets `EmptyListedSince` 2026-10-24 12:00 UTC.

## Defect found while writing the test

The repository re-validated the query with `domain.NewQuery`, which dropped the new field, so the window never reached SQL. The integration test failed (4 courses listed instead of 3); the repository now carries the window through validation.

## Tests

- `go test ./internal/course/...` and `go vet` pass; integration `./internal/course/adapters/postgres` passes (Kiri, testcontainers).
- e2e `catalog.spec.ts`: 6 passed on the live stack.
- Ordered `make check` on `1f94eabc` (C2 + C2b together), Kiri, 06:32–07:04 UTC: **GATE_EXIT=0** — frontend unit 1508 passed, 0 failed; e2e 170 passed.

## Independent review

Codex on `37bb1009`: **PASS**, no findings (`docs/design-pass/_wip-review-c2b.md`). It found the list predicate used only by the catalog; the admin lists, analytics, enrollment and the course page are unchanged.

## Limitation

An environment that is not reseeded keeps the September seed dates, so its catalog is empty until `make seed` runs again (the rule working as intended on data without lessons). `make check` reseeds.
