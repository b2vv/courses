# Report M09 — media API surface

## Outcome

M09 implements the three JSON operations from `docs/media/plan.md` §2:

- `POST /api/v1/admin/lessons/{lessonId}/video` mints a metadata-only upload ticket;
- `GET /api/v1/admin/courses/{courseId}/videos` returns one DTO per non-evicted asset;
- `POST /api/v1/admin/videos/{videoId}/retry` performs the application retry and enqueues
  the resulting attempt.

All three reuse the exported identity `Authenticate` middleware. The two mutations place
`RequireSameOrigin` outside authentication, then apply the media staff-role gate before path,
query, body, ownership, or repository work. Responses carry `Cache-Control: no-store`.

The accepted §18 construction amendment is implemented without changing behavior: the JSON
operations now live on an application `Tickets` type holding only the PostgreSQL repository,
queue, and clock, while byte-handling methods remain on `Upload`, which holds storage. No
storage adapter or media mount was added to `api`; its composition has no nil dependency.
`work/tasks.md`, `docs/media/report.md`, and every `generated/` directory were left untouched.

## Changed

- `backend/internal/media/adapters/http/handler.go` adds `RegisterAdmin`, the three handlers,
  explicit transport-to-media actor mapping, UUID/query validation, an 8 KiB strict mint
  decoder, DTO mapping, fixed public errors, and no-store headers.
- `backend/internal/media/adapters/http/handler_test.go` adds kin-openapi request/response
  contracts for P01, P02, P07, B01, B04, B15, R01–R09, R22, R37, and R39. The tests exercise
  the real M04 application use cases over fake repository/queue ports and the real identity
  middleware.
- `backend/internal/media/adapters/asynq/queue.go` supplies the producer required by the retry
  operation: payload `{assetId,attempt}`, task id `assetId:attempt`, queue `media`, three
  retries, 30-minute timeout, 24-hour retention, duplicate-task success, and safe dependency
  failure mapping.
- `backend/internal/media/adapters/asynq/queue_test.go` pins that payload, all five D-10 queue
  options, duplicate-task success, and redaction of Redis failure detail over a fake client.
- `backend/internal/media/application/upload.go` implements accepted amendment §18.2 by moving
  `MintTicket`, `ListCourseVideos`, and `Retry` to the narrow `Tickets` application type over
  repository + queue; storage-backed chunk/complete behavior remains on `Upload`.
- `backend/internal/media/application/upload_test.go` updates the fake-port constructions to
  exercise ticket/list/retry through `Tickets` and chunk/complete/cancel through `Upload`, with
  the same assertions and behavior.
- `backend/cmd/api/main.go` constructs the media repository, queue, application, and handler,
  then registers the handler on the 4a authorization seam. It constructs `Tickets` directly;
  no nil or `MediaStorage` slot exists in the API composition.
- `backend/openapi/openapi.yaml` adds the three paths, media request/response schemas, response
  status sets, no-store headers, and the six §8 error codes. Generated Orval output was not
  written; the brief assigns regeneration to the reviewer and forbids it in this environment.

## Discussion

### R22 amendment accepted — recorded in plan §18.1

The original R22 precondition named a storage failure even though mint and the course-video
status read do not call `MediaStorage`, and D-15 forbids storage in the API. This was accepted
as a plan inaccuracy. `docs/media/plan.md` §18.1 now amends R22's precondition to a dependency
failure carrying filesystem-looking private detail, forced through the repository/application
seam.

The existing M09 test is retained as written: it forces a wrapped `dependency_unavailable`
containing `/media`, a filename, and ffmpeg text through the repository/application seam for
both operations, and proves the fixed 503 body discloses none of it. No storage adapter was
added to the API.

### Narrow application construction accepted — recorded in plan §18.2

The original shared constructor forced `cmd/api` to provide a storage slot even though its
surface must not handle media bytes. The gate owner accepted the issue and lifted M09's
application NOT-list only for this correction in `docs/media/plan.md` §18.2. Construction is
now split: `Tickets` owns mint/list/retry over repository + queue, while `Upload` retains
chunk/complete and the storage dependency. `cmd/api` constructs `Tickets` with no nil anywhere,
so it cannot receive or expose the byte-handling methods. Fake-port tests were updated without
changing their behavior.

### Queue producer ownership accepted — recorded in plan §18.3

The M09 task requires a functioning retry endpoint and `cmd/api` wiring, but no concrete
`TranscodingQueue` existed. Passing `nil` would commit T9 and then panic at
the application queue call. The plan names
the producer and its exact D-10 options in `adapters/asynq/queue.go`, but the task table does
not originally assign that file to a specific slice. The gate owner accepted this outcome in
`docs/media/plan.md` §18.3: M09 owns the producer; worker consumer wiring remains M12-owned and
untouched.

## Review

