# Ryuk race in `make integration` — diagnosis (Courses 1, 2026-10-03)

Symptom: intermittent `reaper: from container "<id>": wait for reaper <id>: unexpected container status
"created"`, 3 times on 2026-10-03, always first `postgres.Run` in
`internal/discussion/adapters/postgres`; a plain rerun passes.

## Root cause — from testcontainers-go v0.44.0 source (module cache `courses-gomod`)
1. **One session for all packages.** `internal/core/bootstrap/bootstrap.go`: session id =
   `sha256(parentPid, parentCreateTime)`. Every test binary under one `go test ./...` has the same parent
   (the `go` command) → same session → one shared reaper, by design.
2. **Race on reuse.** `reaper.go:311` `reuseOrCreate`: «Look for an existing reaper created in the same test
   session but in a different test process execution e.g. when running tests in parallel». Process B finds
   the ryuk container that process A has *created but not yet started* and goes `fromContainer`.
3. **`created` is fatal.** `fromContainer` waits with `wait.ForLog("Started")` + `ForExposedPort`; the wait's
   `checkState` (`wait/wait.go:46-56`) returns `unexpected container status %q` for any status other than
   running / exited / OOMKilled — `created` is not treated as "not yet".
4. **No retry.** `reaper.go:230-249` `retryError` retries only conflict/connection/timeout errors; this one
   becomes `backoff.Permanent` → the test fails at once.
→ A race inside the library, not our code. `discussion` likely just loses the timing more often.

## Upgrade?
`proxy.golang.org` @latest = **v0.44.0** (2026-08-07) — we are on the newest; no released fix.

## Our cleanup without ryuk
All 15 test files that start containers call `Terminate`/`CleanupContainer` in cleanup (checked by grep).
Without ryuk, containers leak only when the test process is killed hard (`go test -timeout` panic,
Ctrl-C) — cleanups do not run then.

## Options (magnitudes)
- **A. `TESTCONTAINERS_RYUK_DISABLED=true`** in `infrastructure/tests/integration.sh` `docker run -e`:
  removes the shared reaper → no race. Time cost 0. Leak only on hard kill (document a cleanup command).
- **B. `-p 1`** (packages one at a time): no concurrent reuse → no race; ryuk kept. Time cost: the 45
  packages' test times sum to **74.0 s** (integration run 2026-10-03, `gate-t05e.log`) versus the parallel
  run's wall clock — roughly +50 s per integration run.
- C. Upgrade — not available (v0.44.0 is latest).

Upstream research by Courses 2: `work/research-ryuk.md`.
