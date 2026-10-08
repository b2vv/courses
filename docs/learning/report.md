# Learning phase — report

Live log of the learning phase (student overview for instructor and admin, author
block from the queue). Spec: [spec.md](spec.md). Plan: [plan.md](plan.md). The full
row-by-row report is LR6; this file records each slice as it closes.

## Slices

| Slice | State | Commits |
|---|---|---|
| LR0 | done 2026-10-08 | `43fabc9`, `26830a5`, `4a7166c`, `87753f8` |
| LR1 | done 2026-10-08 | `e7214ae` |
| LR2 | next | — |
| LR3–LR6 | pending | — |

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

**Next.** LR2 — backend of the overview: port, SQL (same progress formula as the
roster), service, HTTP. integration (Postgres) + contract.
