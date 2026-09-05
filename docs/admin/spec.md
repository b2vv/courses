# Admin course management — Phase 4 specification

Status: draft, awaiting GATE 1. Intent only: stack, files and architecture belong to the
plan. Roadmap source: `work/PRODUCT-BRIEF.md` §21 (Phase 4) and §24 (standing decisions).

## Problem

The platform authenticates three roles and shows a public catalog, but content can enter
the database only through hand-run seed SQL. Nobody can create, structure, publish or
retire a course through the product, so there is nothing real to sell, upload to (Phase 5)
or play back (Phase 6). Phase 4 also introduces the first server-authorized business
endpoints: today role checks exist only as client-side navigation convenience, and the
first protected handler has nothing to inherit (briefing §1, architectural fact #1).

## User scenarios

Actors: **admin** (course management), **instructor** (admitted to the admin shell today;
extent of Phase 4 access is Q1), **student / anonymous** (must be refused by the server).

- **S1 Create.** An admin creates a course with a title, slug, description and an owner
  (a user with the instructor or admin role). It appears in the admin course list as
  `draft` and is absent from the public catalog and from its public URL.
- **S2 Structure.** An admin adds modules to a course and lessons to modules, renames
  them and reorders them; the admin editor shows the resulting structure.
- **S3 Publish.** An admin publishes a draft; the course becomes visible in the public
  catalog and on its public page.
- **S4 Archive.** An admin archives a published course; it leaves the catalog and its
  public URL stops resolving (Q8).
- **S5 Edit.** An admin edits fields of a course, module or lesson. Edits to a published
  course are immediately public; there is no draft-of-published or versioning (Q5 fixes
  deletion; versioning is out of scope).
- **S6 Refusal.** Anonymous requests receive the authentication-required answer; a
  student receives the role refusal. Enforcement is server-side on every admin operation;
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
- Instructor workspace and ownership-scoped views ("my courses only"): Phase 4 records
  ownership and never builds views on it (`work/PRODUCT-BRIEF.md` §24.3).
- User management (creating/disabling users, assigning roles): fixtures and future
  bootstrap remain the only way users exist.
- Registration, password reset, MFA — the identity out-of-scope list stands.

## Open questions (GATE 1)

| # | Question | Recommendation |
|---|---|---|
| Q1 | Does the instructor role get Phase 4 course-management access — all courses, or none? | Mutations admin-only; instructor keeps shell access and sees the course list read-only. Ownership-scoped instructor work is its own phase. |
| Q2 | Owner assignment: chosen at create from existing instructor/admin users; reassignable later? | Required at create, reassignable by admin; both are ordinary updates. |
| Q3 | Reverse transitions: unpublish (`published→draft`)? restore (`archived→published`)? | Allow `archived→published` as the undo; no `published→draft` — retiring means archive. Matrix stays three states, three transitions. |
| Q4 | May an empty course (no modules/lessons) be published? | Yes. The catalog already renders structureless courses; a structural gate is invented policy. |
| Q5 | Is hard delete available, and for which states? | Delete drafts only (course/module/lesson in a draft course); published/archived are retired, never deleted. Module/lesson deletion inside a published course: allowed, it is an edit (S5). |
| Q6 | Does the public course page show the program (module/lesson titles) in Phase 4? | Yes — the data exists and it is the first visible payoff; text only, no media claims. |
| Q7 | What is a lesson in Phase 4? | Title + plain-text description + order. No rich text (TipTap waits for a real editor need), no media fields beyond a nullable placeholder for Phase 5. |
| Q8 | Public URL of an archived course? | Not-found. No enrollment exists, so nobody loses access they had; an "archived" page is invented surface. |
