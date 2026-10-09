# Learning phase — report

Live log of the learning phase (student overview for instructor and admin, author
block from the queue). Spec: [spec.md](spec.md). Plan: [plan.md](plan.md). The full
row-by-row report is LR6; this file records each slice as it closes.

## Slices

| Slice | State | Commits |
|---|---|---|
| LR0 | done 2026-10-08 | `43fabc9`, `26830a5`, `4a7166c`, `87753f8` |
| LR1 | done 2026-10-08 | `e7214ae`, `2f4ae01` |
| LR2 | done 2026-10-08 | `dcbabf6` |
| LR3 | done 2026-10-08 | `be94327`, `4df5e5f` |
| LR4 | done 2026-10-09 | `6c39762` |
| LR5 | done 2026-10-09 | `ebd9c72` (+ e2e nginx burst fix `fade1b8`) |
| LR6 | done 2026-10-09 | this report + the LR2 review nits |

## LR0 — G8: an administrator is never a learner

**What.** Enroll, StartPurchase and GrantCourse refuse to make an administrator a
learner. One named code `admin_not_learner`, HTTP **403** (the `*_forbidden` family —
a role refusal, not a state conflict; the plan's "409 — звірити наявні" resolved to
403 against the existing convention, Courses 1, 2026-10-08). The recipient's role is
read through a new `GrantRepository.IsAdministrator`, implemented over the `users`
table the grants adapter already reads for the roster (no new port or constructor
param). `CanPurchaseBy` excludes the administrator, so the learning outline offers
them no purchase. The storefront `CourseView` hides "Записатися" from an admin
(`viewerIsAdmin`), who still sees the course and the start-learning link (exempt).

**Verification (measured, 2026-10-08).**
- unit `TestAdministratorIsNeverALearner` — enrol, purchase and grant-to-admin each
  refuse with `admin_not_learner` and write no membership/order (0 create calls).
- contract (HTTP) — admin enrol, admin purchase and grant-to-admin-recipient each
  return 403 with `"code":"admin_not_learner"`, asserted against the OpenAPI schema.
- integration (Postgres) — `Lo18 …`: `IsAdministrator` true for an admin, false for a
  student; grant to a second admin returns `ErrAdminNotLearner` and writes 0 rows.
- `TestLearningCourseCanPurchaseBy` — admin without grant → `canPurchase=false`.
- full `make e2e-local` — **108 passed (12.8m)**; includes Rt05 (a student, not an
  admin, buys and reviews) and Lo19 (admin sees the free course, no enrol button).
- `go test ./...` (backend) green; frontend typecheck/lint/format/unit (505) green;
  `make generate` touched only `errorResponseErrorCode.ts` and `learningCourse.ts`.

**Non-blocking from GATE 2 (Courses 2) addressed in LR0.** All three write paths are
closed at the service, not only the UI (`GrantCourse` now refuses an admin recipient,
not just a self-grant).

**Limitations.** None specific to LR0. The OpenAPI error enum gained `admin_not_learner`
here (ahead of LR1) because the LR0 contract test asserts the code against the schema.

## LR1 — OpenAPI for the overview and block-author

**What.** The contract surface the next slices implement. `GET /api/v1/teach/learners`
and `/api/v1/admin/learners` return a `LearnerPage` (`LearnerRow` — no email, G3;
`LearnerCourse`; `nextCursor` nullable). Four `POST …/block-author` /`…/unblock-author`
paths for comments and reviews return `MembershipGrant`. New error codes
`learner_course_forbidden` (403) and `grant_transition_invalid` (409). Orval models
regenerated.

**GATE 2 non-blocking applied.** #1 — the author-without-membership path uses the
real `grant_transition_invalid`, not the invented `membership_transition_invalid`
(added to the error enum here). #2 — `LearnerPage.nextCursor` is `string|null`, the
roster's precedent, so the last page returns `null` not `""`.

