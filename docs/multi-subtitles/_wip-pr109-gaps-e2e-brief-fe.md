# PR #109 — close E2E_ONLY and INCONCLUSIVE mutations with plain tests (part fe) — brief for Codex

Owner decision 2026-11-17 ("Закрити всі тестами"). Both re-review axes (`_wip-pr109-review2-codex-{standards,spec}.md`) block on these
rows. A static trace to an e2e row does not count. `asr.spec.ts` is excluded in CI, so each behavior needs a **plain deterministic test
that CI selects** (unit or integration), with executed baseline/mutant evidence.
Contract: `/Users/strelia/project-data/ai-factory-kit/skills/arch/code-review/references/behavior-coverage.md`.

- **Copy (the ONLY place you edit):** `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut`, now at `3e7de7e` (= branch HEAD `ad3612c` + docs). Clean. Never touch
  `/Users/strelia/projects/courses`.
- **Rows:** FE-B B007, B055, B145, B146, B147, B148, B152, B153, B155, B156, B184, B185, B192 (journal fe-b) and FE-A X-019 (journal fe-a). The exact mutation diff for each row is in `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/_wip-pr109-mutations-fe-b.md` (X-019: `_wip-pr109-mutations-fe-a.md`) (grep the id); the reviewers'
  reachability notes are in `docs/multi-subtitles/_wip-pr109-review2-codex-standards.md` (table "E2E_ONLY reachability") inside the copy.
- First re-apply each mutation against the **current** tests: some may already be killed by the 208 closure tests. If the whole suite
  fails attributably, record KILLED with the verbatim line and add no test.
- Otherwise add the missing discriminator at the seam the reviewers name. Examples:
  - H38–H40: route-level auth wiring through the real router or middleware, not a no-op auth;
  - Q05: task handler → Claim track id;
  - A14: regeneration marker;
  - S68: repository duration, integration;
  - FE-B: the CourseEditor/route composition prop chain via the real route composition.
  - X-019 (fe-a, `use-transcript-recovery.ts:140`): add a named async-retirement test that observes the rejection being contained (no unhandled rejection) and editing/Save retained; the mutated run must fail **that named test**, not only a global unhandled rejection.
- **Edit tests only.** Production bytes must stay identical: `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --name-only` shows only test files and test fixtures. If a
  correct test fails on unmutated code, mark the row DEFECT, keep the test out of the copy, and record the failure; that is STOP + QUESTION.
- Use no sleeps. Use independent expected values.

WORKING FILE: `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/_wip-pr109-gaps-e2e-fe.md` (append only). Start it with a `NEXT:` line before the first tool, then write
FACT / DECISION / NEXT / QUESTION entries.

Record one row per item: `id | test file::name | baseline PASS | mutated FAIL line (verbatim, named test) | KILLED-by-existing / CLOSED / DEFECT`.
Restore production bytes after each row.

## End
In `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut/frontend`: `npm run format`, `npm run format:check`, `npm run lint`, `npm run typecheck`, `npx rstest run` (whole suite) — exit codes and counts.
Final: the counts, `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --stat` (including untracked files), and the SHA-256 of each changed file.
