# Report M04 — stopped on missing domain and port contracts

## Outcome

M04 implementation stopped before creating `application/upload.go`,
`application/transcode.go`, or their tests. The brief says: “If a domain rule or port method
is missing, STOP and report rather than adding it here.” The M03 surface cannot express the
required application behavior without changing the domain or ports, which this slice does not
own.

Only this report was added. `work/tasks.md`, Git, and every `generated/` directory were left
untouched.

## Blocking contract gaps

1. **The first-PATCH free-space gate cannot preserve `created`.** D-15 requires checking
   `MediaStorage.Free` against `3 × declared_bytes` before T2 and leaving the row unchanged in
   `created` on insufficient space. The application has no read method for obtaining the
   asset's `declared_bytes`, while `AssetRepository.BeginUpload(ctx, id, now)` performs the
   state operation without a policy callback. Calling `BeginUpload` first would commit T2 too
   early; calling `Free` first lacks the required byte count. The 4b callback-into-repository
   idiom named by the brief cannot be applied through the current port.

2. **Sweep cannot delete bytes before deleting the row.** D-7 requires, for C1–C4, purging
   the four storage-key families and only then deleting the repository row, with a retryable
   row remaining if byte deletion fails. `AssetRepository` exposes `SweepCandidates`, but no
   method or callback for deleting a successfully purged candidate. Deleting inside
   `SweepCandidates` would reverse the required ordering and could orphan bytes.

3. **Publish cannot enforce the claim-token-before-bytes invariant at the application seam.**
   D-8/D-10 require the repository transaction to lock and validate `(processing,
   claim_token)` before any stale HLS destination is removed or the working directory is
   renamed. `AssetRepository.Publish` has no storage callback, while `MediaStorage.Publish`
   is a separate call. Either call order violates the required atomic sequence. This is the
   same repository-policy callback shape the brief explicitly directs M04 to mirror.

4. **The twelve-step claim flow has no load/re-read operation.** D-10 steps 1 and 2 require
   loading before claim and re-reading after a zero-row claim to apply the exhaustive nil list
   and distinguish an unclaimable `processing` row. `AssetRepository` exposes no load/read
   method and `Claim` returns no explicit claimed/not-claimed result. Inferring claim success
   from returned fields would be an undocumented port contract and still would not implement
   the required pre-claim load step exactly.

5. **The first-chunk container refusal has no domain rule.** T2/T8 and D-8 require the
   application to recognize accepted MP4/MOV `ftyp` brands or Matroska EBML magic before any
   byte is written, failing the `created` row with `unsupported_source` otherwise. The domain
   exposes codec-level `AcceptSource(SourceProbe)` only; it has no first-chunk/container
   policy. Implementing that business rule inside M04 would violate the brief's ownership
   boundary.

6. **R15's no-retry outcome has no inward retry-disposition vocabulary.** R15 requires the
   job error to wrap `asynq.SkipRetry`, but M04 is also required to import only domain and
   ports. Neither contains an equivalent permanent-job marker for a later Asynq adapter to
   translate. Importing Asynq in application would violate the module boundary; silently
   moving the assertion to M08 would not execute this brief exactly.

These are contract-shape blockers, not implementation choices. No workaround was attempted.

## Read-through performed

- Read the brief's sources in its specified order: `docs/media/plan.md` §2 application rows,
  §1, D-10, D-7, D-15, §8, and R14–R19/R31/R32/R33/R36/R42; then both 4b administration
  files, both 4a authoring files, and all M03 media domain and port files.
- Cross-checked D-7/D-8/D-10/D-15 against the exact M03 method signatures and the M03 report.
- Confirmed no existing media application helper supplies any of the missing contracts.

## Verification performed

- Direct scope read-through: no application, adapter, transport, frontend, dependency, SQL,
  or generated file was added or changed.
- `gofmt` was not run because the STOP condition fired before any Go file was created or
  edited; there was no Go formatting target.
- No filesystem write was denied. Therefore there is no denied-write error text to report.
- No Git command was run.

## Checks not run — denied by the brief or inapplicable after STOP

None of these checks is represented as passed:

- `make format-check` — not run; its Go half uses Docker. Local `gofmt` was inapplicable
  because no Go file changed.
- `make lint` — not run; its Go-vet half compiles through Docker.
- `make typecheck` — not run; Go compilation is denied.
- `make test` — not run; Go compilation and test execution are denied, and no M04 tests were
  created after the mandatory STOP.
- `make integration` — not run; Docker/Testcontainers are denied and adapters are outside
  M04.
- `make check-generated` — not run; it invokes Git and writes generated output, both denied.
- `make build` — not run; it requires Docker and compilation.
- `make e2e` — not run; it requires Docker Compose and M04 owns no browser surface.
- Aggregate `make check` — not run because it includes the denied checks above.

## Required predecessor repair

Revise the M03 domain/port contract in its own slice, then rerun M04. At minimum, the revised
surface must make the first-PATCH policy atomic, provide post-purge row deletion, put storage
publication behind the locked claim check, make claim disposition explicit, define the
container-sniff rule, and define an inward permanent-job outcome that the Asynq adapter can
map to `SkipRetry`.
