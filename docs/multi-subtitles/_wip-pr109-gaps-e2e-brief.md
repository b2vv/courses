# PR #109 — close E2E_ONLY and INCONCLUSIVE mutations with plain tests (part {PART}) — brief for Codex

Owner decision 2026-11-17 ("Закрити всі тестами"). Both re-review axes (`_wip-pr109-review2-codex-{standards,spec}.md`) block on these
rows. A static trace to an e2e row does not count. `asr.spec.ts` is excluded in CI, so each behavior needs a **plain deterministic test
that CI selects** (unit or integration), with executed baseline/mutant evidence.
Contract: `/Users/strelia/project-data/ai-factory-kit/skills/arch/code-review/references/behavior-coverage.md`.

- **Copy (the ONLY place you edit):** `{COPY}`, now at `3e7de7e` (= branch HEAD `ad3612c` + docs). Clean. Never touch
  `/Users/strelia/projects/courses`.
- **Rows:** {ROWS}. The exact mutation diff for each row is in `{LOGS}/_wip-pr109-mutations-{SRC}.md` (grep the id); the reviewers'
  reachability notes are in `docs/multi-subtitles/_wip-pr109-review2-codex-standards.md` (table "E2E_ONLY reachability") inside the copy.
- First re-apply each mutation against the **current** tests: some may already be killed by the 208 closure tests. If the whole suite
  fails attributably, record KILLED with the verbatim line and add no test.
- Otherwise add the missing discriminator at the seam the reviewers name. Examples:
  - H38–H40: route-level auth wiring through the real router or middleware, not a no-op auth;
  - Q05: task handler → Claim track id;
  - A14: regeneration marker;
  - S68: repository duration, integration;
  - FE-B: the CourseEditor/route composition prop chain via the real route composition.
  {EXTRA}
- **Edit tests only.** Production bytes must stay identical: `git -C {COPY} diff --name-only` shows only test files and test fixtures. If a
  correct test fails on unmutated code, mark the row DEFECT, keep the test out of the copy, and record the failure; that is STOP + QUESTION.
- Use no sleeps. Use independent expected values.

WORKING FILE: `{LOGS}/_wip-pr109-gaps-e2e-{PART}.md` (append only). Start it with a `NEXT:` line before the first tool, then write
FACT / DECISION / NEXT / QUESTION entries.

Record one row per item: `id | test file::name | baseline PASS | mutated FAIL line (verbatim, named test) | KILLED-by-existing / CLOSED / DEFECT`.
Restore production bytes after each row.

## End
{GATES}
Final: the counts, `git -C {COPY} diff --stat` (including untracked files), and the SHA-256 of each changed file.
