# Learning phase — report

Live log of the learning phase (student overview for instructor and admin, author
block from the queue). Spec: [spec.md](spec.md). Plan: [plan.md](plan.md). The full
row-by-row report is LR6; this file records each slice as it closes.

## Slices

| Slice | State | Commits |
|---|---|---|
| LR0 | done 2026-10-08 | `43fabc9`, `26830a5`, `4a7166c` |
| LR1 | next | — |
| LR2–LR6 | pending | — |

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

**Next.** LR1 — OpenAPI for the overview + block-author + `authorBlocked`, and
generation (`make check-generated`). GATE 2 non-blocking #1 (use the real code names,
not `membership_transition_invalid`) and #2 (keyset last page `nextCursor=null`, not
`""`) apply to LR1/LR2/LR3.
