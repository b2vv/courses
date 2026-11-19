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

# e2e-hygiene R6 — specs delete the archived courses they create (2026-11-19)

Source: `work/AGENDA.md` M19 (rebuild `fe61852a`), `docs/e2e-hygiene/report.md` §Limitations. Owner go 2026-11-19
("M19"). Branch `r6-e2e-archived-cleanup` from `main` @ `fe61852a`. Size: Small (test code only, no product change).

## Checkpoint (Small)
- Problem (measured 2026-11-19 on Kiri, `select status, prefix, count(*) from courses`): after the R5 gate and two proof
  runs, 39 archived `e2e-%` courses (ratings 9, enrollment 8, instructor 5, discussion 4, learning 4, wayforpay 3,
  admin 2, transcript 2, purchase 1, accounts 1) plus 1 draft `e2e-transcript-*`. Mechanism: these specs end with an
  archive call through the admin API (published courses cannot be deleted by the API, 409 `course_not_deletable`), so
  each run adds rows that admin statistics count (AGENTS "Decisions" bullet 2).
- Decision (Courses 1, 2026-11-19): every course these specs create is registered right after its create response and
  removed in `afterAll` by `deleteE2ECourse` (R5 helper), `Promise.allSettled` + `AggregateError` like
  `lesson-builder.spec.ts:16`. Archive calls that only served as cleanup are removed. Archive calls that are the
  behaviour under test (e.g. `admin.spec.ts` P20, `ratings.spec.ts` Rt26/Rt30) stay unchanged; their courses are
  deleted afterwards too. Seeded fixtures (`archived-course`, `draft-course`, seed published courses) are not touched.
- Rows: H5 each of the ten specs (plus the transcript draft) deletes every course it created, including on failure;
  H6 no test assertion changes; H7 proof: `e2e-%` counts by status are equal before and after two runs of the ten specs
  without reset; H8 after the ordered gate on a fresh database, `e2e-%` courses = 0 in every status.
- Verification: format/lint/typecheck; the ten specs (+ catalog) twice on Kiri without reset, counts recorded;
  independent Codex review; ordered `make check` before merge (owner consent per run).
- review_mode=independent_subagent (Codex). Writer Courses 1 via Codex.
