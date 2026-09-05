# Phase 4 living note — admin course management

Persist channel: repository file (no CLAUDE.local.md channel exists); flushed early and
often per spec-flow. Superseded entries stay, corrections append.

## 2026-09-05 — setup and spec

- Threshold called: Phase 4 crosses the spec-flow threshold (new subsystem, public API
  change, both stacks) — `.claude/standards.md` §Поріг.
- Branch `feat/admin-crud` from `main` = `e9edfd5`; artifacts home `docs/admin/`.
- `spec.md` written: intent only, 8 open questions carried to GATE 1 with
  recommendations. Standing decisions §24 of the product brief are treated as closed and
  not re-asked.
- GATE 1 (grilling) pending: user answers required on Q1–Q8 before any plan sentence.
- Delegation note: implementation is expected to go to Codex via detached briefs
  (working model), with the option of handing slices to peer session `courses-df`.

## 2026-09-05 — GATE 1, Q1 answered (scope widened)

User decision on Q1: **both roles mutate; instructor is ownership-scoped to own courses
now** — this amends brief §24.3 (ownership-scoped mutations arrive in Phase 4; the
instructor-workspace phase keeps students/progress/media views). Additionally named:
user administration (admin blocks/restores/deletes users, resets passwords; soft delete
with unique-email-among-non-deleted), per-course user blocking by instructors, and a
possible super-admin role (admin cannot delete an admin, only block). Placement of these
is being grilled next; per-course blocking has no substrate before enrollment exists.
