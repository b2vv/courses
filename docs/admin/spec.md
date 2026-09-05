# Admin course management — Phase 4 specification

Status: GATE 1 decisions taken 2026-09-05; awaiting the user's shared-understanding
confirmation. Intent only: stack, files and architecture belong to the
plan. Roadmap source: `work/PRODUCT-BRIEF.md` §21 (Phase 4) and §24 (standing decisions).

## Problem

The platform authenticates three roles and shows a public catalog, but content can enter
the database only through hand-run seed SQL. Nobody can create, structure, publish or
retire a course through the product, so there is nothing real to sell, upload to (Phase 5)
or play back (Phase 6). Phase 4 also introduces the first server-authorized business
endpoints: today role checks exist only as client-side navigation convenience, and the
first protected handler has nothing to inherit (briefing §1, architectural fact #1).

## User scenarios

Phase 4 delivers two slices in order: **4a — course management** (this document's
scenarios), then **4b — identity administration** (outlined below, detailed when 4a
ships).

Actors: **admin** (manages all courses), **instructor** (creates courses and manages
only the courses they own — ownership-scoped from day one, GATE 1 Q1), **student /
anonymous** (must be refused by the server).

- **S1 Create.** An admin creates a course with a title, slug, description and an owner
  chosen from users with the instructor or admin role; an instructor creates a course
  and automatically owns it. It appears in the admin course list as `draft` and is
  absent from the public catalog and from its public URL.
- **S2 Structure.** An authorized editor (admin on any course, instructor on an owned
  course) adds modules to a course and lessons to modules, renames them and reorders
  them; the admin editor shows the resulting structure. A lesson is a title, plain-text
  description and order — no rich text, no media fields (Phase 5 adds its own).
- **S3 Publish.** An admin publishes a draft; the course becomes visible in the public
  catalog and on its public page.
- **S4 Archive.** An admin archives a published course; it leaves the catalog and its
  public URL stops resolving (Q8).
- **S5 Edit.** An admin edits fields of a course, module or lesson. Edits to a published
  course are immediately public; there is no draft-of-published or versioning (Q5 fixes
  deletion; versioning is out of scope).
- **S6 Refusal.** Anonymous requests receive the authentication-required answer; a
  student receives the role refusal; an instructor mutating a course they do not own
  receives the ownership refusal. Enforcement is server-side on every admin operation;
  the client guard is navigation convenience, never the mechanism.
- **S7 Admin UI.** The admin shell lists courses with status and opens a course editor
  for structure and fields. Forms validate input, uk/en localization, keyboard
  accessible, visible focus, ≥44px controls.

## Acceptance criteria

Measurable rows are delegated to `acceptance-spec` at plan time; these are the binding
outcomes they must cover.

1. Every course has an owner from its first migration; creating a course without an
   owner is impossible.
2. `draft` and `archived` courses are absent from the public catalog and public course
   page; publish and archive toggle exactly that visibility.
3. Status transitions follow the agreed matrix (Q3); an illegal transition is refused
   with an explicit business error, never a 500.
4. Courses, modules and lessons support create/read/update(/delete per Q5) with explicit
   ordering; course slugs are unique.
5. Every admin operation requires an authenticated session **and** an authorizing role,
   enforced at the handler level and proven by contract tests.
6. The OpenAPI document and generated clients change together; the drift gate stays
   green.
7. The admin UI performs every scenario above through the API, localized and accessible.
8. The full manifest gate (`make check`) passes with the new surface, from a clean
   checkout.

## Out of scope

- Video upload, transcoding, storage; lesson media beyond a named placeholder (Phase 5).
- Playback, enrollment, student progress (Phase 6); transcripts and comment threads
  (Phase 7); orders and payments (Phase 8).
- Instructor workspace (students, media, progress, moderation views): its own late
  phase. Phase 4 *does* enforce ownership on mutations (GATE 1 amended brief §24.3) but
  builds no instructor-specific views beyond the shared admin shell.
- Per-course user blocking by instructors: no substrate until enrollment exists;
  deferred to the enrollment-carrying phase (GATE 1, Q1a).
- Registration, self-service password reset, MFA — the identity out-of-scope list
  stands. Admin-driven user administration is slice 4b, not out of scope.

## Slice 4b — identity administration (outline)

Detailed specification follows when 4a ships; the agreed shape (GATE 1):

- Admin lists users; blocks and restores them (`users.disabled_at` already refuses
  authentication — 4b gives it an operator). Admin soft-deletes users: a deleted status,
  with email uniqueness enforced among non-deleted users.
- Admin resets a password by setting a new one directly — no mailer exists, so no email
  flow is promised.
- No fourth role. Domain invariants instead: an admin cannot be deleted and cannot have
  their role changed through the API; an admin may be blocked, but not by themselves.
  A super-admin role becomes its own decision when a real admin hierarchy appears.

## GATE 1 decisions (2026-09-05)

| # | Decision |
|---|---|
| Q1 | Both roles mutate; **instructor is ownership-scoped now** (amends brief §24.3). User administration split off as slice 4b; per-course blocking waits for enrollment. |
| Q1b | No super_admin role; admin-protection invariants in the domain instead. |
| Q2 | Admin and instructor both create courses; instructor becomes owner automatically, admin picks the owner; only admin reassigns ownership. |
| Q3 | Transitions: `draft→published→archived` plus `archived→published` restore. Nothing else. |
| Q4 | An empty course may be published; publish validates fields only. |
| Q5 | Hard delete for draft courses only (cascade); modules/lessons deletable always as structure edits. Published/archived courses are archived, never deleted. |
| Q6 | The public course page shows the program (module/lesson titles) — text only. |
| Q7 | Lesson = title + plain-text description + order. No media columns yet. |
| Q8 | Archived course public URL answers not-found, indistinguishable from draft. **Standing principle recorded in brief §24: a purchased course never loses access on archive — ownership, not rental** (binds Phases 6/8). |