**Decision — `authorBlocked` moved to LR3 (Courses 1, 2026-10-08).** The plan placed
`authorBlocked` on `ModerationQueueItem` in LR1. Added as a required field it reddens
the discussion module's queue contract test at once, because the handler does not emit
it yet; emitting it (an `EXISTS` over `course_grants` per item) is LR3 backend work.
So the schema field, its handler emission and the Lo14 "state not identifier"
assertion move to LR3 together, and LR1 leaves `ModerationQueueItem` unchanged.

**Verification (measured, 2026-10-08).** `make generate` + `make check-generated` —
clean (Orval round-trips; only the new learner models, `courses.ts`,
`errorResponseErrorCode.ts` and the model index changed). Backend `go build ./...`
and `go test ./...` green (the additive OpenAPI change breaks no contract test — the
discussion queue test, which failed on a required `authorBlocked`, is green after the
deferral). Frontend typecheck, lint, unit (638 cases) green. No HTTP contract test for
the learner endpoints yet — they gain handlers in LR2/LR3, which is where their
contract tests land.

**Carry-forward from the LR1 review, closed in `2f4ae01`.** `grant_transition_invalid`
entered the enum in LR1, but `writeError` had no case for `ErrGrantTransitionInvalid`,
so blocking a non-member through the roster (a reachable direct-API path) fell to a
500 instead of the contract's 409. Added the mapping and a contract test; this is the
same refusal LR3's block-author needs for an author who is not a member.

**Next.** LR2 — backend of the overview: port, SQL (same progress formula as the
roster), service, HTTP. integration (Postgres) + contract.

## LR2 — learner overview backend

**What.** `GrantRepository.Learners` reads the overview in two steps: a keyset page
of users ordered by `(lower(display_name), user_id)`, then the visible course detail
for that page. The per-member progress percent is a single shared SQL fragment
(`progressJoins`/`progressExpr`) that `Roster`, `MyMemberships` and `Learners` all
use, so Lo02 cannot drift from the roster (the plan's "shared fragment" over a DB
function, since the phase adds no migration). A row carries only the viewer's visible
courses, their state and progress, and `lastActivityAt = max(lesson_progress.updated_at)`
(null without progress); no email (G3); a deleted account is anonymised to `""`. The
service separates the instructor view (own courses) from the admin view (every course,
administrator only); an instructor filtering by a course they do not teach — or by an
unknown course — is refused `learner_course_forbidden`, not told whether it exists.

**Verification (measured, 2026-10-08).** unit `TestLearnersScopesAuthorizesAndClampsLimit`
(scope, authz, limit clamp, filter passthrough). contract in `TestEnrollmentHTTPMatchesOpenAPI`
(page shape with no email, malformed-query rejection, foreign-course 403, student-on-admin
403). integration `Lo01-Lo09 …` against Postgres (scoping, aggregation, progress equals the
roster's, revoked excluded, `lastActivityAt`, keyset stable when a learner becomes visible
between pages, admin breadth, state/name/course filters). `go build ./...`, `go test ./...`
and the **full `make integration`** green; `Roster`/`MyMemberships` stayed green through the
shared-fragment refactor.

**LR0 follow-up closed here.** The full `make integration` surfaced that `L2b` asserted an
administrator could self-enrol, which G8 forbids. `go test ./...` and the browser e2e do not
run the integration-tagged suite, so it passed LR0's gate; the case now asserts
`admin_not_learner`. Lesson recorded: an enrolment/membership change runs `make integration`
in the same slice, not only `go test` and e2e.

**Next.** LR3 — backend of the block-from-queue: `AuthorOf(kind, id)`, `BlockAuthor`/
`UnblockAuthor` over the same `transitionMembership`, G6 in both the queue and roster paths,
and `authorBlocked` on `ModerationQueueItem` (schema + handler + Lo14), deferred from LR1.
integration + contract. LR3's block-author inherits the 409 `grant_transition_invalid`
mapping added in `2f4ae01`.

## LR3 — block from the queue + authorBlocked

