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
| LR4–LR6 | pending | — |

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
