# M14 report — frontend media adapter and application

## Files

- `frontend/src/domain/media/video.ts` — defines the bounded media model, upload ticket,
  fresh/resume transfer target, progress/result types, limits and the explicit
  `upload_resume_unverified` outcome.
- `frontend/src/ports/LessonVideoRepository.ts` and
  `frontend/src/application/admin/lesson-video.ts` — expose and validate mint, list, retry,
  `File` transfer and cancel use cases without transport dependencies.
- `frontend/src/adapters/api/lesson-video-schema.ts` — strictly validates ticket, asset and
  course-list responses before mapping them to domain values.
- `frontend/src/adapters/api/lesson-video-repository.ts` — maps the generated JSON operations
  and their business errors, while delegating binary transfer/cancel to the tus client.
- `frontend/src/adapters/api/tus-client.ts` — builds the exact tus request headers, parses
  response offsets/lengths, maps statuses, computes bounded chunks, performs one timed
  exchange at a time, and persists a versioned fingerprint keyed by ticket/video id.
- `frontend/src/adapters/api/transport.ts` — provides the shared cancellation, deadline and
  dependency-failure classification used by both JSON and tus transports.
- `frontend/src/app/composition/admin-video-queries.ts` and `admin.ts` — provide stable keys,
  bounded polling, mutations/invalidation, and request-scoped media composition.
- The adjacent `.test.ts` and route-fixture files pin the domain/application/schema,
  transport, query and composition contracts without adding an editor component.
- `docs/media/plan.md` — records R44, updates the row counts and M14 ownership, and makes the
  accepted P11 evidence split explicit.
- `work/report-m14.md` — this report.

No M15 field/editor/locale work, e2e, backend, dependency, generated output, task-tracker
entry or phase report was added or changed.

## Discussion

### Interrupted-file identity — accepted and implemented

The earlier report showed that length-only resume could append a different same-sized file's
tail to an existing prefix. Plan §24 accepted the correction. A fresh transfer now stores
`originalName`, `size` and `lastModified` under a versioned key for its ticket id before it
contacts the upload service. Resume reads that record and compares it with the reselected
`File` before `HEAD`:

- a match may continue from the parsed server offset;
- a mismatch or absent/malformed record returns `upload_resume_unverified` before any upload
  request, leaving M15 to render cancel-and-restart;
- a repeated fresh-start call cannot overwrite a different valid fingerprint; and a ticket
  with no prior valid fingerprint cannot accept a positive server offset;
- even after fingerprint verification, the server's `Upload-Length`, the resume target's
  `declaredBytes`, and `File.size` must all agree.

The fingerprint is removed after successful completion or successful `DELETE`; a newly
written fingerprint is also discarded if its first `HEAD` unexpectedly reveals a positive
offset with no previously verified record.

### MSW/happy-dom custom-header loss — implemented test split

Plan §24 records the measured environment defect: custom request and response headers do not
survive this repository's MSW + happy-dom path. Relying on resolver assertions or mocked
`Upload-Offset` headers would therefore test the harness rather than the client.

The header-dependent rules are exported as pure functions and directly unit-tested:
`buildTusHeaders`, `parseTusHead`, `parseTusOffset`, `tusStatusCode`, and `nextTusChunk`.
Direct in-memory transport responses exercise the matching-fingerprint resume from offset 3
and the retained declared-length refusal. MSW is used only for a PATCH body/status, JSON
status/bodies and `DELETE`. The real request/response header path remains M16 Compose evidence;
none of the M14 rows claims otherwise. Plan §24 labels this arrangement provisional and
records B-2 to move non-rendering adapter tests to a Node environment after this phase; that
future harness change is outside M14.

### P11 task boundary — accepted split

M14 proves the query policy with a real QueryObserver and MSW: polling continues through
`processing`, stops at `ready`, does not start for a lone `created` row, and the explicit
in-tab-transfer flag selects the interval. The updated P11 row leaves the editor-to-flag
wiring with M15, so this task did not mount or add `LessonVideoField`.

## Acceptance rows

| Row | M14 evidence | Wire/UI remainder | Status here |
|---|---|---|---|
| P11 | QueryObserver + MSW cadence and the pure live-transfer predicate | M15 mounts the editor and supplies the live flag | written, not run |
| B09 | Existing M09-generated operations/models are consumed without editing `generated/` | the drift command itself is denied here | unverified |
| R23 | JSON 409/403/500 are mapped through MSW; tus 400 is covered by the pure status map plus an MSW PATCH body/status and an application transfer-passthrough assertion; exact tus headers are covered by pure construction/parsing tests | M16 proves custom headers over a real Compose connection | written, not run |
| R44 | Mismatch and absent fingerprint both refuse before `HEAD`/`PATCH`; a match resumes at the direct `Headers` offset; declared length is rechecked | M15 presents cancel-and-restart; M16 proves the real header exchange | written, not run |

## Transport vocabulary

