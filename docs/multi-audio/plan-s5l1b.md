# Plan — multi-audio S5-L1b (audio upload + enqueue)

Base `origin/main@57d3832` (S5-L1a merged PR #99). Writer=Courses 1,
review_mode=independent_subagent. All product decisions closed by parent plan D-S5L1-2
(owner 2026-11-08: byte-columns + sweep); architecture mirrors proven video upload.

## Мета
Enable the uploaded-track path end-to-end: mint an audio-ticket → tus upload → enqueue
audio transcode → worker promotes track → master regenerated. Closes the last gap before
the uploaded-track UI (S5-L2) and frontend player (S7).

## Scope (small, no new product decisions)

All design choices are D-S5L1-2 instantiation:
- Migration adds byte-columns + expires_at to audio_tracks (mirrors video_assets)
- Mint creates a Created track with declared_bytes and ownership check
- Tus route reuses the chunked-upload protocol at a separate path
- Completion triggers EnqueueAudioTrack (already wired in asynq/queue.go)
- Track-level sweep for abandoned Created tracks (mirrors asset-level sweep)
- Self-eviction (D-S5L1-6): leave as-is per recommendation (cap-consistent, no strand)

## Surface map (facts, verified 2026-11-09)

- `audio_tracks` (00025) has NO byte columns, NO expires_at, NO claim_token.
- `domain.AudioTrack` struct has NO DeclaredBytes/ReceivedBytes/ExpiresAt fields.
- `CreateAudioTrack` (postgres/audiotrack.go:67) is NOT on the port interface — only on concrete repo, only used in tests.
- `ports.TranscodingQueue.EnqueueAudioTrack` exists (queue.go:19); asynq impl exists (queue.go:40-50); NO application call-site.
- Video upload: `Upload.AppendChunk` (upload.go:57) → repo `AppendChunk` → `ChunkPolicy`/`CompletePolicy` → enqueue. Keyed on `video_assets.id`.
- Tus server (tus/server.go): uploadService interface {HeadUpload, AppendChunk, CancelUpload} all return `domain.VideoAsset`. Route `/uploads/{videoId}`.
- Sweep (transcode.go:279): `SweepCandidates` SQL selects `video_assets` where `evicted_at IS NOT NULL OR (state IN (created,uploading) AND expires_at < now)`. Audio tracks have `ON DELETE CASCADE` but no separate sweep.
- `processAudioTrack` (transcode.go:174) already handles Queued→Ready (PromoteAudioTrack) and failure (FailAudioTrack). L1 wired terminal fail; L2 wired eviction.

## Рішення (instantiate D-S5L1-2, author Courses 1, 2026-11-09)

### D-L1b-1 — Migration 00027: byte-columns + expires_at on audio_tracks
Add `declared_bytes bigint`, `received_bytes bigint NOT NULL DEFAULT 0`,
`original_name text NOT NULL DEFAULT ''`, `expires_at timestamptz` to audio_tracks.
No claim_token — audio tracks don't need retry/reclaim (simpler lifecycle per owner's
"lean" model: Created→Uploading→Queued, one attempt per track; retry = delete + re-mint).
CHECK constraints mirror video_assets: `declared_bytes > 0 WHEN state <> 'ready' AND state <> 'failed' AND kind = 'uploaded'`.
Embedded tracks keep defaults (0/NULL) — they never upload.
Down: DROP the columns (ALTER TABLE ... DROP COLUMN).

### D-L1b-2 — Separate AudioUpload service (not polymorphic Upload)
Video `Upload` works with `domain.VideoAsset` throughout: tus.Server interface, AppendChunk
return type, storage keys. Reusing it for audio would require generics or type assertions.
Instead: `application.AudioUpload` service with `ports.AssetRepository` audio-track methods.
Separate `tus.AudioServer` at `/uploads/audio/{trackId}` with its own `audioUploadService`
interface returning `domain.AudioTrack`. The two tus servers share the same `tus.tusHeaders`,
`tus.validID`, etc. helpers.
**Closed alternative:** polymorphic Upload with generics — rejected, unnecessary complexity
for two concrete types with different tables/keys/enqueue paths.

### D-L1b-3 — Mint on audioService (handler.go), not uploadService
`POST /admin/videos/{videoId}/audio-tracks` (mint) joins the existing audioService interface
alongside List/SetDefault/Rename/Delete. The handler already has `audioService`; adding mint
there keeps one interface per entity. The mint response carries uploadURL pointing at the
audio-tus route.
**Closed alternative:** separate mintAudioService — rejected, one more field on Handler for
one method.

### D-L1b-4 — Track-level sweep: SweepAudioTrackCandidates
New port method: `SweepAudioTrackCandidates(ctx, now, limit) ([]domain.AudioTrack, error)`.
SQL: `SELECT ... FROM audio_tracks WHERE kind='uploaded' AND state IN ('created','uploading')
AND expires_at < $1 LIMIT $2`. Sweep purges the track's source leaf, then
`DeleteAudioTrackSwept(ctx, trackID)` (DELETE row). Called from `Transcode.sweep` after the
asset-level sweep, same cadence.
No claim_token sweep — track has no reclaim.

