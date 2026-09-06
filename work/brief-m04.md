# Brief M04 — media application use cases over fakes

Mode: ponytail-full. Application sees domain + ports only. Mirror the 4a/4b idiom.

## Read first, in this order
1. `docs/media/plan.md` — **§2** application rows (`application/upload.go`,
   `application/transcode.go` and their tests: the exact function list), **§1** the state
   matrix (T1–T9) your use cases drive, **D-10** (the twelve `Process` steps, the claim-token
   guards, step 1/2's EXHAUSTIVE nil list — `ready`, `failed`, evicted, absent,
   `lesson_id IS NULL`, `created`, `uploading` → nil; only `queued`/`processing` continue;
   and step 12's cause-sensitive branch: `DeadlineExceeded` → `Fail(render_failed)`, plain
   `Canceled` → `Requeue`), **D-7** (recovery classes C5/C6 and reclaim C1–C4),
   **D-15** (dependency failure: Redis at enqueue, disk exhaustion, Postgres down — the
   terminal state is DEFERRED, never guessed), **§8** error vocabulary, and your owned rows
   **R14–R19, R31, R32, R33, R36, R42** — read each row's text; your fake-port tests are
   their evidence.
2. `backend/internal/identity/application/administration.go` + test (4b) — the closest
   sibling: bounded deadlines, cancellation kept distinct from dependency failure, and the
   policy-callback-into-the-repository idiom.
3. `backend/internal/course/application/authoring.go` + test (4a) — the other sibling.
4. `backend/internal/media/domain/` and `ports/` (M03, just landed) — what you consume.

## The task (plan §10 M04, 2h)
- `backend/internal/media/application/upload.go`: `MintTicket`, `BeginUpload`,
  `AppendChunk`, `CompleteUpload`, `CancelUpload`, `ListCourseVideos`, `Retry`.
- `backend/internal/media/application/transcode.go`: `Process` (the twelve steps over the
  four ports), `Sweep`, `Recover`.
- Bounded deadlines and cancellation like the siblings; dependency failure passes through
  and is never relabelled as a business refusal.
- `application/*_test.go` with FAKE ports only: every transition, retry exhaustion,
  malformed source, no-op on `ready`/absent, duplicate enqueue, enqueue-after-commit
  failure, cancellation (both causes — this is R42's point), disk exhaustion, dependency
  failure. The fakes must record call counts so "the repository was never called" is
  provable where a row demands it.

## What this task is NOT
No postgres adapter (M05/M06), no filesystem or ffmpeg adapters (M07/M08), no HTTP (M09),
no tus server (M10/M11), no frontend. If a domain rule or port method is missing, STOP and
report rather than adding it here.

## Environment limits — name the denial, do not work around it
No Docker: no Go compile or test. gofmt + read-through. List every unrun check. Denied
write → report the exact text and stop. **No git; do not edit work/tasks.md.** No generated/.

## Report
`work/report-m04.md`.
