# Review — d5b8726: Ryuk back on with a TEMPORARY retry (testsupport/containers.Start)

Reviewer: Courses 2. Date: 2026-10-03. Checked against: `work/research-ryuk.md` §7–§8 and the
testcontainers-go **v0.44.0 source** in the local module cache (not the docs).

## Verdict

**Approve. No blocking issues; 4 non-blocking.**

Both questions get a "safe" answer for the current call sites:
- matching on the error text cannot catch anything other than the reuse race;
- the retry can only terminate a container that this same call has just created.

The main gap is proof: the retry has never run against the real library (see №1).

What I ran myself: `go test ./internal/testsupport/containers/` → ok. I did not run
`make integration`.

## Q1 — Is matching on the error text safe across postgres / redis / generic starts? — Yes

- **The prefix appears exactly once.** `isReaperRace` needs **both** `wait for reaper` and
  `unexpected container status "created"`. In the whole v0.44.0 module (core + `modules/postgres`),
  `wait for reaper` appears exactly once: `reaper.go:354`, in `fromContainer`, on the **reuse** path.
- **Every container type goes through that same path.** Reaper acquisition happens inside
  `DockerProvider.CreateContainer` → `connectReaper` (`docker.go:1013-1019`, wrapped as
  `reaper: %w`). That is true for every container type: `postgres.Run`, `testcontainers.Run`,
  `GenericContainer` (redis/asynq/generic). The outer prefixes differ (`generic container: create
  container: reaper: from container …`), but both substrings survive the wrapping.
- **The target's own failure is not matched.** If the Postgres/Redis container itself fails its wait
  in state `created`, the error is `unexpected container status "created"` **without**
  `wait for reaper`, so it is not retried.
- **The creating process's own reaper failure is not matched either.** Those errors come from
  `newReaper`/`RunContainer`, not `fromContainer`, so they do not contain `wait for reaper`. That is
  correct: those are the #3918 origin, and retrying them would not help.
- **Coupling.** The match is tied to v0.44.0's wording. That is acceptable for a TEMPORARY shim with
  a named removal trigger (B-27). An upgrade either brings #3868 (and the shim goes) or changes the
  text (and the shim silently stops matching, which falls back to the old failure, not to anything
  worse). See №3.

## Q2 — Can a retry terminate something it must not? — No, for the current call sites

- **Only a target from this call.** On the race, `CreateContainer` returns `ctr, err` *after*
  `ContainerCreate` and *before* any start (`docker.go:1360-1361`). `GenericContainer` passes that
  `c` back (`generic.go:79-81`). `testcontainers.Run` preserves it (`generic.go:130-140`). So the
  terminated object is a **created-not-started target made by this very call**.
- **Never a shared container.** That holds because **no call site uses container reuse**
  (`Reuse`/`WithReuseByName`; grep over `internal/`: none). With reuse, `ReuseOrCreateContainer`
  could return a container shared with another test, and terminating it would be wrong (see №2).
- **Never a reaper.**
  - The shim terminates only the returned target.
  - On the failed reuse path the library returns a nil reaper.
  - `cleanupLocked` (`reaper.go:143-152`) terminates only `r.instance`, which is still nil (it is set
    only after success). So **the sibling's Ryuk is never touched**.
- **Typed nil is safe.** A typed-nil `*PostgresContainer` goes through `TerminateContainer` →
  `isNil` (`cleanup.go:100`), which is a no-op.
- **No double terminate.** After terminating, the shim returns the zero value. Call sites register
  `t.Cleanup(Terminate)` only after `err == nil`, so nothing is terminated twice.
- **Non-race errors are unchanged.** They return the original container and error, exactly as before
  the shim.

## Non-blocking

1. **The retry is proven only by fakes.** AGENTS says "a fake proves the fake". The 4 green runs never
   hit the race. A deterministic production-path test is cheap:
   1. Set `TESTCONTAINERS_SESSION_ID` to a unique value.
   2. Create (but do not start) a container named `reaper_<that id>`, using the real
      `testcontainers/ryuk:0.14.0` image and the three reaper labels.
   3. Start a goroutine that runs `docker start` on it after ~1 s.
   4. Call `containers.Start(postgres.Run…)`.

   Expected: the first attempt fails with the real library message, the shim terminates the target,
   and the next attempt reuses the now-running Ryuk. This proves the text match and the
   terminate→retry on the real path. Tag it `integration`, and remove it with B-27.
2. **Guard reuse.** Add one line in the package doc, and ideally a check: `Start` must not wrap a
   request with container reuse. A future `WithReuseByName` call site would turn "terminate the
   returned target" into "terminate a shared container".
3. **Pin the coupling.** Optionally, add a unit test that reads `go.mod` and fails when the
   testcontainers-go version is no longer `v0.44.0`, pointing at B-27. That makes the upgrade force a
   decision instead of silently disabling the match.
4. **The #3918 orphan costs time.** A permanently-`created` orphan matches the same text. Each package
   then spends the full 30 s budget (about 15 create→terminate cycles of its target) before failing
   with the "#3868" message. That is bounded and still fails loudly. It is just slower than before for
   that rare case; worth one sentence in B-27.

## Also fine

- **Budget and context.** Each attempt keeps the caller's context. The 30 s budget bounds only the
  loop, which avoids creating #3918 orphans through short deadlines.
- **Script.** `integration.sh` no longer disables Ryuk. The comment now points at the shim and B-27.
  The hard-kill leftovers note is kept, and Ryuk being on again means hard kills are reaped
  automatically.
- **B-27.** It records the trigger and the removal steps (delete the package and its call sites, then
  run `make integration` three times). That meets AGENTS' rule for decisions that close an
  alternative.
