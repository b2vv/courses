# Course ratings and reviews — plan

Status: **GATE 2 passed 2026-10-03** — critique green (0 blocking, 6 non-blocking closed), defense green after one re-check (Q2 closed), constitution: one violation fixed (threshold → `contracts/rating`). Reviewers: two fresh `gate-reviewer` subagents (Opus 4.8); trace in `work/_wip-ratings-gate2-{critique,defense}.md`. Spec: [spec.md](spec.md) (GATE 1 passed, AC1–AC12,
G1–G10). Branch `feat/ratings`. Threshold (`.claude/standards.md` §Поріг): **met** — new subsystem,
public API change (catalog `Course`), both backend and frontend layers.

Roles as in the payment phase: Courses 1 plans, gates and commits; code by Codex from briefs;
review by Courses 2. Test mode only — no production server (owner, 2026-10-03).

## 1. Placement

| Concern | Where | Why there |
|---|---|---|
| Reviews: write, edit, delete, list, eligibility | `backend/internal/discussion/{domain,application,ports,adapters}` | `discussion` already owns user-written course content, the report queue, the moderation log and the instructor workspace (I01–I03). R3 requires reusing them. A new module would copy the whole moderation path. |
| Rating summary and `lessonCount` on the catalog card | `backend/sql/queries/courses.sql` (`ListPublishedCourses`, `GetPublishedCourse`) — `course` module | The catalog query belongs to `course`. It reads the reviews table in its own SQL, the same way four modules already read `course_grants` (ADR 013 rule B, ADR 014; e.g. `discussion/adapters/postgres/repository.go:20-27`). No package import, no port. |
| Display threshold R4 (≥ 3) and the rounded average | `backend/internal/contracts/rating` — one pure function `Summary(count, starsSum) (average, visible)` | A business rule, not presentation: the API returns `rating: null` below 3, so no client can show a premature average. Two modules need it (`course` for the card, `discussion` for the course page), and AGENTS puts a cross-module rule in `internal/contracts/…` as a domain-safe pure function; copies pinned by a behaviour test are option C that ADR 013 rejected. Each module's SQL returns only `count` and `sum(stars)`; the contract decides visibility and rounding. |
| No self-grant (G5) | `backend/internal/orders/application/service.go` `GrantCourse` | The rule is about grants; `orders` owns them. |
| Buy entry for an admin (G2) | `course` learning response + `frontend/src/presentation/learn` | Today the only buy entry is a locked lesson (`LockedLessonPanel.tsx`), never shown to an admin because `PurchaseRequired` is false for them (`access.go:26-31`). |
| Frontend | `frontend/src/{domain,application,adapters,presentation,app,routes}` per manifest | — |

Not in `internal/contracts`: the eligibility rule needs facts no other module decides on, and
ADR 013 keeps `contracts` for the shared access predicate only.

## 2. Reuse first

| Reused | From | How |
|---|---|---|
| Report → queue → dismiss/remove → audit flow | `discussion` I01–I03 (`discussion_comment_reports`, `discussion_moderation_log`, teach routes, `teach-queries.ts`, `routes/teach.courses.$courseId.tsx`) | Reviews get a sibling reports table and share the log; the teach queue lists both kinds. |
| Facts query with grant join | `discussion/adapters/postgres/repository.go:20-34` (`lessonFactsQuery`) | The review-facts query follows it: course owner, status, paid, active grant, plus progress counts. |
| Body validation style | `discussion/domain/comment.go:11,40` (rune count, `utf8.ValidString`) | Same check with 20–2000. |
| Author display name | comment queries join `users.display_name` (`repository.go:38`) | Same join for reviews. |
| HTTP envelope, error codes, request id | `backend/platform/httpserver/` | — |
| API transport, query boundary, UI primitives | `frontend/src/adapters/api/transport.ts`, `app/composition/*-queries.ts`, `presentation/ui/` | — |
| nginx `zone=api` limit and body limit | `infrastructure/nginx` | New write routes sit under `/api/v1/learn/…` and `/api/v1/teach/…`, already covered. |

**New, because nothing existing fits:** the reviews table and its reports table (reports are
`comment_id NOT NULL` — `00010`:13 — so they cannot point at a review); the review-facts query
(progress counts are new); the public reviews list; the catalog aggregate; the `canPurchase` fact.

## 3. Data model — migration `00013_ratings.sql`

