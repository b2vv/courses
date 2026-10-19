# Ryuk retry follow-up — your stop was right; the precondition changes, proceed

Keep logging to `work/codex-ryuk-retry.log`. Do not commit.

You correctly found that the target is created before the reaper is acquired (`docker.go:1314-1323`, then
`connectReaper` at `:1360-1362`). Courses 1 read on: on that error the library **returns the created target
together with the error** — `docker.go:1360-1361` `return ctr, err // No wrap…`; `generic.go:136-137`
`return c, fmt.Errorf("generic container: %w", err)`; `postgres.go:168-182` wraps a non-nil `ctr` in
`*PostgresContainer` and returns it with `run postgres: %w`. So the caller receives the partially created
container and can remove it.

New precondition for the retry (replaces «no target container behind»): **before each retry, terminate the
container returned with the race error** using `testcontainers.TerminateContainer` (it handles nil and typed
nil). Confirm the propagation for every start kind you wrap (postgres module, redis module, GenericContainer /
testcontainers.Run) and cite the lines in the package doc and the log; if one kind does not return the
container with the error, stop and say which.

Then do the whole original brief (`work/brief-codex-ryuk-retry.md`) with this change:
- `Start` is constrained so it can terminate: `func Start[T testcontainers.Container](ctx context.Context,
  start func(context.Context) (T, error)) (T, error)` (the package may import testcontainers-go; it is test
  support). On the race error: `TerminateContainer(c)` (log, do not fail the retry, if termination errors),
  then back off and retry. On any other error: return `c, err` unchanged so the call site's existing cleanup
  still sees the container.
- Unit tests add: the returned container is terminated before each retry (fake container counting
  Terminate calls); on a non-race error the container is returned, not terminated.
Everything else as written: all call sites, re-enable ryuk with the short `TEMPORARY(testcontainers-go#3868)`
comment, backlog B-27 with the trigger and removal steps, verification commands. Do not stop before all
items are done.