### D-L1b-5 — No container sniffing on audio upload first chunk
Video upload sniffs the first chunk via `domain.AcceptContainer` (checks MP4/WebM header).
Audio upload SKIPS this: audio can be raw AAC/MP3 without a container. Validation happens
at transcode time via `AcceptAudioTrackSource(probe)` which checks the ffprobe codec.
This keeps the upload fast and avoids false rejections.

### D-L1b-6 — Self-eviction: leave as-is (D-S5L1-6)
Per plan recommendation and S3/S5-L0 review: the current PromoteAudioTrack eviction is
cap-consistent and reclaims bytes. No code change.

## Кроки (leaf-first)

### Step 1 — domain + migration
- Add `DeclaredBytes int64`, `ReceivedBytes int64`, `ExpiresAt time.Time` to `AudioTrack` struct
- Add `OriginalName string` to AudioTrack (for the upload's original filename)
- Update `NewUploadedTrack` validator: require `DeclaredBytes > 0`, `ExpiresAt` non-zero, `ReceivedBytes == 0`
- Add `ValidateAudioTicket(declaredBytes, originalName)` (mirror `ValidateTicket`)
- Migration 00027: ALTER TABLE audio_tracks ADD COLUMN declared_bytes/received_bytes/original_name/expires_at
- Unit tests for new validation

### Step 2 — repo: MintAudioTrack + AppendAudioChunk + sweep
- Port: `MintAudioTrack(ctx, actor, assetID, declaredBytes, originalName, language, label, expiresAt, now) (AudioTrack, error)` — ownership check (same join as other audio methods), insert with byte columns
- Port: `AppendAudioChunk(ctx, trackID, offset, now, ChunkPolicy, CompletePolicy) (AudioTrack, error)` — mirror AppendChunk but on audio_tracks
- Port: `HeadAudioUpload(ctx, trackID) (AudioTrack, error)` — load if Created/Uploading and not expired
- Port: `CancelAudioUpload(ctx, trackID, now) error` — fail the track
- Port: `SweepAudioTrackCandidates(ctx, now, limit) ([]AudioTrack, error)` + `DeleteAudioTrackSwept(ctx, trackID) error`
- Postgres implementations mirroring video repo methods
- Integration tests: mint with ownership, append→state transitions, completion→Queued, sweep expired

### Step 3 — application: AudioUpload + AudioTickets + sweep wiring
- `AudioTickets.MintAudioTicket(ctx, actor, assetID, declaredBytes, originalName, language, label)` — validate, compute expiresAt, call repo
- `AudioUpload.AppendChunk(ctx, trackID, offset, body)` — mirror Upload.AppendChunk without container sniff; on completion EnqueueAudioTrack
- `AudioUpload.HeadUpload(ctx, trackID)` — delegate to repo
- `AudioUpload.CancelUpload(ctx, trackID)` — delegate to repo
- Wire sweep: `Transcode.sweep` calls `SweepAudioTrackCandidates` after asset sweep
- Unit tests: mint validation, completion→enqueue

### Step 4 — tus: AudioServer
- `tus.AudioServer` struct with `audioUploadService` interface {HeadUpload→AudioTrack, AppendChunk→AudioTrack, CancelUpload}
- Register at `/uploads/audio/{trackId}`
- HEAD/PATCH/DELETE mirroring Server but returning AudioTrack fields
- Server test mirroring server_test.go

### Step 5 — HTTP handler + OpenAPI
- Extend `audioService` interface: add `MintAudioTicket(...) (AudioTrack, error)`
- Handler method `mintAudioTicket`: decode `mintAudioTrackRequest{DeclaredBytes, OriginalName, Language, Label}`, call service, respond 201 with `audioUploadTicketDTO{TrackID, UploadURL, ChunkBytes, MaxBytes, ExpiresAt}`
- Register: `POST /api/v1/admin/videos/{videoId}/audio-tracks` (this path already exists for LIST in L1a via GET; POST is a new verb on the same path)
- OpenAPI contract.go: add RouteBinding, NameMap entries (MediaAudioUploadTicket, mintAudioTrackRequest), overlay
- Handler contract test (AT09: mint success, AT10: mint non-owner, AT11: mint invalid body)
- Coverage + fidelity tests updated

### Step 6 — composition root + wiring
- `cmd/api/main.go`: create AudioTickets, AudioUpload, AudioServer; wire into Handler and mux
- `cmd/media-worker/main.go`: no changes (processAudioTrack already wired)

### Step 7 — ordered gate
Full ordered `make check`: format-check, lint, typecheck, test, check-generated, build, integration, e2e.
Existing video upload→Ready must stay green. EnqueueAudioTrack path newly reachable (e2e won't test it — no UI; integration tests prove it).

## Тести
- Domain: ValidateAudioTicket, NewUploadedTrack with bytes
- Repo integration: mint with ownership denial, append→Uploading→Queued transitions, completion enqueue trigger, sweep expired tracks, cancel
- Tus: AudioServer HEAD/PATCH/DELETE contract
- Handler: AT09-AT11 (mint success/non-owner/invalid)
- OpenAPI: coverage (mux⊇Routes) + fidelity (byte-identical)
- Existing: video upload→Ready green, audio metadata CRUD green

## Відкриті питання
None. All decisions closed by D-S5L1-2 (owner) and this plan.
