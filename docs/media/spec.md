# Media upload and transcoding — Phase 5 specification

Status: GATE 1 decisions taken 2026-09-06. Intent only. Roadmap: `work/PRODUCT-BRIEF.md` §21 (Phase 5)
with the media architecture of §11 and the Media module of §10.

## Problem

Lessons exist but carry no video. The media-worker runs with an empty handler mux
(`backend/cmd/media-worker/main.go:50`), ffmpeg sits unused in its image, the `media-data`
volume is empty, and `/media/*` is refused at the edge. Nobody can put a video on a lesson,
so the platform still cannot deliver a course — Phase 6 (playback) has nothing to play and
Phase 8 (orders) has nothing worth selling.

## User scenarios

Actors: **admin** (any course), **instructor** (own courses only — the 4a ownership rule),
**student / anonymous** (no access to any of this; media bytes stay refused until Phase 6).

- **S1 Upload.** An authorized editor opens a lesson in the course editor, chooses a video
  file, and uploads it. The lesson shows the upload progressing, then a processing state.
- **S2 Processing.** The upload is queued; the worker probes the source, transcodes it into
  HLS renditions, writes a master playlist and a thumbnail, and atomically publishes the
  result. The admin sees the state change without reloading (or on a refresh — Q5).
- **S3 Status and failure.** Every video carries an explicit state
  (`created → uploading → queued → processing → ready | failed`). A failed job shows why in
  operator terms, and can be retried.
- **S4 Replace.** An editor uploads a new video for a lesson that already has one; the new
  asset supersedes the old one once it is ready, and the old bytes are cleaned up (Q6).
- **S5 Refusal.** Anonymous, student, and a non-owning instructor are refused server-side on
  every upload and status operation. Media bytes remain unreachable to everyone in Phase 5.
- **S6 Admin visibility.** The course editor shows each lesson's video state, and a failed
  asset is visibly distinguishable from a missing one.

## Acceptance criteria (binding outcomes; measurable rows delegated to acceptance-spec)

1. An authorized editor can upload a video for a lesson they may edit; the bytes land in
   the media volume and the source is never lost on a partial or interrupted upload.
2. The video state machine advances exactly as specified, and every terminal state
   (`ready`, `failed`) is reached from a real worker outcome, never guessed.
3. Transcoding produces a master playlist plus the agreed renditions and a thumbnail,
   published atomically — a consumer never sees a half-written asset.
4. A failed or cancelled job leaves no partial output and no orphaned temporary files, and
   is retryable with bounded attempts.
5. Every upload and status operation is refused server-side for non-authorized actors;
   ownership follows the 4a rule.
6. `/media/*` remains refused at the edge for everyone in Phase 5 — no playback is claimed.
7. The OpenAPI document and generated clients change together; the drift gate stays green.
8. The full manifest gate (`make check`) passes with the new surface from a clean checkout,
   including the worker actually processing a job.

## Out of scope

- Playback: signed URLs, `auth_request`/`secure_link`, hls.js, the player, progress
  (Phase 6). Phase 5 produces the bytes; nobody may read them yet.
- Transcripts and captions (Phase 7); orders and payments (Phase 8).
- S3/MinIO storage: the local filesystem is the first adapter; the port must not leak
  filesystem assumptions, but no second adapter is written.
- DRM of any kind; the brief forbids claiming protection the system lacks.
- Automatic speech recognition; no Python runtime.
- Per-course user blocking, role change (4b out-of-scope list stands).

## GATE 1 decisions (2026-09-06)

| # | Decision |
|---|---|
| Q1 | **Resumable upload (tus)**, terminated in a **separate upload service** — a third executable beside `api` and `media-worker` (brief §2 rule 12's shape), or tusd, writing chunks onto the media volume and notifying the API on completion. The Go API never receives video bytes, so brief §2 rule 13 holds for ingest as well as egress. An ADR records the protocol and the placement. |
| Q2 | Maximum 2 GiB; the accepted set is probed with ffprobe (mp4/mov/mkv carrying H.264/HEVC video and AAC/MP3 audio); anything else is refused with an explicit business error before a job is queued (recommendation, uncontested). |
| Q3 | Ladder 360p/720p/1080p-only-if-the-source-supports-it, H.264/AAC, HLS with **fMP4 (`.m4s`) CMAF-style segments** — chosen so Phase 6 players and any future DASH reuse the same bytes with no second remux. ADR required by the brief. |
| Q4 | Upload authorization reuses the 4a ownership rule exactly: admin any lesson, instructor only lessons in courses they own. No new authorization concept (recommendation, uncontested). |
| Q5 | The admin UI polls asset status on a bounded interval while a job is in flight and stops at a terminal state. No websockets or SSE — no second consumer exists for a push channel (recommendation, uncontested). |
| Q6 | Replacing a lesson's video **retains the previous asset**: exactly two generations live per lesson (current + one previous, a one-step rollback). A third upload evicts the oldest, so growth is bounded without a quota subsystem. |
| Q7 | Uploads are allowed for lessons in published courses — 4a already makes edits to a published course immediately public, and a lesson without video is the weaker product (recommendation, uncontested). |
| Q8 | Asynq performs bounded automatic retries for transient failures; after exhaustion the asset rests in `failed` with a manual admin retry. A source rejected as malformed is never retried (recommendation, uncontested). |
