# Identity administration — Phase 4b specification

Status: GATE 1 decisions taken 2026-09-06. Intent only. Builds on the 4a authorization seam
(`docs/admin/report.md` §6; briefing fact #1). Roadmap: `work/PRODUCT-BRIEF.md` §21 (4b
right after 4a) and §24.8 (the agreed shape).

## Problem

Users exist only through the local seed and the (future) bootstrap path; once created,
nobody can manage them through the product. An admin cannot block a compromised or
departing account, cannot restore one blocked by mistake, cannot remove a test account,
and cannot reset a password for someone locked out. The identity domain already carries
`disabled_at` and refuses a disabled user at authentication (`internal/identity/domain`),
but no operator reaches it. 4a built the server authorization seam and
`Role.CanAdministerAccounts()` (admin-only); 4b is its first non-course consumer.

## User scenarios

Actors: **admin** (the only actor with account administration), **instructor / student /
anonymous** (refused by the server on every 4b operation).

- **S1 List.** An admin opens a user directory and sees every user — email, display name,
  role, and status (active / blocked / deleted) — with search/paging like the course
  list.
- **S2 Block.** An admin blocks an active user; that user can no longer authenticate, and
  existing sessions stop working. The admin cannot block themselves.
- **S3 Restore.** An admin restores a blocked user; they can authenticate again.
- **S4 Soft delete.** An admin deletes a user; the row is retained as deleted (audit and
  email-uniqueness), the user cannot authenticate, and their email is freed for reuse only
  among non-deleted users (a new active user may take a deleted user's email; two
  non-deleted users may not share one).
- **S5 Password reset.** An admin sets a new password for a user directly (no mailer
  exists, so no email flow is promised); the user can then sign in with it.
- **S6 Admin protection.** An admin cannot be deleted and cannot have their role changed
  through the API; an admin may be blocked, but not by themselves. (No super_admin role —
  invariants live in the domain.)
- **S7 Refusal.** Every 4b operation requires an authenticated admin; instructor,
  student and anonymous receive the server refusal, enforced at the handler.
- **S8 Admin UI.** A user-administration screen under `/admin`: the directory, per-row
  actions (block/restore/delete/reset), localized, accessible, same kit as 4a.

## Acceptance criteria (binding outcomes; measurable rows delegated to acceptance-spec)

1. Blocking a user prevents new authentication and invalidates existing sessions;
   restoring re-enables both.
2. Soft delete retains the row, blocks authentication, and enforces email uniqueness only
   among non-deleted users; a deleted user's email is reusable by a new user.
3. An admin cannot delete an admin, cannot change an admin's role, and cannot block
   themselves — refused with explicit business codes, never a 500.
4. Password reset sets a valid Argon2id credential the user can authenticate with; the
   old password stops working.
5. Every 4b operation is refused server-side for non-admins and proven by contract tests.
6. The OpenAPI document and generated clients change together; the drift gate stays green.
7. The admin UI performs every scenario, localized and accessible.
8. The full manifest gate (`make check`) passes with the new surface, from a clean
   checkout.

## Out of scope

- User creation / self-registration, self-service password reset, MFA, OAuth (identity
  out-of-scope list stands; only admin-driven administration is added).
- Per-course user blocking / enrollment-scoped access (no substrate until enrollment).
- Instructor workspace and any instructor-facing user views.
- Audit log UI (the deleted/blocked timestamps are retained; surfacing a log is later).
- Email delivery of any kind.

## GATE 1 decisions (2026-09-06)

| # | Decision |
|---|---|
| Q1 | Block and soft-delete invalidate existing sessions immediately, in the same transaction as the status change. |
| Q2 | Password reset reuses the login password bounds and revokes all the user's sessions (recommendation, uncontested). |
| Q3 | Role change is out of scope for 4b — the four named operations only (block/restore/delete/reset). |
| Q4 | The directory shows deleted users under a deleted status filter, mirroring the course list's status tabs (recommendation, uncontested). |
| Q5 | Blocking or deleting the last active admin is refused — a domain invariant, checked transactionally. |
| Capability model | **Capabilities are additive** (2026-09-06). The `role` enum names only authoring/admin capability; any authenticated user can enrol and learn once enrollment exists, so an instructor is naturally a student of another instructor's course with no role change. Recorded in `work/PRODUCT-BRIEF.md` §24.10; binds Phase 6 enrollment. 4b changes no role, so it only consumes this model, it does not build on it.
