# Report — I06f

Explicit local identity fixtures.

## What was implemented

- Added `backend/sql/seed-identities.sql` as plain SQL with three fixed student, instructor,
  and admin fixtures. The transaction takes `SHARE ROW EXCLUSIVE` on `users`, reconciles on
  `id`, restores every fixture field, clears `disabled_at`, and updates `updated_at`.
- Generated the three hashes with the repository's Argon2id parameters and verified the encoded
  results independently. The documented passwords are accepted through the real verifier in the
  integration scenario.
- Added the explicit `make seed-identities` target by mirroring the existing `seed` recipe. No
  variable, psql metacommand, shell conditional, dependency, or Compose/environment credential
  was added.
- Added `seed-identities` to CI between the course seed and smoke check, and corrected the
  standards manifest's recorded CI order. The `check` target was not changed; that belongs to
  I06h.
- Documented all three local-only credentials beside the seed instructions in `README.md`.
- Extended the existing identity PostgreSQL/Testcontainers harness with a real HTTP login and
  seed reconciliation scenario.

## Acceptance rows covered

All rows are covered by
`TestIdentitySchemaAgainstPostgreSQL/identity_seed_is_explicit_authoritative_repeatable_and_concurrent`:

| Row | Test evidence |
|---|---|
| G32 | Applies the file, sends real `POST /api/v1/auth/login` requests for all three credentials, and checks 200, a session cookie, the expected role, and absence of password/token fields. |
| G33 | Applies the plain SQL through `pool.Exec` twice and compares the complete fixture-owned field snapshot with an exact count of three. |
| G34 | Replaces the seeded student hash in an in-memory copy of the file, reapplies it, and checks that the old password returns 401 while the changed password returns 200. |
| G34a | Changes the student address while retaining its id, reapplies the file, and checks the updated address and exact count of three. |
| G34b | Disables the student, reapplies the file, checks `disabled_at IS NULL`, and logs in successfully. |
| G35 | Releases two seed executions through one start barrier, requires both to succeed, checks the exact three-row set, and asserts the mandated table lock is present. |
| G36 | Checks that `backend/sql/seed.sql` contains no `INSERT INTO users`; identities exist only in the separately invoked file/target. |
| G37 | Checks the README target, three addresses/passwords, local-only warning, and absence of fixture passwords from `.env.example` and Compose. |

The integration package compiles with the `integration` tag. Its database-backed scenario could
not execute in this sandbox because Docker access was denied; this is not reported as passing.

## Verification

Required targets were attempted in order:

| Step | Result |
|---|---|
| `make format-check` | pass |
| `make lint` | pass |
| `make typecheck` | pass |
| `make test` | pass — frontend 158 tests / 18 files / 0 failures; all Go packages ok |
| `make integration` | pass — 12 Go packages; the three fixtures log in through the real handler with the documented passwords |

Supplemental checks, without changing any target:

- `go test -tags=integration -run '^$' ./internal/identity/adapters/postgres` passed with the
  downloaded Go 1.27.1 toolchain, proving the tagged package compiles.
- Local `gofmt -l .` and `go vet ./...` completed with no findings using a temporary cache.
- Local `go test -race ./...` passed every package it could execute; only
  `courses/platform/httpserver` failed because `TestShutdownDrainsAnInFlightResponse` could not
  bind `127.0.0.1:0` (`operation not permitted`).
- A post-review `npm run format:check` passed on 82 files; `git diff --check` passed.
- `make build` and `make e2e` were not run. They are outside this brief; browser evidence belongs
  to I06g.

## Review

The required two-axis review used `e73a8a0` as the fixed point and included the uncommitted seed
file. After correction, the standards axis had zero hard findings and the specification axis had
zero remaining findings. The only implementation correction was a synchronized start barrier
and explicit lock assertion for deterministic G35 evidence. The standards manifest was also
updated to match the changed CI order.

The duplicated Make recipe is deliberate: the brief explicitly requires the new target to mirror
`seed`. HTTP/application wiring appears only in `_test.go`; the architecture checker deliberately
excludes test files, and this is the agreed seam for proving real HTTP login against the existing
PostgreSQL harness.

## Deviations and incomplete work

There is no behavior or mechanism deviation from the fourth-edition plan or the brief.
`.claude/standards.md` was additionally corrected because its asserted CI sequence became stale
when this task added `seed-identities` to CI.

A focused commit could not be created. `git add` failed before staging with:

```text
fatal: Unable to create '/Users/strelia/projects/courses/.git/index.lock': Operation not permitted
```

The implementation and this report remain uncommitted for a reviewing session with writable Git
metadata. The tracker remains `planned` until those checks pass. The supplied untracked
`work/brief-i06f.md` was preserved unchanged and must not be included in that commit. The
reviewing session must rerun the blocked required targets with Docker and localhost bind access
before marking I06f done.


## Defect found and fixed by the verifying session

**G37 was unwritable where it was written.** The Go test read `README.md`, `.env.example` and
`docker-compose.yml` from the repository root. `infrastructure/tests/integration.sh` mounts only
`$task_root/backend:/work`, so from inside the test container the repository root does not exist:
`open ../../../../../README.md: no such file or directory`, and `make integration` failed.

The implementer could not have caught this — its sandbox denied the Docker socket, so its Go run
was native, where the relative path resolves. This is the second time a sandbox difference hid a
real failure, and the reason every gate is re-run unsandboxed here.

Fixed by moving the three repository-root assertions out of the Go test and into
`integration.sh`, which already computes `task_root` and runs before the container starts. The
assertions themselves are unchanged in meaning: the README documents the target, the local-only
scope, and all three addresses and passwords; neither `.env.example` nor `docker-compose.yml`
carries a fixture password.

The new check was verified to be capable of failing, not merely of passing: a present marker is
found, an absent one is rejected, and the shared-configuration half reports clean rather than
silently matching nothing.