```
discussion_reviews (
  id uuid PK,
  course_id uuid NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
  author_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  stars smallint NOT NULL CHECK (stars BETWEEN 1 AND 5),
  body text NOT NULL CHECK (char_length(btrim(body)) BETWEEN 20 AND 2000),
  completed_lessons integer NOT NULL CHECK (completed_lessons >= 0),
  total_lessons integer NOT NULL CHECK (total_lessons >= completed_lessons),
  created_at timestamptz NOT NULL, edited_at timestamptz NULL,
  removed_at timestamptz NULL, removed_by uuid NULL REFERENCES users(id),
  CHECK (removed_by IS NULL OR removed_at IS NOT NULL),
  UNIQUE (course_id, author_id)
)
INDEX discussion_reviews_course_newest ON (course_id, created_at DESC, id DESC)
INDEX discussion_reviews_counted ON (course_id) INCLUDE (stars) WHERE removed_at IS NULL

discussion_review_reports (same shape as discussion_comment_reports, review_id FK ON DELETE CASCADE,
  UNIQUE (review_id, reporter_id)); partial index on open reports

discussion_moderation_log: comment_id DROP NOT NULL; ADD review_id uuid NULL;
  CHECK (num_nonnulls(comment_id, review_id) = 1)
```

Semantics that the shape carries:
- **Author delete (G10)** is a hard `DELETE`: the row and its reports go, the unique slot frees, the
  author may write again. **Admin removal (AC7)** sets `removed_at/removed_by`: the row keeps the
  unique slot, so the author cannot write a replacement, and delete of a removed review is refused.
- **Progress (R2, G6)** is stored as two integers captured at write/edit, not a percentage, so the
  display rounds once and the fact stays exact. `total_lessons` counts all lessons of the course:
  a lesson has no publish state of its own (`00003_authoring.sql:35`); spec AC5's "published
  lessons" means the lessons of the published course. `completed_lessons` counts this author's
  `lesson_progress` rows with `completed_at` set (≥ 95 %, `playback/domain/progress.go:74`).
- **Counted review** (R4, AC8) = a row with `removed_at IS NULL`; a deleted one no longer exists.
- **A lesson without a ready video can never be completed** (`completed_at` needs a duration,
  `progress.go:73-74`), yet it is in `total_lessons`. So a course with such a lesson caps every
  review below 100 %. This is deliberate: the percentage states how much of the course as published
  the author had done, and an uncompletable lesson is part of that course. Named as a limitation in
  the report; Rt12/Rt13 use a course whose lessons all have ready videos.

## 4. Eligibility — one domain function

`discussion/domain`: `CanReview(facts ReviewFacts, actor) error`, facts read in one query:
owner id, published, paid (`price_minor IS NOT NULL`), active grant for the actor, lessons with
progress for the actor, completed lessons, total lessons. Order of checks and codes:

1. actor is the owner → `review_owner_forbidden` (G2)
2. course not visible to the actor through the student path → `review_not_available`
3. paid and no active grant → `review_purchase_required` (G1, G2: role access and preview do not count)
4. no lesson with progress → `review_not_started` (R1)

Writes: create is `INSERT … ON CONFLICT (course_id, author_id) DO NOTHING`; zero rows inserted means a
row already exists, and a follow-up read maps it to `review_removed` when it is removed or
`review_already_exists` otherwise — so two concurrent creates give one 201 and one 409, never a raw
constraint error. Codes: create (`review_already_exists` if a row exists, `review_removed` if it is removed),
edit (same eligibility; `review_removed`), delete (author only; `review_removed` if removed),
format (`review_invalid` for stars or body). Remove (admin only, G9): `review_moderation_forbidden`
for anyone else, owner included.

## 5. HTTP contract (OpenAPI + Orval)

| Method · path | Who | Result |
|---|---|---|
| `GET /api/v1/courses/{slug}/reviews?cursor=` | anyone | summary `rating: {average, count} \| null` + reviews newest first (keyset, 20 per page): display name, stars, body, completed/total, created/edited |
| `GET /api/v1/learn/courses/{courseId}/review` | signed in | own review or null, and `eligibility: ok \| <code>` so the form can explain a refusal |
| `PUT /api/v1/learn/courses/{courseId}/review` | signed in | create or edit own review; codes from §4 |
| `DELETE /api/v1/learn/courses/{courseId}/review` | author | 204; `review_removed` |
| `POST /api/v1/learn/reviews/{reviewId}/reports` | signed in, not the author | 201 or 200 on repeat (UNIQUE) |
| teach queue (existing route) | owner ∨ admin | lists review reports next to comment reports, with `kind`. One SQL: `UNION ALL` of the two open-report aggregates (comments as in `moderation.go:32-49`, reviews from `discussion_review_reports`), ordered by the existing key `(open_count DESC, latest_report DESC)` then `(kind, id) DESC`; the cursor carries all four. A review row has no lesson (lesson fields null). |
| `POST /api/v1/teach/reviews/{reviewId}/remove` | admin | 200; log entry with `review_id` |
| `POST /api/v1/teach/reviews/{reviewId}/reports/dismiss` | owner ∨ admin | as for comments |
| catalog `Course` | anyone | `+ rating: {average, count} \| null`, `+ lessonCount`. **No duration field in this phase** (AC11 makes it conditional): a real total needs the current ready asset of every lesson — a per-lesson `LATERAL` like `discussion/…/repository.go:28-34` — inside the hottest public read, and any lesson without a ready video makes the total false. Decision recorded in `assumptions.md`. |
| learning course | signed in | `+ canPurchase` = published ∧ paid ∧ no active grant ∧ not owner (role ignored) |
| admin grant (existing) | admin | `+ grant_self_forbidden` when target = actor (G5) |

