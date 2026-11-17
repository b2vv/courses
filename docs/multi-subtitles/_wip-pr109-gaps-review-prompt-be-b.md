# Review: package `gaps-be-b` (tests closing 21 mutation GAPs, backend media/config/cmd + Makefile e2e-reset)

Independent reviewer, read-only: no edits, no git writes, no Docker. You may run `go vet` and unit `go test` if your sandbox allows.

- Delta: uncommitted files: `backend/cmd/api/main_test.go`, `backend/cmd/media-worker/main_test.go`,
  `backend/internal/media/adapters/postgres/{audio_track_metadata_integration_test.go,repository_integration_test.go}`,
  `backend/internal/testsupport/startup/wiring_integration_test.go`, `frontend/src/adapters/api/e2e-reset.test.ts`,
  `frontend/src/adapters/api/test-fixtures/reset-docker.mjs`. Frozen patch `docs/multi-subtitles/_wip-pr109-gaps-be-b.patch` (sha256 `b9ffe939…`).
  Note: after the closure run the author renamed `gaps_integration_test.go` → `audio_track_metadata_integration_test.go` and its `Gap`
  identifiers/subtests (content otherwise identical; integration rerun ok on Kiri) — the patch already holds the renamed form.
  Branch @ `16489a5`. Frontend test files of package `gaps-fe-b` are reviewed separately; other dirty files are out of scope.
- Inputs: GAP rows in `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/_wip-pr109-mutations-be-b.md` (grep row ids), closure journal
  `docs/multi-subtitles/_wip-pr109-gaps-be-b-journal.md`. Contract: `behavior-coverage.md` as in the previous packages.

Check: 1) only test/fixture files, no production change; 2) observable behavior, independent expectations, lock waits observed, no sleeps;
3) all 21 rows: the named test reaches the mutated path, failure attributable; 4) placement: `e2e-reset` (a Makefile target) is tested from
the frontend unit suite (`src/adapters/api`, node env) via a fake docker — is that the right home and does it run in `make test`; does it
spawn real `make`/`docker` anywhere or touch real volumes? A new package `internal/testsupport/startup` holding only an integration test —
convention fit? `cmd/*/main_test.go` — do they start real servers/ports? 5) flakiness, cleanup, standards (`AGENTS.md`, `.claude/standards.md`).

Finding format as before. End with `VERDICT (gaps-be-b): SHIP | BLOCK` and counts.