The implementation workflow's code-review gate was adapted to the explicit files because Git
is forbidden. Standards and specification were reviewed independently by read-only inspection.

### Standards

No documented-standard violation was found after the accepted §18.2 correction. The review
confirmed the route middleware order, inward dependency direction, absence of API storage or
nil dependencies, structural separation from byte-handling methods, and absence of generated
writes.
Two non-blocking smell judgements were retained: `serveMedia` has a long positional test-helper
signature, but introducing a request-fixture abstraction for one test file would add machinery;
and `mediaHeaders` is broad for a one-header function, but deliberately matches the 4a template's
`adminHeaders` naming.

### Specification

The first pass found two partial assertions: R08 checked only the last queue call, so extra
enqueues could pass, and R37 did not verify the incremented attempt/task attempt. Both were
corrected to require exactly one new enqueue and its exact `{assetID, attempt}`.

The follow-up found no remaining blocking implementation issue. The producer then gained a
focused fake-client unit test for its payload, task ID, queue, retry, timeout, retention, duplicate
conflict, and dependency-error behavior. The dormant `unsupported_media` JSON error mapping is
kept because §8 asks for the six codes to be mapped once, although the actual first-chunk refusal
belongs to the later tus surface. R22 now follows its accepted §18.1 precondition.

Review summary: Standards — zero blockers and two non-blocking smell judgements. Specification —
zero remaining implementation blockers; all three Discussion outcomes are accepted in plan §18.

## Verification performed

- `gofmt -w` over the five Go files — exit 0 after correcting one missing parenthesis found
  by the formatter.
- `gofmt -d` over the five Go files — exit 0 with no output.
- After the §18.2 split, `gofmt -w` over the four affected Go files — exit 0.
- Final `gofmt -d` over those four files — exit 0 with no output.
- Static read-through checked all three route templates; same-origin-before-authentication on
  both mutations; role mapping; UUID/query guards before application access; strict member,
  type, UTF-8, trailing-content, and 8 KiB body handling; every §8 status mapping; DTO members;
  queue payload/options; and API composition over the narrow `Tickets` type without a storage
  adapter or nil dependency.
- Static acceptance inventory found every owned row ID in `handler_test.go`.
- No filesystem write was denied, so there is no denied-write error text to report.
- No Git command was run and no commit was created.

## Checks not run — denied by the brief

The exact denial is: **“No Docker: no generation, no compile, no tests. gofmt + read-through
against the existing generated symbols.”** None of the following is represented as passed:

- OpenAPI/Orval generation — not run; generation and writes inside `generated/` are denied;
- focused `go test ./internal/media/adapters/http` — not run; Go compilation and tests are
  denied, including every kin-openapi contract row;
- focused `go test ./internal/media/adapters/asynq` — not run; Go compilation and tests are
  denied, including the producer payload/options check;
- direct Go compilation, `go test`, `go vet`, and `go build` — not run;
- `make format-check` — not run; its Go path uses Docker;
- `make lint` — not run; its Go-vet path compiles through Docker;
- `make typecheck` — not run; Go compilation is denied;
- `make test` — not run; compilation and tests are denied;
- `make integration` — not run; Docker/Testcontainers are denied;
- `make check-generated` — not run; generation, generated writes, and Git-backed drift checks
  are denied;
- `make build`, `make up`, `make migrate`, `make seed`, `make seed-identities`, and
  `make smoke` — not run; Docker Compose is denied;
- `make e2e` — not run; Docker Compose is denied and M09 owns no browser surface;
- aggregate `make check` — not run because it includes the denied checks above.

## Limitations and next task

Compilation, OpenAPI validation, contract execution, generated-client drift, and runtime queue
behavior await the reviewing session. The next planned slice is M10, the tus protocol server;
M12 still owns worker consumer wiring and metrics, as reaffirmed by plan §18.3.

## Reviewer verification (measured in the reviewing session)

The `## Discussion` was accepted on all three points and recorded as plan **§18** (dated):
R22's precondition amended to a dependency failure through the repository/application seam
(its real subject was always that the 503 discloses no path, filename or ffmpeg text);
the construction split so the API **cannot** receive byte-handling methods — `Tickets`
over repository+queue, `Upload` keeping storage, and **zero `nil` in the media wiring of
`cmd/api`**, which turns D-15 from a convention a caller must respect into something the
type system enforces; and `adapters/asynq/queue.go` assigned to M09, with M12 still owning
the consumer.

Gates: `make generate` clean, `make test` pass (media HTTP contract suite 7.9s, all media
packages green), `make integration` 19 packages ok, format-check, lint, typecheck and
check-generated pass.

This was the first task under the two-channel protocol, and the value was in what it caught
that a stop would not have: all three items were things Codex could have silently worked
around with code that runs. One would have wired a storage adapter into the API to satisfy
an impossible acceptance row. Provenance: implemented with Codex (gpt-5.6-sol); reviewed and
verified here.