Average is rounded to one decimal on the server; count is the counted reviews.

## 6. Frontend

- **Catalog card** (`presentation/storefront/CatalogView.tsx`): `★ avg (count)` or «Новий курс»,
  and the lesson count. No duration in this phase (§5).
- **Course page** (`CourseView.tsx`): rating summary and a reviews section, newest first, «Показати
  ще» by cursor; each review shows name, stars (accessible name «4 з 5»), text, «пройдено N% на
  момент відгуку», date and a report button for signed-in non-authors.
- **Learning page**: review form for the eligible (React Aria radio group for stars, text field with
  live count 20–2000), own review with edit/delete, and the reason text for each refusal code.
- **Admin buy entry**: when `canPurchase` is true and no lesson is locked for the viewer, the course
  outline shows «Купити курс», which starts the existing purchase flow (`purchase-queries`).
- **Teach queue**: review reports listed with the course; «Зняти» shown only to an admin; owner sees
  «Відхилити скаргу» as for comments.
- a11y as AGENTS: keyboard, visible focus, ≥ 44 px, light/dark, 200 % zoom; axe in e2e.

## 7. Side effects

- **Shared persisted state:** a new published-content kind counted on public pages. e2e that
  publish courses must clean up reviews with the course (FK cascade covers it). Two e2e count
  published courses — `catalog.spec.ts:52` (cards) and `recovery.spec.ts:91` (`pagination.total: 3`):
  adding fields does not change counts, but a ratings e2e that publishes a course must archive it.
- **Moderation log contract change:** `comment_id` becomes nullable. The log is write-only today —
  the only SQL touching it is the `INSERT` at `moderation.go:118`, no reader exists — so nothing
  breaks; the integration test of the table asserts both entry kinds and the CHECK.
- **Catalog query cost:** an aggregate join on the hottest public read. Measured in §9 against R6.
- **orders change (G5):** a new refusal in `GrantCourse`; existing admin-grant e2e (Pay10) grants to
  another user and stays green. Self-grants already present in a database are left as they are.
- **Abuse surface:** writes are per user per course (UNIQUE) and behind auth and `zone=api`.
- **Dependency failure.** The only dependency is Postgres; nothing new is external.
  - *Catalog:* the aggregate is a join inside `ListPublishedCourses`, the same statement on the same
    database, so it cannot fail apart from the catalog. When Postgres fails, the catalog fails as it
    does today: `dependency_unavailable` 503 and the existing error state (`CatalogStates.tsx`).
    Degrading to `rating: null` would need a second query, which R6 rejects; no partial card is
    shown.
  - *Course page:* the reviews list is its own request (`GET /courses/{slug}/reviews`). On failure the
    reviews section shows an error with «Спробувати ще», and the rest of the page (description,
    outline) renders from its own query (Rt28).
  - *Learning page:* the own-review request fails → the review block shows the same error state; the
    lesson and player are unaffected. A failed write keeps the typed text in the form (Rt29).
  - *Writes and moderation:* a repository error maps to `dependency_unavailable` 503 through the
    existing `statusFor` (`discussion/adapters/http/handler.go:395-396`); nothing is half-written,
    because each write is one statement or one transaction (remove = update + log insert together).

## 8. Migration plan and rollback

- `00013_ratings.sql` is additive except one relaxation (`comment_id DROP NOT NULL`). Up applies on
  a populated database without backfill. Down: delete log rows with `review_id`, drop the review
  tables, restore `NOT NULL`, drop the column — test-only, like earlier Downs.
- **Rollback of the feature:** revert the merge; run Down on disposable databases only. Retained
  test databases recover forward. The catalog contract fields are additive, so an older frontend
  ignores them.
