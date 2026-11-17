# Review: package `gaps-e2e` (plain tests closing the 24 E2E_ONLY mutations and X-019 of PR #109)

Independent reviewer, read-only: no edits, no git writes, no Docker. You may run `npm run typecheck`, `go vet` and unit `go test` if your
sandbox allows.

- Why: both re-review axes blocked on these rows (`docs/multi-subtitles/_wip-pr109-review2-codex-{standards,spec}.md`, findings F3 /
  Major 1–2). Brief: `docs/multi-subtitles/_wip-pr109-gaps-e2e-brief-{fe,be}.md`.
- Delta: uncommitted test files on branch `ms2-3-asr-per-track` @ `3e7de7e` (= `ad3612c` + docs). Frontend:
  `admin-audio-queries.test.ts`, `TranscriptEditor.test.tsx`, `route-guards.test.tsx` (patch `_wip-pr109-gaps-e2e-fe.patch`, `7db55481…`).
  Backend under `backend/internal/transcript/`: asynq `queue_test.go`, http `draft_handler_test.go` + new `route_auth_test.go`, postgres
  `draft_integration_test.go` + `draft_track_integration_test.go`, application `inventory_regenerate_test.go` + `recognize_test.go`
  (patch `_wip-pr109-gaps-e2e-be.patch`, `f027ee1d…`). Other dirty files are out of scope.
- Evidence: `docs/multi-subtitles/_wip-pr109-gaps-e2e-{fe,be}-journal.md`. Per row: KILLED-by-existing or CLOSED, with a verbatim
  failing named test. FE: 6 KILLED-by-existing, 8 CLOSED including X-019. BE: 11 CLOSED. The original mutation diffs are in the mutation
  journals outside the repo:
  `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/_wip-pr109-mutations-{fe-a,fe-b,be-a}.md`.

Check, for **all 25 rows**:
1. Is the detector CI-selected? The FE unit suite runs under `make test`, and integration under `make integration`; `asr.spec` does not count.
2. Is the failure attributable to the mutation, with a named test and not a collection error? For X-019, is it a named assertion and not
   only an unhandled rejection?
3. Do H38–H40 exercise the real route and auth wiring rather than a no-op? Does the "existing kill" claim hold for the 6 FE rows?
4. Only test files changed. The removed lines in `draft_handler_test.go` (6) must not weaken any assertion.
5. Test quality: observable behavior, no sleeps, no leaks; standards.

End with `VERDICT (gaps-e2e): SHIP | BLOCK`, the counts, and `rows accepted N/25`.
