# Report M06 — publish, eviction, recovery and sweep queries

## Outcome

M06 implements the media repository's publish-and-evict transaction, C1–C4 sweep queries,
state-agnostic `DeleteSwept`, and the C5/C6 recovery transaction. Every write uses the M05
`pool.Begin` idiom at PostgreSQL's default READ COMMITTED isolation; no course-module
`RepeatableRead` isolation was copied.

## Changed

- `backend/internal/media/adapters/postgres/publish.go` adds `Publish`, `SweepCandidates`,
  `DeleteSwept`, and `Recover`. `Publish` locks and verifies the processing claim before its
  policy can touch bytes, locks the lesson's non-evicted generations, publishes the row,
  stamps the lowest ready generation when a third exists, commits, and returns the evicted
  asset for the caller's post-commit purge.
- `backend/sql/queries/media.sql` adds sqlc-owned C1–C4 candidate selection, unconditional
  swept-row deletion, threshold-free C5 recovery, and ten-minute C6 recovery. C5 consumes a
  shared per-call budget before C6; C6 increments `attempt` and returns only rows requiring a
  fresh enqueue.
- `backend/internal/media/adapters/postgres/repository_integration_test.go` adds B06, B14,
  B24, and B29 coverage plus direct C1–C4, C6, shared-budget, and publish-policy rollback
  checks. The sweep checks use the real PostgreSQL repository through the existing application
  `Sweep` seam and a storage-boundary fake: all four key families precede deletion, and an HLS
  purge failure leaves the orphan row present and discoverable on the next pass.

No domain, port, application production code, migration, filesystem/ffmpeg adapter, worker,
HTTP/tus/frontend surface, or dependency was changed. The brief calls B06 “migration over a
populated DB”, but the plan's actual B06 row is third-generation eviction; populated migration
coverage is B07 and already belongs to M02. This slice follows the plan's actual B06 row.

## Read-through verification

- Publish order is target row lock → `(state='processing', claim_token)` check → active lesson
  generation locks → `PublishPolicy` → guarded ready update → lowest-generation eviction →
  commit. A stale claim returns an empty publish result and never invokes the policy.
- A policy error rolls back the row transition and any eviction. The application already owns
  post-commit cleanup; the repository returns the evicted asset rather than unlinking bytes in
  the transaction.
- Sweep predicates are exactly C1 evicted, C2 orphaned, C3 expired created/uploading, and C4
  failed for more than 30 days. `DeleteSwept` has no state predicate, so byte purge success is
  the caller's only prerequisite.
- Recovery runs C5 first with no age predicate, clears claims and refreshes `updated_at`, then
  spends only the remaining call budget on C6 rows older than ten minutes. C5 rows therefore do
  not fall through into C6 in the same transaction. Only C6 rows are returned for enqueue.
- SQL fixture writes remain one parameterised command per `Exec`; no multi-command extended-
  protocol fixture was added.
- The publish path uses the `video_assets_by_lesson` partial index shape; generation and pending
  invariants remain backed by the migration's two unique indexes. Sweep/recovery scans remain
  deliberate per plan D-7.
- Direct text checks found no `RepeatableRead` or `BeginTx` in the M06 implementation.

## Review

The required code review was adapted to the brief's Git denial and run over the explicit M06
file set on two independent axes.

- Standards: no hard breach. The remaining duplicated sqlc projections and generated-row
  mapping are non-blocking maintenance risk imposed by distinct generated row types. The
  one-statement delete transaction is retained because the brief explicitly requires the M05
  transaction idiom. A first-pass nil-embedded storage fake was replaced by an explicit fake
  whose unused methods fail visibly.
- Specification: the first pass found a double recovery limit and partial purge-order evidence.
  Both were fixed. The re-review found no remaining M06 specification issue or scope creep.

## Verification performed

- `gofmt -w` ran on the two changed Go files: exit 0.
- `gofmt -d` ran on the same files: exit 0 with no output.
- Read-through and targeted text checks covered transaction entry points, lock/claim order,
  C1–C6 predicates, shared recovery budget, purge/delete order, owned acceptance-row names,
  and forbidden isolation/scope strings.
- No filesystem write was denied, so there is no denied-write error text and the mandatory stop
  condition did not fire.
- No Git command was run. No commit was created because the brief explicitly denies Git,
  overriding the generic implementation workflow's commit step.
- `work/tasks.md` and `docs/media/report.md` were not edited. No file inside a `generated/`
  directory was written.

## Checks not run — denied by the brief

None of the following is represented as passed:

- `sqlc` / `make generate` — not run; sqlc is denied and generated-directory writes are
  forbidden. The new generated methods do not exist until a reviewer regenerates them.
- Go compilation — not run; explicitly denied.
- Testcontainers PostgreSQL — not run; Docker and Testcontainers are explicitly denied, so
  B06, B14, B24, B29, C1–C6, recovery-budget, and policy-rollback runtime evidence is pending.
- `make format-check` — not run; its Go path uses Docker. Local `gofmt` is the only formatting
  evidence.
- `make lint` — not run; its Go-vet path compiles through Docker.
- `make typecheck` — not run; Go compilation is denied.
- `make test` — not run; Go compilation and test execution are denied.
- `make integration` — not run; it requires Docker and Testcontainers.
- `make check-generated` — not run; it invokes generation/Git-backed drift checking, both
  denied in this session.
- `make build` — not run; it requires Docker and Go compilation.
- `make up`, `make migrate`, `make seed`, `make seed-identities`, and `make smoke` — not run;
  they require Docker Compose and are outside this repository-only slice.
- `make e2e` — not run; it requires Docker Compose and M06 owns no browser surface.
- Aggregate `make check` — not run because it includes the denied checks above.

## Limitation and next task

This slice has formatting and read-through evidence only. A reviewer must run sqlc before any
Go compile, then run the ordered gates in an environment with Docker/Testcontainers. Physical
directory deletion remains M07 filesystem-adapter evidence; M06 proves the boundary ordering
and retryable-row contract without implementing that forbidden scope. The next planned slice
is M07.

## Reviewer verification (measured in the reviewing session)

One review fix, and it was my briefing error rather than the implementer's: `Requeue` and
`Fail` fell between the M05 and M06 briefs — neither listed them — so the repository did
not satisfy `ports.AssetRepository` and the package failed to build. Both are guarded
UPDATEs of the same class as the methods already present (the claim token guards `Requeue`;
the from-state and the token both guard `Fail`, so a reclaimed or already-advanced asset is
never overwritten), so they were implemented here rather than spawning another task.

After the fix: `make integration` exit 0 —
`courses/internal/media/adapters/postgres` green in 7.4s with all four owned rows: B06 (a
third ready generation evicts and reaps the oldest), B14 (an orphan stays retryable until
its purge succeeds), B24 (recovery makes a stale publish lose its claim), B29 (recovery
reclaims processing claims of every age — the threshold-free C5 that closed the round-2
stranding finding). Verified by inspection: the claim is checked under `FOR UPDATE` before
the policy touches bytes, and `Purge` runs before `DeleteSwept`. format-check, lint,
typecheck and check-generated pass. Provenance: implemented with Codex (gpt-5.6-sol);
reviewed, corrected and verified here.
