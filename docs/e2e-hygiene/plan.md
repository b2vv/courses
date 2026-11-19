# e2e-hygiene R5 — specs remove the published courses they create (2026-11-18)

Source: `docs/admin-editor-ux/plan.md` §5 R5; `work/AGENDA.md` (rebuild `8be7b428`), next move. Owner go 2026-11-18
("Берись за R5"). Branch `r5-e2e-fixture-cleanup` from `main` @ `8be7b428`. Size: Small (test code only, no product change).

## Checkpoint (Small)
- Problem (measured 2026-11-18 on Kiri): after four `chromium` project runs without `e2e-reset`, 28 of 31 published
  courses were e2e leftovers (`e2e-lb-*`, `e2e-refund-*`, `e2e-te-*`, 4 each); `catalog.spec.ts:3` then failed
  because the seeded course left catalog page 1. Mechanism: `lesson-builder.spec.ts:85`, `refund.spec.ts:107`,
  `teach-editor.spec.ts:113` publish via the admin API and nothing deletes them (AGENTS "Decisions" bullet 2: whoever
  adds a globally counted instance cleans up after itself).
- Decision (Courses 1): one shared e2e helper deletes a course and its non-cascading rows by SQL through the portable
  `docker compose exec postgres` call (the pattern of `subtitles-cleanup.ts` / `audio-tracks.spec.ts:227`); the two
  existing copies move to it. Archive via API is rejected: archived rows are still counted by admin stats
  (MS4 review w2, `docs/multi-subtitles/report.md`).
- Rows: H1 each of the three specs registers every created course for deletion before anything can fail and deletes
  it in `afterAll`; H2 the helper removes rows that do not cascade (grants/audit, assets, orders/payment events/refunds,
  per the FK definitions in `backend/migrations`); H3 existing callers use the shared helper with unchanged behaviour;
  H4 proof: published `e2e-%` count before = after a run of the three specs without reset, and `catalog.spec`
  passes after repeated runs.
- Verification: format/lint/typecheck; the three specs + subtitles + audio-tracks + catalog run twice on Kiri without
  reset, counts recorded; independent Codex review; ordered `make check` before merge (owner consent per run).
- review_mode=independent_subagent (Codex). Writer Courses 1 via Codex.
