# PR #109 — close E2E_ONLY and INCONCLUSIVE mutations with plain tests (part be) — brief for Codex

Owner decision 2026-11-17 ("Закрити всі тестами"). Both re-review axes (`_wip-pr109-review2-codex-{standards,spec}.md`) block on these
rows. A static trace to an e2e row does not count. `asr.spec.ts` is excluded in CI, so each behavior needs a **plain deterministic test
that CI selects** (unit or integration), with executed baseline/mutant evidence.
Contract: `/Users/strelia/project-data/ai-factory-kit/skills/arch/code-review/references/behavior-coverage.md`.

- **Copy (the ONLY place you edit):** `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut-be`, now at `3e7de7e` (= branch HEAD `ad3612c` + docs). Clean. Never touch
  `/Users/strelia/projects/courses`.
- **Rows:** BE-A A14, R06, H11, H14, H16, H19, Q05, S68, H38, H39, H40. The exact mutation diff for each row is in `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/_wip-pr109-mutations-be-a.md` (grep the id); the reviewers'
  reachability notes are in `docs/multi-subtitles/_wip-pr109-review2-codex-standards.md` (table "E2E_ONLY reachability") inside the copy.
- First re-apply each mutation against the **current** tests: some may already be killed by the 208 closure tests. If the whole suite
  fails attributably, record KILLED with the verbatim line and add no test.
- Otherwise add the missing discriminator at the seam the reviewers name. Examples:
  - H38–H40: route-level auth wiring through the real router or middleware, not a no-op auth;
  - Q05: task handler → Claim track id;
  - A14: regeneration marker;
  - S68: repository duration, integration;
  - FE-B: the CourseEditor/route composition prop chain via the real route composition.
  
- **Edit tests only.** Production bytes must stay identical: `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut-be diff --name-only` shows only test files and test fixtures. If a
  correct test fails on unmutated code, mark the row DEFECT, keep the test out of the copy, and record the failure; that is STOP + QUESTION.
- Use no sleeps. Use independent expected values.

WORKING FILE: `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/_wip-pr109-gaps-e2e-be.md` (append only). Start it with a `NEXT:` line before the first tool, then write
FACT / DECISION / NEXT / QUESTION entries.

Record one row per item: `id | test file::name | baseline PASS | mutated FAIL line (verbatim, named test) | KILLED-by-existing / CLOSED / DEFECT`.
Restore production bytes after each row.

## End
In `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut-be/backend` (GOCACHE/GOTMPDIR/TMPDIR inside the copy, Kiri env for integration): `gofmt -l .` empty, `go vet ./...` and `go vet -tags integration ./...`, `go test ./...`, integration of the dependent packages — exit codes and counts.
Final: the counts, `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut-be diff --stat` (including untracked files), and the SHA-256 of each changed file.
