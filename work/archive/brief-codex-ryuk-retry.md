# Codex brief — temporary retry for the reaper create→start race, then re-enable ryuk

Repository /Users/strelia/projects/courses, branch `feat/ratings`. Read first: `work/_wip-ryuk-race.md`
(diagnosis + the «Correction» section at the end), `work/research-ryuk.md` §7–§8,
`infrastructure/tests/integration.sh` (lines ~238–250, it currently disables ryuk). testcontainers-go
v0.44.0 source is in your Go module cache. Log each step to `work/codex-ryuk-retry.log` (first line before
the first tool: `NEXT:`). Do **not** commit. Write the tests first, and do not stop before every item is done.

## Why
The observed failure is **transient**: a package process finds a sibling package's reaper between
ContainerCreate and start; `fromContainer` (`reaper.go:335`) waits on logs, `checkTarget` sees `created` on
the first poll (`wait/log.go:170`, `wait/wait.go:55`) and `retryError` (`reaper.go:230`) makes it permanent.
A moment later that reaper is running, so retrying our container start succeeds. Upstream fix PR #3868
(retry `created` inside `lookupContainer`) is unreleased. The owner wants the reaper back, with a temporary
retry that is removed when the fix ships.

## Build
1. **Package `backend/internal/testsupport/containers`** (test support only; standard library + nothing
   from our modules):
   ```go
   // TEMPORARY(testcontainers-go#3868): remove this package and its call sites when a testcontainers-go
   // release contains PR #3868; then run the integration gate three times.
   func Start[T any](ctx context.Context, start func(context.Context) (T, error)) (T, error)
   ```
   - Calls `start`; on success returns. Retries **only** when the error is the reaper race: the message
     contains both `wait for reaper` and `unexpected container status "created"`. Any other error is
     returned at once, unchanged.
   - Backoff: start ~200 ms, doubling, capped at 2 s, total budget 30 s (bounded by `ctx` too); then return
     the last error wrapped with a hint that names #3868.
   - Never removes or touches the reaper container itself.
   - Before writing it, confirm in the source that a failed reaper acquisition leaves **no target
     container** behind (i.e. the reaper is obtained before the target's `ContainerCreate` — e.g.
     `docker.go` around line 1013) and that the `reaperSpawner` state allows a fresh `reuseOrCreate` on
     the next call (on this error `fromContainer` returns a nil reaper, so nothing is terminated). Write what
     you found, with lines, into the package doc comment and the log. If either is false, stop and say so.
2. **Call sites:** wrap every testcontainers start in integration tests with `containers.Start` — the 15
   files under `backend/internal/**` that call `postgrescontainer.Run`, the redis module's `Run`,
   `testcontainers.GenericContainer` or `testcontainers.Run` (find them with grep; include
   `catalog_measure_test.go`). Keep each file's existing cleanup (`Terminate` / `CleanupContainer`) exactly
   as it is.
3. **Re-enable ryuk:** remove `-e TESTCONTAINERS_RYUK_DISABLED=true` from `infrastructure/tests/integration.sh`
   and replace the comment above `docker run` with a short one: ryuk is on; the reaper create→start race is
   handled by `internal/testsupport/containers.Start` — `TEMPORARY(testcontainers-go#3868)`, remove both when
   a release contains #3868. Keep the cleanup hint for hard kills.
4. **Backlog:** append to `work/backlog.md` an entry «B-27 — TEMPORARY retry for the testcontainers reaper
   race» with: what, why (links/issue numbers), the **trigger** (a testcontainers-go release containing
   #3868), and the removal steps (delete the package and call sites, run `make integration` three times).
   Follow the file's existing entry style.

## Tests (required)
`backend/internal/testsupport/containers/start_test.go`, no Docker: success first try; the race error twice
then success (3 calls); a non-race error returned at once (1 call); the budget exhausted → error mentions
#3868; a cancelled context stops retrying. Use a fake clock or tiny durations injected through an unexported
variable so the test runs in milliseconds.

## Verify
`go vet ./...`, `go test ./internal/testsupport/...`, `go build ./...` (integration tests compile with
`go test -tags=integration -run XXX ./...` — do this to prove every call site compiles). Docker steps
(`make integration`) are run by Courses 1. Final answer: files changed, the two source facts you confirmed
(with lines), tests added, what you ran, what you could not.