- Generated code (sqlc, Orval) is regenerated in the same slice as the SQL/OpenAPI it follows;
  `check-generated` must stay green (needs staging).

## 9. R6 measurement

Seed for the measurement (not for e2e): 100 published courses with 3 lessons each, 20 000 reviews spread
unevenly (Zipf-like, the top course 2 000), then `VACUUM ANALYZE` (steady state, as autovacuum leaves a
running database). `EXPLAIN (ANALYZE, BUFFERS)` on the production `ListPublishedCourses` text with the
catalog defaults (empty search, sort `recent`, page 20), 5 runs after a warm-up, median. Harness:
`backend/internal/course/adapters/postgres/catalog_measure_test.go`, build tag `measure` (not in the gate).
**Budget: median ≤ 5 ms** (owner, 2026-10-03; replaced the earlier ">2× the query without the aggregate",
see `assumptions.md`). Over budget → R6 step up (materialized view), recorded as a decision first.

**Measured 2026-10-03 (T04, `f598d0e`):** 1.306 ms median with the aggregate (index-only scans, 0 heap
fetches); the same query without it 16 µs. Within budget — the read-time aggregate stays.

## 10. Constitution (AGENTS.md)

| Rule | How this plan keeps it |
|---|---|
| Domain imports only std lib and `internal/contracts` | eligibility is a pure function in `discussion/domain`; the threshold and rounding are one pure function in `internal/contracts/rating`, imported by `course/domain` and `discussion/domain` |
| Cross-module rule lives in `internal/contracts` (ADR 013) | `contracts/rating` imports no module; the existing test that forbids `contracts` from importing a module covers it |
| Modules never import each other | catalog reads the reviews table in `course` SQL; no package import (ADR 013/014 precedent) |
| Ports only for real external dependencies | no new port: reviews use the existing discussion repository port extended with methods |
| Generated models stay in adapters | sqlc/Orval types mapped explicitly |
| Explicit business error codes | §4 and §5 |
| No fetch in presentation; Query for server state | review queries in `app/composition`, adapters via `transport.ts` |
| Ordered gate; skipped is reported as skipped | §12 and the report |

## 11. Acceptance scenarios

Groups: happy path (Rt01–Rt05), boundaries (Rt06–Rt14), refusals and failures (Rt15–Rt29). Levels: unit (domain),
contract (HTTP), integration (SQL against Postgres), e2e (browser against Compose).

