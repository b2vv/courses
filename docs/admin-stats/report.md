# admin-stats — phase report

| | |
|---|---|
| Date | 2026-10-10 |
| Branch | `feat/admin-reconstruction` |
| Spec / plan / tasks | `spec.md`, `plan.md`, `tasks.md` (GATE 1 ✔, GATE 2 ✔) |
| Closing gate | ordered `make check` EXIT=0 — format/lint/typecheck, unit (frontend 684, backend all), integration, check-generated (0 drift), build, **e2e 114 passed (0 failed, 0 did-not-run)** |
| Commits | T3 `4d3a141` · T4 `54edc3a` · T5 `e30b257` · T6 `f838f93` · T7 `92da56d` · T8 `f7fd1a5` · T9 `d9c0ae4` · T10 `1473fcf` · T11 (this) |

A new admin analytics subsystem: a dedicated `adminstats` backend module
(domain → application → ports → postgres/http adapters) exposing two admin-only
read endpoints, and a `/admin/analytics` screen that renders the platform
snapshot, an instructor leaderboard, and a metric trend.

## 1. What was built

- **Backend module `internal/adminstats`** — `Snapshot(ctx)` aggregates six
  platform domains (users, courses, grants, orders, reports, media) plus a
  per-instructor leaderboard and an engagement block; `Timeseries(...)` buckets
  one of four metrics by day/week over a UTC window. Raw read-only pgx over the
  shared pool (orders/discussion precedent), no sqlc. Range/metric/bucket
  validation is pure domain.
- **HTTP** — `GET /api/v1/admin/stats` and `/admin/stats/timeseries` behind
  `authenticate → RequireAdministrator`, `Cache-Control: no-store`. The
  single-instructor flag (`PLATFORM_SINGLE_INSTRUCTOR`, default true) rides in
  the snapshot payload.
- **Frontend** — a domain view model, a Zod schema, an `AdminStatsRepository`
  port + API adapter, a query boundary, the `StatTile` and `TrendChart` UI
  primitives, and the `/admin/analytics` route + `AnalyticsView` (hero tiles,
  six KPI sections, the instructor leaderboard, a metric-switching trend, and an
  honest placeholder for deferred metrics), with nav entry, icon and i18n.

## 2. Acceptance accounting (plan §6.1, row by row)

Each row is tied to the test that proves it, not to the aggregate gate.

| # | Scenario | Proven by |
|---|---|---|
| 1 | Snapshot as admin | contract `adminstats/adapters/http` `TestSnapshotEndpoint` (200, revenue number, no-store) |
| 2 | Analytics screen numbers | e2e `admin-analytics.spec` (hero tile value `/\d/`, sections render) |
| 3 | Trend signups sum = count | repo `TestAdminStatsTimeseriesAgainstPostgreSQL/SignupsSumEqualsCountInWindow` + e2e trend renders |
| 4 | Revenue by paid_at | repo `.../RevenueBucketsByPaidAtNotCreatedAt` |
| 5 | Empty platform = zeros | repo `TestAdminStatsSnapshot.../EmptyPlatformIsAllZeros` + contract `TestSnapshotEmptyLeaderboardIsArray` |
| 6 | Empty trend window = `[]` | repo `.../EmptyWindowIsNonNilEmpty` + contract `TestTimeseriesEmptyIsArray` + unit `TrendChart` empty state |
| 7 | Max window (366 buckets) | domain unit `TestResolveRangeCapBoundary` (366 passes, 367 `range_too_wide`) |
| 8 | Video without duration | repo (null avg over non-null only) + contract (nullable) + view renders `—` (unit + e2e) |
| 9a | single-instructor ON hides | e2e (leaderboard + instructor heading absent, payload `singleInstructor=true`) + unit `AnalyticsView` hide + contract payload |
| 9b | single-instructor OFF ordered | repo `TestAdminStatsInstructorsLeaderboard` (A>B>C by revenue) + unit `AnalyticsView` renders rows in order |
| 9c | Instructor without sales = 0 | repo leaderboard test (the no-sales instructor stays at 0, not dropped) |
| 10 | Non-admin refusal | contract `TestSnapshotAuthorization` (guest 401, student/instructor 403) + e2e instructor (not-found + API 403 `grant_forbidden`) |
| 11 | Unknown metric | contract `TestTimeseriesRangeErrors/unknown metric` + domain `TestParseMetric`/`TestParseBucket` |
| 12 | Bad range | contract (inverted, too wide) + domain `TestResolveRangeFromAfterTo` / `TestResolveRangeCapBoundary` |
| 12b | Unparseable date | contract `TestTimeseriesRangeErrors/unparseable from` + domain `TestResolveRangeParseFailure` |
| 13 | Honest degradation | e2e (deferred-metrics placeholder visible, no fabricated number) + unit `AnalyticsView` |
| 14 | Consistency with list | e2e `admin-analytics.spec` (`stats.active + blocked == users.list total`, via the API) |

