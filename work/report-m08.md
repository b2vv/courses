# Report M08 — blocked before implementation

## Outcome

M08 is blocked by a missing contract required by the read-first plan. No ffmpeg adapter or
adapter test remains in the repository.

The blocking requirement is `docs/media/plan.md` D-11:

> Thumbnail, a second short invocation: `-ss <1 s, or 0 if the source is shorter> ...`

The existing inward contracts cannot supply the fact needed to choose that fixed argument:

- `domain.SourceProbe` contains only `Width`, `Height`, `VideoCodec`, and `AudioCodec`.
- `ports.Transcoder.Thumbnail` accepts only the context, source key, and destination key.
- `application.Transcode.Process` does not pass its probe to `Thumbnail`.

Always emitting `-ss 1` would fail the stated sub-second-source branch. Always emitting
`-ss 0`, retaining mutable probe state in the adapter, or running an extra private probe from
`Thumbnail` would be a workaround rather than implementation of the existing contract. The
brief says: “Do not modify the domain, ports or application — if a contract is missing, STOP
and report.” Therefore implementation stopped. A contract-owning slice must first decide how
the thumbnail call receives duration (for example, a duration-bearing probe passed into the
call); this report does not choose or implement that change.

## Second specification conflict

`work/brief-m08.md` says cancellation “kills the process group.” The recorded cancellation
decision in `docs/media/plan.md` D-10 says the opposite: use `exec.CommandContext`'s default
direct-process `SIGKILL` plus `WaitDelay`, with “No `Setpgid` and no process-group signalling.”
R20 likewise observes that `/bin/sleep` exits within `WaitDelay`; it does not specify a process
group. This must be reconciled before an exact cancellation test can be claimed. No
process-group workaround was added.

## Read-first and scope audit

The prescribed sources were read in order: the required D-11, D-8, verified-facts, port, and
acceptance-row material in `docs/media/plan.md`; ADR 009; the existing transcoder port and
source domain; then `work/PRODUCT-BRIEF.md` section 11's no-user-argument rule. The existing
three-method `Transcoder` interface is present, but it is insufficient for the D-11 thumbnail
branch above.

An implementation and test draft was created during the red/green pass, then removed when the
mandatory specification review exposed the contract blocker. No domain, port, application,
worker, PostgreSQL, HTTP, tus, frontend, dependency, or generated file was changed.
`work/tasks.md` and `docs/media/report.md` were not edited. No file inside a `generated/`
directory was written. No test requiring ffmpeg and no skip-if-missing guard was added.

## Review

The implementation workflow's two-axis review was adapted to the explicit M08 files because
Git is forbidden.

- Standards axis: the discarded draft had no documented-standard violation. Literal expected
  argv fixtures duplicated fixed sections by design so the expected value was independent of
  the implementation.
- Specification axis: found the blocking thumbnail contract gap above. It also confirmed the
  brief/plan cancellation conflict and noted that a robust R20 test must prove `/bin/sleep`
  started before cancellation, not merely sleep for a guessed interval.

The review caused the required stop; its draft findings were not shipped as partial code.

## Verification performed

- Read-through compared the requested behavior with the exact existing `SourceProbe` and
  `Transcoder` signatures and established that the required duration cannot cross either
  boundary.
- `gofmt -w` and `gofmt -d` ran successfully on the discarded Go draft before the blocker was
  found. They are not claimed as evidence for an implementation because those files were
  removed.
- No filesystem write was denied, so there is no denied-write error text to report.
- No Git command was run and no commit was created.

## Checks not run — denied by the brief

The exact environment denial is: **“No Docker: no Go compile or test. gofmt + read-through.”**
The following checks are unrun and none is represented as passed:

- focused `go test ./internal/media/adapters/ffmpeg` — not run; Go test execution is denied,
  and the adapter was removed after the blocker;
- direct Go compilation, `go test`, `go vet`, and `go build` — not run; compilation and tests
  are denied;
- `make format-check` — not run; its Go path uses Docker;
- `make lint` — not run; its Go vet path compiles through Docker;
- `make typecheck` — not run; Go compilation is denied;
- `make test` — not run; Go compilation and tests are denied;
- `make integration` — not run; Docker is denied and it is outside M08's unit-test evidence;
- `make check-generated` — not run; it performs Git-backed drift checking and Git is denied;
- `make build` — not run; Docker and Go compilation are denied;
- `make up`, `make seed-identities`, `make seed`, and `make smoke` — not run; Docker Compose is
  denied and these checks are outside M08;
- `make e2e` — not run; Docker Compose is denied and it is outside M08;
- aggregate `make check` — not run because it contains the denied checks above;
- ffmpeg and ffprobe execution — not run; the brief requires unit evidence without ffmpeg,
  and the only permitted process-behavior fixture would be `/bin/sleep` after the contract and
  cancellation wording are corrected.

## Required next decision

Revise the inward thumbnail contract so D-11's seek choice is expressible, and reconcile the
brief's process-group phrase with D-10's accepted no-process-group decision. M08 can then be
implemented without changing scope or inventing an adapter-side workaround.
