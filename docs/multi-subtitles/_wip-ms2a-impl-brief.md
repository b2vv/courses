# MS2a — implementation brief for Codex (run 1 of 3: implementation; tests are run 2)

Repository: current working directory (`courses`). Branch `ms2-3-asr-per-track` @ `011ff68`.
Shared working tree with another session: do NOT run git checkout/reset/stash/clean/rm, do not commit,
do not touch files outside the list below. Only edit; Courses 1 reviews, runs gates and commits.

WORKING FILE: docs/multi-subtitles/_wip-ms2a-impl.md
Append AFTER EVERY STEP (append, not rewrite). First entry before the first tool: a `NEXT:` line.
Then FACT / DECISION / NEXT / QUESTION. Final answer = concise summary of the file.

## Read first
- `AGENTS.md` (architecture: domain is stdlib-only; modules never import each other; adapters map
  explicitly; no unused ports).
- `docs/multi-subtitles/plan-ms2-3.md` — ACCEPTED plan. Authoritative decisions for this run:
  **MS-D1 (backend part), MS-D2, MS-D3, MS-D10 (storage + media domain/repository only)**. Read
  MS-D4/MS-D5 for context only.

## Scope of THIS run
1. **Migration `backend/migrations/00031_asr_draft_per_track.sql`** (goose, `SET LOCAL lock_timeout`
   like neighbours, comments in the house style explaining WHY):
   - `audio_tracks.multilingual boolean NOT NULL DEFAULT false` (MS-D10).
   - `asr_drafts`: add `audio_track_id uuid REFERENCES audio_tracks(id) ON DELETE CASCADE`,
     `track_language text` (same grammar CHECK as `language` in 00008 + `= lower()`; `und` allowed),
     `track_multilingual boolean NOT NULL DEFAULT false`.
   - Backfill per MS-D2: match the track of the draft's own asset with
     `source_leaf = COALESCE(draft.source_leaf,'original')` and `state <> 'failed'`;
     `track_language = lower(track.language)`; unmatched drafts DELETEd.
   - Then `audio_track_id`, `track_language` NOT NULL; PK → `audio_track_id`; keep `lesson_id`
     column (still FK cascade to lesson) + index `asr_drafts_lesson (lesson_id)`; keep `asset_id`,
     `generation`; DROP `source_leaf` (+ its constraint); keep `asr_drafts_running`.
   - Down (test-only): keep one draft per lesson (the default track's, else newest `updated_at`),
     restore `source_leaf` from the track's leaf, PK back to `lesson_id`, drop new columns
     (including `audio_tracks.multilingual`). Check `migrations/README.md` for any registry.
2. **Media module** (storage only): `domain.AudioTrack.Multilingual bool`; postgres adapter reads it
   in every SELECT of audio tracks and writes it on INSERT (sqlc if the module uses sqlc — then
   regenerate with the repo's make target / document the command you could not run). Constructors
   keep their signatures in this run; the env default and HTTP are run MS2b/MS2c. Do not change
   media HTTP DTOs yet.
3. **Transcript domain** (`internal/transcript/domain/draft.go`): `TranscriptDraft` gets
   `TrackID string`, `TrackLanguage Language` (validated like Language but `und`-family allowed),
   `TrackMultilingual bool`; `SourceLeaf` stays as the leaf read through the track join (no longer
   stored on the draft row). Replace `Stale(readyGeneration, defaultLeaf)` with the MS-D5 rule:
   `Stale(currentTrackLanguage Language, currentMultilingual bool) bool` =
   (`TrackLanguage != lower(current)` AND `TrackLanguage` not und-family) OR
   `TrackMultilingual != currentMultilingual`. Add an `IsUndetermined(Language) bool` helper
   (`und` or `und-*`) in domain if none exists (MS1 may already have one in media — do not import
   media; transcript keeps its own).
4. **Transcript ports** (`ports/draft.go`) per MS-D3:
   - `GetLessonDraft` → `ListLessonDrafts(ctx, lessonID) ([]LessonDraft, error)`: one per track of
     the lesson's latest Ready non-evicted asset that has a draft; each carries the draft, token,
     the track's CURRENT language and multilingual (so Stale is computable), and `IsDefault`.
   - `PendingRecognition` gets `TrackID`; pending = every `ready` track of the latest Ready
     non-evicted asset of a lesson that has NO draft row (O7: rename/multilingual change does not
     create pending). `TaskMarker` stays (empty for no draft).
   - `Claim(ctx, lessonID, assetID string, generation int32, trackID string, requested bool, now)`:
     ONE statement `INSERT … SELECT FROM audio_tracks JOIN video_assets` verifying lesson, asset,
     generation, track belongs to asset and is `ready`, and that the asset is the lesson's latest
     Ready non-evicted asset; writes `track_language = lower(track.language)`,
     `track_multilingual = track.multilingual`; `ON CONFLICT (audio_track_id) DO UPDATE … WHERE $requested`;
     `RETURNING` track_language, track_multilingual and the track's `source_leaf`.
     Wrong/foreign/missing track → `domain.ErrNotFound` (C21); track of a non-latest Ready asset →
     `domain.ErrDraftSuperseded`; conflict without requested → `Claimed:false`. Distinguishing
     NotFound vs Superseded may use a follow-up read when the INSERT inserted nothing, as today.
   - `Complete`/`Recover` keyed by `audio_track_id` + token; `RecoveredDraft.Draft.TrackID` set.
   - `SaveDraft` REMOVED (MS-D1). `LessonAsset` → replace with what Regenerate needs:
     `LessonTrack(ctx, lessonID, trackID) (*PendingRecognition, error)` returning the track iff it
     is a `ready` track of the latest Ready asset (nil otherwise), with duration.
   - Keep the long WHY comments; rewrite the ones that describe the per-lesson model.
5. **Postgres adapter** (`adapters/postgres/draft.go`) implements the above. Remove
   `readyAudioForLesson` usage if it becomes dead.
6. **Application / asynq / HTTP — compile-through with the plan's semantics**:
   - `ports.RecognitionJob` and the asynq payload carry `TrackID` (`trackId` JSON); TaskID
     `asset:generation:track:marker`; `decodePayload` refuses a payload without `trackId`
     (`errInvalidTaskPayload`, no retry).
   - `Recognize.Process` claims with the job's track; `EnqueuePending` enqueues per track;
     `Recover` per track. Hint/multilingual plumbing to the Transcriber is run MS2b — keep the
     current hint source for now but do not invent a new one.
   - `Drafting`: remove `Publish`, `PublishRequest`, `TranscriptPublisher` and its `NewDrafting`
     parameter + composition-root wiring (MS-D1). `GetDraft`/`Regenerate` adapt to the new ports
     with the MINIMUM to compile and keep current HTTP endpoints behaving for the default track
     (GET /draft returns the default track's draft; regenerate targets the default track) — the
     new per-track HTTP is run MS2c. Remove the `/draft/publish` handler and route now; do NOT run
     openapi generation (Courses 1 runs `make generate`).
   - Update existing unit-test fakes/fixtures ONLY as far as needed to compile; do not delete test
     cases except those for removed code (Publish, SaveDraft). List every test you changed.

## Out of scope for this run
Whisper hint rule and `--multilingual` (MS-D4/M2-R12), `ASR_MULTILINGUAL_DEFAULT`, media HTTP,
new transcript HTTP endpoints, openapi regeneration, frontend, new integration tests (run 2).

## Done = report
- Files changed; `cd backend && go build ./... && go vet ./...` result (you may run them);
  `go test ./internal/transcript/... ./internal/media/...` without integration tag if it runs.
- Any place where the plan was ambiguous: QUESTION in the working file, then choose the reading
  closest to the plan text and say so. If a plan rule would cause a bug, stop and argue it
  (QUESTION) instead of implementing around it.
