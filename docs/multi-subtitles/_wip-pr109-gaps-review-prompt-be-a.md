# Review: package `gaps-be-a` (tests closing 53 mutation GAPs, backend transcript + migration 00031)

Independent reviewer, read-only: no edits, no git writes, no Docker. You may run `go vet ./internal/transcript/...` and unit
`go test ./internal/transcript/...` if your sandbox allows; integration tests need Docker — do not run them.

- Delta: uncommitted test files under `backend/internal/transcript/` in the working tree (frozen
  `docs/multi-subtitles/_wip-pr109-gaps-be-a.patch`, sha256 `c1fcb5d0…`). Branch `ms2-3-asr-per-track` @ `46c773b`. Other dirty
  files (including frontend test files of package `gaps-fe-a`, reviewed separately) are out of scope.
- Inputs: the GAP rows and proposed tests in the be-a mutation journal (outside the repo:
  `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/_wip-pr109-mutations-be-a.md`,
  large — grep row ids) and the closure journal `docs/multi-subtitles/_wip-pr109-gaps-be-a-journal.md` (per row: test, baseline PASS,
  mutated FAIL line). Contract: `/Users/strelia/project-data/ai-factory-kit/skills/arch/code-review/references/behavior-coverage.md`.

Check:
1. Only test files changed; no production change; no existing assertion weakened (2 removed lines).
2. Each closing test asserts observable behavior with independently specified expected values, not implementation echoes;
   concurrency tests use events/locks, not sleeps, and prove the race path is reached.
3. Sample at least 12 rows across HTTP handler, queue, whisper engine, application recognize/draft, domain, and postgres
   integration including migration 00031: does the named test reach the mutated path, so the quoted failure is attributable?
4. Flakiness: shared DB state, fixture leakage between subtests, time dependence; cleanup of what the tests create.
5. Standards: `AGENTS.md` Architecture/Verification, `.claude/standards.md`, Go test conventions in the repo. Judge the file names
   `gaps_test.go` / `gaps_integration_test.go` (they name the process, not the behavior) against repo conventions.

Finding format: `file:line`, impact, how verified, BLOCKER or non-blocking, confidence. End with `VERDICT (gaps-be-a): SHIP | BLOCK`
and counts.