**What. LR3a (orders, `be94327`):** `AuthorOf(kind, id)` resolves a comment (via
lesson→module→course) or review (via `course_id`) to its author and course — a SQL read of
the discussion tables, no module import (ADR 013). `BlockAuthor`/`UnblockAuthor` run the SAME
`transitionMembership` as the roster, now threading a reason (`moderation_queue:<kind>:<id>`)
into the audit — one rule, one record (Lo17). G6 (admin or course owner) is enforced inside the
transition; an administrator author is never a member (G8), so there is nobody to block and the
transition refuses `grant_transition_invalid` (the 409 mapped in `2f4ae01`). Unknown item → named
404 (`comment_not_found`/`review_not_found`). Four `teach` routes return the membership grant.

**LR3b (discussion, `4df5e5f`):** the moderation queue carries `authorBlocked` per item — an
`EXISTS` over `course_grants` (state=`blocked`) for the item's author and course — so the UI can
label the control. State only, never the author's id (Lo14). `ModerationQueueItem` gains the
required `authorBlocked` in OpenAPI (the field deferred from LR1); Orval regenerated.

**Verification (measured, 2026-10-08).** unit (`TestBlockAuthorFromQueue…`: resolve, reason,
unknown, stranger-G6, non-member-G8). contract (block/unblock comment+review → grant; unknown →
named 404; queue `authorBlocked` reflects state and leaks no author id). integration against
Postgres (orders `Lo12-Lo17…`: AuthorOf both kinds + unknown, block writes the roster's blocked
grant with the origin reason, idempotent, unblock; discussion queue `authorBlocked` false→true
after a real blocked grant). `go build`/`go test ./...`, FE typecheck/lint, `check-generated`
green. Authoritative gate green: full `make integration` (every package ok) + `make e2e-local`
**108 passed (13.3m)** — the shared membership state the transition touches held across the full
browser suite.

**Next.** LR4 — frontend of the overview (teach + admin). LR5 — frontend of the queue buttons.

## LR4 — frontend of the learner overview (teach + admin)

**What (`6c39762`).** `/teach/learners` and `/admin/learners` over the `teach` domain:
`LearnerRow`/`LearnerCourse`/`LearnerPage` types (no email — Lo07 at the type level), a
`listLearners(scope, filters, cursor)` port, a schema-validating adapter, an infinite
query, and a shared `LearnerOverview` presentation. A row shows the name (anonymised for a
deleted account), last activity (or "no activity yet" when null), and each visible course as
a `Link` into its roster with state and a progress bar. Filters (`courseId`/`state`/`q`) are
Router search params; the cursor is the "Show more" load-more, not a URL field. The «Учні» nav
item is on both the teach and admin surfaces; uk/en strings paired.

**Lo07 held at the frontend, not only by contract (Courses 2 review).** The adapter maps each
row field by field rather than passing the wire object through, so no wire field — email above
all — reaches the domain; the strengthened test carries an `email` on the wire and asserts it is
dropped.

**Verification (measured, 2026-10-09).** unit — adapter validation (valid page maps, email
dropped, malformed → `invalid_response`, 403 → `learner_course_forbidden`), query key, and the
`LearnerOverview` component (rows, former-author, never-active, empty, error+retry, load-more
`isPending` stays focusable, course link href). e2e — Lo10 (row → roster) and Lo11
(axe light/dark, keyboard focus, ≥44px, 200% zoom) under the e2e overlay. `make format-check`,
`typecheck`, `lint` = 0; frontend unit 649 passed; `route-guards` R24 lists both new routes.

## LR5 — block / unblock an author from the queue

