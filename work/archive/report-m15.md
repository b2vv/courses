# M15 report — LessonVideoField and course-editor mount

## Files

- `frontend/src/presentation/admin/LessonVideoField.tsx` — renders the per-lesson file
  control, determinate upload progress, D-12 status selection, bounded stale retry,
  interrupted resume/cancel, closed failure reasons, refusal messages and degraded states.
- `frontend/src/presentation/admin/LessonVideoField.test.tsx` — component evidence for P09,
  B20, R24, R25 and R41, including an axe assertion.
- `frontend/src/presentation/admin/CourseEditorView.tsx` — mounts the field inside each lesson
  fieldset and tracks which lesson ids exist in the last loaded or successfully saved tree.
- `frontend/src/routes/admin.courses.$courseId.tsx` — wires the M14 list/mint/transfer/retry/
  cancel query options into the editor and supplies the explicit live-transfer polling flag.
- `frontend/src/presentation/ui/controls.tsx` — keeps the existing `FileUpload` kit control
  non-submitting when mounted inside the editor form.
- `frontend/src/app/providers/locales/{uk,en}.ts` — adds the complete lesson-video vocabulary
  in both locales.
- `frontend/src/app/styles/main.css` — styles the field, state badge, failure/degraded text and
  action group using existing semantic tokens and the kit's 44px/focus-visible controls.
- `work/report-m15.md` — this report.

No backend, dependency, e2e, tus refusal-semantic, generated, task-tracker or phase-report
change was made.

## Behaviour delivered

- A saved lesson with no active video offers the existing `FileUpload` kit control, restricted
  to MP4/MOV/MKV. A live transfer renders a determinate accessible `ProgressBar` and keeps the
  separate course-video query polling while any transfer in this tab remains active.
- Badge selection follows D-12: a non-terminal row wins; otherwise the highest-generation
  `ready` row wins; otherwise the highest-generation `failed` row wins. Fresh queued and
  processing rows render indeterminate progress.
- Retry is offered for a fully received failed source, queued after 10 minutes, and processing
  after 40 minutes. Incomplete failed uploads do not expose a retry the server will refuse. A
  successful retry is locally suppressed until refreshed status arrives.
- `created`/`uploading` without a live field transfer is presented as interrupted. Reselecting
  uses M14's resume target; `upload_resume_unverified` leaves resume and cancel visible and
  explicitly directs cancel-and-restart. Successful cancel immediately frees the local field
  while the invalidated query refreshes.
- A newly added lesson has no file control until a successful course save returns that lesson
  id. It shows the localized save-first hint instead, avoiding a UI-caused
  `lesson_not_found` request.
- Every business/refusal outcome in the media domain has localized field copy. A failed list
  read disables minting when status is unknown; action failures remain visible without
  permanently removing a safe retry or replacement path.

## Acceptance rows

| Row | Written component evidence | Status here |
|---|---|---|
| P09 | file selection starts a mocked live transfer and renders 50% determinate progress | written, not run |
| B20 | trigger is the keyboard-first kit button, progress exposes `aria-valuenow`, and axe is asserted | written, not run |
| R24 | adding a new lesson leaves only the saved lesson's upload trigger and renders the localized hint | written, not run |
| R25 | interrupted state exposes resume/cancel, refusal preserves cancel-and-restart, and successful cancel frees the field | written, not run |
| R41 | 11-minute queued and 41-minute processing rows expose retry; fresh equivalents expose progress only | written, not run |

The same component suite also pins non-terminal/ready/failed badge precedence, a closed
failure reason with retry, and list/action dependency degradation.

## Verification performed

- `cd frontend && npm run lint` passed after the final edits. Oxlint emitted no warnings and
  the architecture check reported `Architecture boundaries checked (78 files).`
- `cd frontend && npm run typecheck` passed after the final edits (`tsc --noEmit`, including
  the new component test source).
- `cd frontend && npm run build` passed for client and SSR. It emitted the pre-existing
  warning that `src/routes/route-guards.test.tsx` is discovered as a route file without a
  `Route` export; both builds completed.
- A read-only standards/spec review found no remaining documented-standard violation,
  missing M15 behaviour or scope creep. It specifically rechecked the D-12 row ordering and
  thresholds, D-13/§24 refusal path, unsaved-tree boundary, both locale objects, absence of
  presentation `fetch`, and the NOT-list.
- No filesystem write was denied. No Git command was run. No file in `generated/`,
  `work/tasks.md`, or `docs/media/report.md` was written.

## Not run

Environment denial: **Docker is denied; Rstest is denied because it cannot bind its port in
this environment; oxfmt is denied.** No workaround was attempted, no custom tus headers were
asserted through MSW/happy-dom, and no blocked check is represented as passing.

- Targeted `LessonVideoField.test.tsx`, axe, `npm test`, and `make test` were not run because
  they invoke denied Rstest and/or Docker. All M15 acceptance evidence remains unverified.
- `npm run format:check` and `make format-check` were not run because oxfmt is denied; the
  Make target also invokes Docker-backed Go formatting. Neighbour formatting was matched.
- `make lint` was not run because its Go-vet half requires Docker; only the frontend-local
  lint command above ran.
- `make typecheck` was not run; its underlying frontend command was run directly and passed.
- `make integration` was not run because it requires Docker/Testcontainers.
- `make generate` and `make check-generated` were not run. Generation would write inside
  forbidden `generated/` directories, the Go generator requires Docker, and the drift target
  invokes forbidden Git. No generated contract changed in M15.
- `make build`, `make up`, `make seed-identities`, `make seed`, `make smoke`, `make e2e`, and
  aggregate `make check` were not run because they require Docker and/or are outside M15's
  NOT-list. Only the frontend-local build above ran.
- The repository `code-review` skill's git-based diff/log/rev-parse and parallel diff review
  were not run because the brief forbids all Git access and supplies no fixed comparison
  point. The read-only source/spec review above was used; it is not represented as the
  skill's prescribed git review.

## Review findings

### Standards

No remaining finding. The field reuses the UI kit and semantic tokens, keeps transport and
server state out of presentation, adds no dependency or speculative seam, uses no hardcoded
Ukrainian, and preserves the existing editor form/tree/conflict/save flow.

### Specification

No remaining finding. P09, B20, R24, R25 and R41 have written component coverage; the runtime
and axe verdicts are pending the reviewer environment. No bug-causing inaccuracy was found in
the M15 brief or its referenced plan clauses, so no `Discussion` section is required.

Review summary: Standards — 0 open findings. Specification — 0 open findings; 5 acceptance
rows written but unverified under the named Rstest denial.

## Limitation and next task

The reviewer must run oxfmt, the targeted/full Rstest suite including axe, and the ordered
gates in a permitted environment. M16 owns only the Compose browser journey and real tus wire
behaviour; this report does not claim playback, e2e upload/transcoding, or successful runtime
verification.

## Reviewer verification (measured in the reviewing session)

`make format` applied, then format-check, lint and typecheck pass; `make test` green at
**330 passing, 0 failing** (up from 319), axe included.

Two things I specifically checked because they are the easy places to undo earlier work, and
both hold: the field has **zero hardcoded Cyrillic** — the G31-class debt paid in 4a is not
reintroduced — and §24's integrity rule survives all the way to the interface, with a test
named "refuses an unattested resume and keeps cancel-and-restart available". A silent retry
there would have quietly defeated the fingerprint refusal built one task earlier.

No Discussion was raised. Provenance: implemented with Codex (gpt-5.6-sol); reviewed and
verified here.
