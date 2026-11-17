# PR #109 mutation GAP closure — brief for Codex (part be-a)

Owner decision 2026-11-17 ("Закрити всі тестами"): every SURVIVED(GAP) row of part be-a is closed with a behavior test.
Contract: `/Users/strelia/project-data/ai-factory-kit/skills/arch/code-review/references/behavior-coverage.md` ("close an actual
gap with a behavior test at an agreed seam … check that the test reaches the path … rerun baseline and mutation after repair").

- **Copy (the ONLY place you edit):** `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut-be` at `5a56965` (clean). Never touch `/Users/strelia/projects/courses`.
- **Input:** the GAP rows and their proposed tests in `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/_wip-pr109-mutations-be-a.md` (your/the previous run's journal).
- **Edit tests only.** Production files must stay byte-identical; check with
  `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut-be diff --name-only` at the end: only `*.test.ts(x)`, `*_test.go`, or test fixtures/helpers under test-only paths.
  If a correct behavior test fails on the UNMUTATED code, that is a real defect: do not change production code; mark the row
  `DEFECT`, keep the failing test aside (do not leave it in the copy), and record the verbatim failure. That is the STOP + QUESTION channel.
- Prefer one test per behavior that discriminates several related rows. Use the seams named in the rows. No sleeps for races; use events or deferred promises (see `race-test-event-driven-not-sleep`).
  Independently specified expected values; do not copy implementation logic into assertions.

WORKING FILE: `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/_wip-pr109-gaps-be-a.md` (append only). First entry before the first tool: a `NEXT:` line; then
FACT / DECISION / NEXT / QUESTION.

## Per GAP row
1. Write/extend the test. 2. Unmutated: the new test passes. 3. Re-apply exactly that row's mutation: the new test fails —
quote the failing assertion line verbatim. 4. Restore; `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut-be diff --quiet -- <production file>`.
Row record: `id | test file::test name | baseline PASS | mutated FAIL line | CLOSED | DEFECT`.

## End of part
In `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut-be/backend`: `gofmt -l .` empty, `go vet ./...`, `go test ./...`, and the integration tests of the dependent package set (Kiri env, as in the mutation run; GOCACHE/TMPDIR inside the copy) — exit codes and counts.
Final: counts CLOSED / DEFECT / remaining, `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut-be diff --stat`, SHA-256 of each changed test file. If you approach
your time limit, stop at a row boundary with production bytes restored and write `NEXT: resume at <id>`.