**What (`ebd9c72`).** `ModerationQueueItem` gains `authorBlocked`, so each comment and review
row shows **Block author** / **Unblock author**. A `blockAuthor`/`unblockAuthor` port
(`kind: comment|review`) calls the four teach endpoints; the mutation re-reads the workspace so
`authorBlocked` flips. The control is per-row `isPending` (not `isDisabled`, which would drop
focus) scoped by the item id — never the author id (Lo14); errors render as named text. Refusal
codes were verified against the orders handler, not the plan: **403 `enrollment_forbidden`**,
**409 `grant_transition_invalid` | `membership_conflict`**, **404 `comment_not_found` |
`review_not_found`** (the plan's `membership_transition_invalid` does not exist).

**A real toggle bug caught in review (Courses 2), fixed before commit.** `authorPending` was a
bare `authorActing === id` with no `&& mutation.isPending` multiplier, so once an author was
blocked the flipped Unblock button stuck in `isPending` and could never be clicked — the toggle
was one-way. Fixed to mirror `CourseRoster`; the Lo13 e2e now round-trips Block → Unblock →
re-block to guard it. The green unit + single-click e2e had not caught it (the unit hard-codes
`authorPending`, the e2e clicked once) — the exact "a green gate is not evidence for a row" case.

**Verification (measured, 2026-10-09).** unit — adapter (block/unblock both kinds → void, named
refusals, queue `authorBlocked` parsed), `ReviewModerationRow` (label by `authorBlocked`, calls
handler), query fake. e2e — Lo12 (block a comment author → a real **403 `enrollment_blocked`**
on a new POST and **200** on GET: blocked from communicating, still reading) and Lo13 (block a
review author → the toggle round-trips and the review stays visible in the public list).
`make format-check`/`typecheck`/`lint` = 0; frontend unit 653 passed; **full `make e2e-local`
112 passed, exit 0** — the shared membership state the transition touches held across the whole
browser suite.

**e2e nginx burst fix (`fade1b8`, separate commit).** The full `make e2e-local` deterministically
429'd `purchase-wayforpay` (Pay24): the whole suite drives its API from one IP and under host load
the status polling + payment callbacks out-ran the strict `api` zone burst (10 r/s, burst 20). The
429 is at nginx, before the app — unrelated to the learning phase, which touches no payment or
nginx config. Fixed by the existing `login-req.e2e` pattern: the api burst moved to an include
(`api-req.conf`, production burst 20 unchanged), `docker-compose.e2e` mounts a larger-burst variant
(`api-req.e2e.conf`, burst 200). `recovery.spec`'s deliberate-abuse probe now floods 300 concurrent
(was 60) so it still out-runs the generous e2e burst. Production rate limiting is untouched.

## LR6 — the LR2 review nits

**What.** Two non-blocking items from the LR2 review, applied here (both internal; no client
contract change — each still maps to the same HTTP code):
- **`ErrInvalidQuery`.** The learner handler validated `courseId`/`state`/`q`/`limit` and the
  query shape with `ErrInvalidCursor` (identity `roster_cursor_invalid`), which reads wrong for a
  non-cursor parameter even though it maps to `invalid_request`. A new `ErrInvalidQuery`
  (`learner_query_invalid`) now names those; it maps to the same 400 `invalid_request`.
- **`LearnerPage.NextCursor` is `*string`.** The port carried the cursor as `string` with a `""`
  sentinel for the last page; it is now `*string` (nil on the last page), so the null is native
  through the port and the handler emits it without the `"" → null` conversion. The roster keeps
  its string sentinel (out of scope).

**Verification (measured, 2026-10-09).** `gofmt` clean; `go build`/`go vet` (incl. `-tags
integration`) clean; orders unit green; `TestEnrollmentHTTPMatchesOpenAPI` green (the
malformed-query rejection and page shape are unchanged); the orders Postgres integration test
(the Lo06 keyset chains `*NextCursor`) green. The phase-closing gate is the ordered `make check`.

## Acceptance, row by row (Lo01–Lo19)

Each row with where it is proven. Levels: C contract, I integration, U unit, E e2e.

| Row | Proven by |
|---|---|
| Lo01 instructor sees each of their learners once; other instructors' courses excluded | I `TestOrdersRepositoryAgainstPostgreSQL` (scoping); E `learning.spec` Lo10 row present |
| Lo02 row carries only the viewer's visible courses, each with state + the roster's progress % | I shared `progressExpr`/`progressJoins` asserted equal to the roster |
| Lo03 `revoked` hidden; a learner with only revoked memberships has no row | I integration (revoked excluded) |
| Lo04 last activity = max `lesson_progress.updated_at` over visible courses; null without progress | I integration; U/E `neverActive` rendering |
| Lo05 course filter (own only; foreign → coded refusal), state filter, name search | C `TestEnrollmentHTTPMatchesOpenAPI` (foreign-course 403, malformed query); I filters |
| Lo06 keyset stable when a learner becomes visible between pages | I integration (Carol revoked→active between pages) |
| Lo07 email never returned to instructor or admin; deleted account anonymised | C contract (no email in shape); U adapter drops an email on the wire; frontend maps field by field |
| Lo08 student/non-instructor 403 on teach overview; non-admin 403 on admin overview | C contract (student-on-admin 403); route guards |
| Lo09 admin sees learners of every course, same row format | I integration (admin breadth); E `/admin/learners` |
| Lo10 a row links through to the course roster | E `learning.spec` Lo10 |
| Lo11 overview a11y: keyboard, visible focus, ≥44px, light/dark, 200%, axe | E `learning.spec` Lo11 |
| Lo12 block a comment author → membership blocked; author still reads, cannot communicate | I integration (blocked grant); E `learning.spec` Lo12 (403 `enrollment_blocked` on POST, 200 on GET) |
| Lo13 block a review author → same; the review stays visible | E `learning.spec` Lo13 (toggle round-trips; review still in public list) |
| Lo14 queue response carries no author identifier | C contract (`authorBlocked` state only); per-row pending keyed by item id |
| Lo15 named refusals (non-member author, foreign course, unknown item); documented permissions | C + I (orders handler + integration; verified codes `enrollment_forbidden`/`grant_transition_invalid`/`*_not_found`) |
| Lo16 unblock restores `active`; repeat block/unblock idempotent or named conflict | I integration (unblock, idempotent) |
| Lo17 block from the queue and from the roster write the same record (state + audit) | I integration (same `transitionMembership`, origin reason in the audit) |
| Lo18 G8: admin cannot free-enrol, start a paid purchase, or be granted — coded refusal, no membership | C + I `TestAdministratorIsNeverALearner` / `admin_not_learner`; the `L2b` self-enrol case corrected under full `make integration` |
| Lo19 G8 UI: admin sees the course but no Enrol/Buy; an instructor on another course still enrols (En14) | E (Rt05 rewritten to a student buyer; admin-as-participant swept) |

## 5. Limitations

- **Per-course author blocking only.** Block/unblock acts on the `(author, course)` membership
  reached from the reported comment or review. There is no cross-course "block this person
  everywhere" action; an author active in several courses is blocked per course.
- **No cross-course learner/progress overview.** The overview lists learners within the viewer's
  course scope (instructor: own courses; admin: all), one row per learner aggregating their
  visible courses. It is not a per-learner profile or a global progress dashboard
  (`docs/enrollment/report.md` §6 carries the same boundary for the roster).
- **`q` is a name search only**, ILIKE over `display_name` with `%`/`_` escaped, 1–80 runes. It
  does not match email (never exposed) or course titles.
- **The e2e api-zone burst is relaxed for the browser suite** (`api-req.e2e.conf`, burst 200) so
  the one-IP runner is not anti-abuse-429'd; production stays at burst 20. The method: the suite
  mounts the larger-burst include over the strict one, and `recovery.spec` floods 300 concurrent
  to still prove the limit fires. Production rate limiting is unchanged and unverified by this
  relaxation — it is exercised only by `recovery.spec`'s deliberate flood in the e2e stack.
- **Reviewer blocked from the queue is a membership transition, not review removal.** Blocking a
  review's author leaves the review visible (Lo13); removing a review is the separate admin-only
  moderation action (G9), unchanged by this phase.