All 17 rows carry row-level proof. Manual rows: 0.

## 3. Decisions taken in implementation (each closes an alternative)

- **2026-10-10 — admin gate lives only in `RequireAdministrator` middleware.**
  The application service takes no actor and does not re-check admin, so there is
  no unreachable defensive branch behind the gate. The observable 401/403
  behaviour is unchanged and is contract-tested (`TestSnapshotAuthorization`) and
  e2e-tested. Closes plan §1's "service re-checks admin via actor".
- **2026-10-10 (owner) — engagement is implemented, not placeheld.** All three
  metrics (active learners 30d, average completed-lesson fraction, completions)
  ship with the semantics stated in the T7 commit. Closes the "degrade to
  placeholder" alternative the plan left open.
- **2026-10-10 — 9a is proven live (e2e) on the default single-instructor
  stack; 9b's rendering is proven by a component unit test and its data ordering
  by the repository test.** One e2e stack carries one `PLATFORM_SINGLE_INSTRUCTOR`
  value, so both states cannot be exercised live in one run; the default (true)
  is the shipped config and is the one verified end-to-end. Changing the shared
  seed to ≥2 instructors purely to flip the flag was judged a larger risk than
  the split. Closes "both 9a and 9b via live e2e on one stack" (physically
  impossible with one env).
- **2026-10-10 — no engagement index/migration.** The engagement query's
  grant×lesson intermediate set is small at this platform's scale and the
  existing primary keys serve it; documented in `engagement.go` to revisit if
  `lesson_progress` grows. Mechanism recorded, not a bare hypothesis.

## 4. Verification

Ordered `make check` EXIT=0 on `feat/admin-reconstruction`:
format-check, lint, typecheck; unit (frontend **684**, backend all packages);
integration (adminstats repo tests on isolated Postgres among them);
check-generated (0 drift after the OpenAPI + engagement regen); build (all
images); e2e **114 passed, 0 failed, 0 did-not-run** (12.5 min), including the
two `admin-analytics.spec` cases.

Two defects were found and fixed inside the phase, not left for the gate to
mask: T1 had not added the four range/metric error codes nor the engagement
schema to the OpenAPI `ErrorResponse`/snapshot (fixed in T6/T7 with regen); and
the analytics grid forced horizontal overflow at 200% text zoom (`minmax(13rem)`
→ `minmax(min(13rem,100%),1fr)`), caught by the e2e zoom assertion on the first
full run.

## 5. Limitations

- **Deferred metrics.** Watch-through, tags and course popularity are out of
  scope (owner decision during GATE 1); the analytics page shows them as an
  honest placeholder (scenario 13), not fabricated numbers. Design notes:
  `docs/design/admin-analytics-tags-popularity.md`.
- **9b not exercised on a live multi-instructor stack.** Its data ordering is a
  repository test and its rendering a component test (see §3); the live e2e runs
  the default single-instructor config.
- **No engagement index.** See §3 — fine at current scale; revisit if
  `lesson_progress` grows large.
- **Trend window is day/week only**, capped at 366 buckets (≈1 year of days,
  ≈7 years of weeks); wider windows are refused `range_too_wide`.
- **`Delta` primitive** carries a latent axe issue (`aria-label` on a bare span);
  it is not used on the analytics page (the snapshot has no period-over-period
  comparison) and was left untouched — its own test depends on the current
  behaviour.

## 6. Next

Phase closed. The roadmap (`work/PRODUCT-BRIEF.md` §24) remains complete;
admin-stats was an owner-initiated mini-phase via spec-flow. Production
deployment stays a deferred human decision (test mode, no prod server).