| # | Scenario | Precondition | Action | Observable result | Verified by | AC |
|---|---|---|---|---|---|---|
| Rt01 | Rate a free course | free course, student with one lesson in progress | sends 4 stars + 25-char text | 201; the review is on the course page with «пройдено 0%» | e2e `ratings.spec.ts`; contract | AC1, AC4 |
| Rt02 | Rate a paid course | paid course, buyer with an active purchase grant and progress | sends a review | 201 | contract; integration | AC1 |
| Rt03 | Edit own review | own review exists, right still held | changes stars 4 → 2 | the page shows 2 stars and an «змінено» date | e2e; contract | AC6 |
| Rt04 | Delete and rewrite | own review exists | deletes, then writes again | after delete the review is absent and the count drops by 1; the new write returns 201 | e2e; integration | AC6, G10 |
| Rt05 | Admin buys and rates | admin, paid course, no grant | opens the course outline | «Купити курс» is shown; after a stub-completed purchase and progress, a review returns 201 | e2e (offline WayForPay stub) | AC3, G2 |
| Rt06 | Threshold below | course with 2 counted reviews | visitor opens catalog | the card shows «Новий курс» | e2e; unit threshold table | AC8 |
| Rt07 | Threshold reached | course with 3 counted reviews (4, 5, 5) | visitor opens catalog | the card shows `★ 4.7 (3)` | e2e; integration (SQL aggregate) | AC8, AC11 |
| Rt08 | Threshold on course page | same 2 / 3 | visitor opens the course page | «Новий курс» / `★ 4.7 (3)` | e2e | AC8 |
| Rt09 | Removed review not counted | 3 reviews, one removed | catalog | the card shows «Новий курс» | integration; e2e | AC7, AC8 |
| Rt10 | Text length edges | eligible student | sends 19 / 20 / 2000 / 2001 characters (after trim) | 400 `review_invalid` / 201 / 201 / 400 | unit; contract | AC4 |
| Rt11 | Stars edges | eligible student | sends 0 / 1 / 5 / 6 | 400 / 201 / 201 / 400 | unit; contract | AC4 |
| Rt12 | Frozen percentage | review written at 2 of 4 lessons | course gains 2 lessons; author completes all | the review still shows «пройдено 50%» | integration | AC5, G6 |
| Rt13 | Recaptured on edit | same, then the author edits | edit | the review shows «пройдено 100%» of 6 | integration | AC5 |
| Rt14 | Public list, empty and paged | course with 0 reviews; course with 25 | anonymous visitor opens the page | 0: an empty-state text; 25: 20 shown, «Показати ще» loads 5 more, newest first | e2e; contract | AC9 |
| Rt15 | Owner refused | course owner holds a grant and progress | sends a review | 403 `review_owner_forbidden` | unit; contract | AC1, G2 |
| Rt16 | Preview only refused | paid course, user watched a preview lesson, no grant | sends a review | 403 `review_purchase_required` | integration; contract | AC1, G1 |
| Rt17 | Admin by role refused | admin, paid course, no grant, has progress | sends a review | 403 `review_purchase_required` | integration; contract | AC1, G2 |
| Rt18 | No progress refused | free course, signed in, no progress | sends a review | 403 `review_not_started` | unit; contract | AC1 |
| Rt19 | One review per user | own review exists | sends PUT again (create-or-edit) | 200; still exactly one review row for this author and course. Two concurrent first creates → one 201, one 409 `review_already_exists` | contract; integration (T03 concurrency subtest) | AC4 |
| Rt20 | Lost access: no edit | grant revoked after the review (fixture) | edits | 403 `review_purchase_required`; the review still shows and counts | integration | AC6, G4 |
| Rt21 | Owner cannot remove | owner, reported review on own course | calls remove | 403 `review_moderation_forbidden`; the review stays | contract; e2e (no «Зняти» button) | AC7, G9 |
| Rt22 | Admin removes | reported review | admin removes | the review leaves the page; the log shows the admin and `review_id` | e2e; integration | AC7 |
| Rt23 | Removed cannot be replaced | removed review | author sends a create; author deletes | 409 `review_removed`; 409 `review_removed` | contract | AC7, G10 |
| Rt24 | Self-grant refused | admin | grants a course to self | 403 `grant_self_forbidden`; no grant row | unit; contract; integration | AC2, G5 |
| Rt25 | Report own review refused / repeat | author; another user reports twice | report | author: 403; repeat: one report row | contract; integration | AC7 |
| Rt26 | Archive keeps reviews | course with reviews archived, grant holder | opens the course and edits | reviews visible; edit 200; after re-publish the card shows the rating | integration; e2e | AC10, G3 |
| Rt27 | No summary table | migration applied | inspect schema | no summary table, counter column or trigger for reviews | integration (catalog of `information_schema`) | AC12, R6 |
| Rt28 | Reviews request fails | the reviews request returns 503 | visitor opens the course page | the reviews section shows an error with «Спробувати ще»; the outline is shown | FE unit (presentation, query error state) | AC9 |
| Rt29 | Write fails | the review write returns 503 | eligible student submits | an error message is shown and the typed text stays in the form | FE unit | AC4 |

Manual rows: **0**. Every row has an automated level.

## 12. Slices (preview — `work/tasks.md` gets the rows after GATE 2)

| ID | Slice | Closes |
|---|---|---|
| T01 | `contracts/rating.Summary` + discussion domain: eligibility, format, progress capture — unit tables | Rt10, Rt11, Rt15, Rt18, unit parts of Rt06–Rt08 |
| T02 | `orders`: no self-grant | Rt24 |
| T03 | Migration 00013 + discussion repository (review facts, CRUD, reports, log) — integration | Rt02, Rt09, Rt12, Rt13, Rt16, Rt17, Rt20, Rt22, Rt25–Rt27 |
| T04 | Catalog SQL aggregate + `lessonCount` + `canPurchase`; R6 measurement | Rt07, Rt09, §9 |
| T05 | Application + HTTP + OpenAPI + Orval; contract tests | Rt01–Rt04, Rt10, Rt11, Rt14–Rt19, Rt21, Rt23–Rt25 |
| T06 | Frontend: card, course page, form, admin buy entry, teach queue | UI parts of Rt01–Rt09, Rt14, Rt21, Rt22, Rt26; Rt28, Rt29 |
| T07 | e2e `ratings.spec.ts` + a11y | e2e column |
| T08 | Ordered `make check`, report row by row, PR | all |

## 13. Report section

`report.md` will be a delta to this plan: measured gate, rows Rt01–Rt27 one by one with their
evidence, the R6 measurement, named limitations.
