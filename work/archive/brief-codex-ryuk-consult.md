# Consultation (read-only) — keep the testcontainers reaper but remove the race?

You are asked for a design opinion, not code changes. Sandbox is read-only: do not edit files. Answer in
your final message.

Repository /Users/strelia/projects/courses (Go backend in `backend/`). Read first:
- `work/_wip-ryuk-race.md` — our diagnosis from the testcontainers-go v0.44.0 source;
- `work/research-ryuk.md` — upstream issues #3918, #3867, unreleased fix PR #3868;
- `infrastructure/tests/integration.sh` — how integration runs (`docker run golang … go test -count=1
  -tags=integration ./...` with the host Docker socket; macOS Docker Desktop); it now sets
  `TESTCONTAINERS_RYUK_DISABLED=true` as the current fix;
- the 15 integration test files that start containers (e.g.
  `backend/internal/discussion/adapters/postgres/repository_integration_test.go:31-60`).
The testcontainers-go source is in the Go module cache (`go env GOMODCACHE` →
`github.com/testcontainers/testcontainers-go@v0.44.0`: `reaper.go`, `internal/core/bootstrap/bootstrap.go`,
`wait/wait.go`, `internal/config`).

The owner asks: instead of disabling ryuk, is there a robust solution — e.g. a wrapper with our own context
and wait/retry rules (Go `defer` etc.) — so the race cannot happen while the reaper keeps guarding hard
kills? Evaluate concretely, citing source lines:
1. Can our own context (longer deadline, detached from the test's) passed to `postgres.Run` prevent the
   #3918 orphan? Which call actually receives it?
2. If an orphan reaper (`reaper_<sessionId>`, status "created") already exists, can a retry in the same
   process ever succeed, or does `lookupContainer` keep returning it? Is there a supported way to get a
   fresh session or force-remove that orphan (config/env such as a session id override, `ryuk.*`
   properties, `TESTCONTAINERS_RYUK_*`), and is it safe with parallel package binaries?
3. Options like: one shared Postgres per package via `TestMain`; a pre-step that starts the reaper once
   before the parallel run (does a separate `go test` process even share the session?); `-p 1`;
   `RYUK_RECONNECTION_TIMEOUT`; a vendored patch of #3868 via `replace` in go.mod.
4. Your recommendation versus the current `TESTCONTAINERS_RYUK_DISABLED=true`, with the trade-off in one
   paragraph, and — if you recommend a wrapper — its exact shape (where it lives under
   `backend/internal/testsupport/…`, signature, what it retries, what it never retries).
Facts with file:line; mark guesses as guesses. If something cannot be done safely, say so plainly.