`createRequestControl` remains the single implementation of caller cancellation, per-request
deadline and unreachable-dependency classification. The tus transport reuses that control
but keeps protocol-specific status semantics: 400 → `unsupported_media`, 404 →
`video_not_found`, 413 → `upload_too_large`, 503/507 → `dependency_unavailable`, unnamed 4xx
→ `invalid_response`, and unnamed 5xx → `dependency_unavailable`. It does not parse binary
responses as JSON or send cookies.

## Verification performed

- `cd frontend && npm run typecheck` passed after the final edits (`tsc --noEmit`).
- `cd frontend && npm run lint` passed after the final edits; oxlint emitted no warning and
  the architecture check reported `Architecture boundaries checked (77 files).`
- `cd frontend && npm run build` passed for client and SSR. It emitted the pre-existing
  warning that `src/routes/route-guards.test.tsx` is discovered as a route file without a
  `Route` export; both builds completed.
- The required two-axis read-only review was run without Git; its final findings are recorded
  below.
- No filesystem write was denied. One combined patch request was rejected by the patch
  parser, not the filesystem, with the exact text
  `apply_patch verification failed: invalid patch: multiple operations target /Users/strelia/projects/courses/frontend/src/adapters/api/tus-client.test.ts`.
  It made no write; subsequent separate `apply_patch` operations succeeded.
- No Git command was run. No file in `generated/`, `work/tasks.md`, or
  `docs/media/report.md` was written.

## Not run

Environment denial: **Docker is denied; Rstest is denied because it cannot bind its port in
this environment; oxfmt is denied.** No workaround was attempted, and no blocked check is
represented as passing.

- Targeted/new Rstest, `npm test`, and `make test` were not run. P11, R23, R44, schema,
  application and transport tests are written for the reviewer.
- `npm run format:check` and `make format-check` were not run because oxfmt is denied; the
  Make target also invokes Docker-backed Go formatting. Neighbour formatting was matched.
- `make lint` was not run because its Go-vet portion requires denied Docker; only the
  frontend-local lint command above ran.
- `make typecheck` was not run; its underlying frontend command was run directly and passed.
- `make integration` was not run because it requires Docker/Testcontainers.
- `make generate` and `make check-generated` were not run: generation writes within
  `generated/`, its Go toolchain requires Docker, and the drift target invokes forbidden Git.
  B09 remains unverified; no substitute drift check is claimed.
- `make build`, `make up`, `make seed-identities`, `make seed`, `make smoke`, `make e2e`, and
  aggregate `make check` were not run because they require Docker and/or denied test
  execution. Only the frontend-local build above ran.

## Review findings

### Standards

No documented-standard violation was found. The reviewer identified one possible Mysterious
Name: `keepPolling` / `keepCreated` actually held unsubscribe callbacks. They are now named
`unsubscribePolling` / `unsubscribeCreated`.

### Specification

The reviewer found R23 partial because the tus 400 assertion stopped at
`createTusTransport().send`, while the row requires the matching use case. The application
test now proves `unsupported_media` passes unchanged through the transfer use case; the pure
status map and MSW PATCH body/status still prove the two lower layers without relying on
custom headers. No other missing §24 behavior, scope creep or incorrect resume behavior was
found in M14. The follow-up reviewer separately questioned §24's provisional/B-2 wording;
that gate-owner plan text was not introduced or removed here and remains the binding future
work record.

Review summary: standards 1 judgement call, corrected; specification 1 partial-evidence
finding, corrected. No M14 finding remains open.

## Limitation and next task

The reviewer must run oxfmt, the written Rstest suite, Orval drift verification and the
ordered gates in a permitted environment. M15 may now add only the documented editor
affordances, including mapping `upload_resume_unverified` to cancel-and-restart and supplying
the P11 live-transfer flag. M16 owns real tus request/response header behaviour over Compose.
Plan backlog B-2 may replace happy-dom for non-rendering adapter tests only after the phase.

## Reviewer verification (measured in the reviewing session)

Both items accepted and recorded as plan **§24**.

**The integrity hole was real and is closed at the right place.** Resume after reload could
splice a different same-sized file onto the stored prefix and enqueue a silently corrupt
source — worse than obvious corruption, because a spliced file can carry a valid container
prefix, pass ffprobe and reach `ready` as a "successful" video. The client now persists
`{name, size, lastModified}` with the ticket and **refuses to resume on mismatch and on
absence alike** — verified in the code that the refusal happens *before* `HEAD`, so no path
exists on which unverified bytes are appended. Row **R44** covers both refusals.

**The MSW header limitation was measured, not assumed:** custom headers survive neither
direction under MSW + happy-dom, and `Access-Control-Expose-Headers` changes nothing. The
tests are restructured on the §I05 precedent — header logic as pure functions (the test names
say so explicitly), MSW kept for bodies, statuses and `DELETE`, and the wire proof owned by
M16. **This is recorded as provisional**: backlog **B-2** carries the likely fix — the
adapter tests run under `happy-dom` although none of them renders anything — to be taken as
Phase 5 closes, before Phase 6's player adapter speaks in headers too.

Gates: `make format` applied, then format-check, lint, typecheck pass and `make test` is
green at **319 passing, 0 failing**. Provenance: implemented with Codex (gpt-5.6-sol);
reviewed and verified here.
