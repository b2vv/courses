# Review: package `flakefix-ryuk-w1` (testcontainers reaper name-conflict retry)

Independent reviewer, read-only: no edits, no git writes, no Docker. You may run unit `go test ./internal/testsupport/containers/` and
`go vet` if your sandbox allows. Integration tests need Docker; do not run them.

- Context: `make check` attempt 10 (branch `ms2-3-asr-per-track` @ `972921d`) failed at `make integration`. `cmd/media-worker`
  `TestRegenerateMastersAfterLegacyRelabel` hit `create container: reaper: new reaper: run container: container create: Error response from
  daemon: Conflict. The container name "/reaper_<session>" is already in use`. The temporary retry in
  `backend/internal/testsupport/containers/start.go` handled only the `wait for reaper … unexpected container status "created"` shape.
- Delta: the uncommitted `git diff -- backend/internal/testsupport/containers` (frozen `docs/multi-subtitles/_wip-flakefix-ryuk.patch`).
  `isReaperRace` also accepts the name conflict; two unit tests are added. The author's mutation check: disabling the new clause fails
  `TestStartRetriesTheReaperNameConflictRace`.

Check against the installed testcontainers-go v0.44.0 (`$(go env GOMODCACHE)/github.com/testcontainers/testcontainers-go@v0.44.0`,
`reaper.go` `reuseOrCreate`, `lookupContainer`, `newReaper`, `retryLocked`; `docker.go` CreateContainer):
1. Is the mechanism right? Packages of one `go test` share a session id, both miss the lookup, and the loser gets a name conflict.
   Does a retry then reuse the winner's reaper, rather than loop or orphan anything?
2. Is the match narrow enough (only reaper name conflicts) and robust to the real message?
3. Is the returned target container handled as in the existing race (terminated before retry)?
4. Do the guard tests and the TEMPORARY removal note still hold?
5. Is retrying the right fix, versus something else?

Finding format: `file:line`, impact, how verified, BLOCKER or non-blocking, confidence. End with `VERDICT (flakefix-ryuk-w1): SHIP | BLOCK`.
