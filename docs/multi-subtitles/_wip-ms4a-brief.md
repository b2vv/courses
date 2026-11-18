WORKING FILE: docs/multi-subtitles/_wip-ms4a.md
Append AFTER EVERY STEP (append, not rewrite). The first entry comes before the first tool:
a `NEXT:` line stating what you intend to check. Then use the labels FACT / DECISION / NEXT / QUESTION.
Final answer = a concise summary of the file.

# MS4 run (a) — backend: Ready audio-track languages on the learner lesson (O7 / M3)

Repository `/Users/strelia/projects/courses`, branch `ms4-player-captions` (do not switch branches,
do not commit, do not touch files outside this scope; other sessions share the tree — never run
git checkout/reset/stash/clean). Read first: `AGENTS.md`, `docs/multi-subtitles/plan-ms4.md`
(ACCEPTED) §0, §2, §3 M3, §4 row M03. The plan is the contract.

## Implement
1. `backend/internal/course`: each learner lesson gains `AudioTracks []LearningAudioTrack{Language string; IsDefault bool}`.
   - Source: Ready audio tracks (`audio_tracks.state = 'ready'`) of the asset chosen by the existing
     lateral `current_video` in `adapters/postgres/learning.go:55-63`, **only when that asset is Ready**
     (extend the lateral with the asset id; add a second lateral/aggregate for the tracks). That asset is the
     one playback plays (`backend/internal/playback/adapters/postgres/playback.go:17-23`) whenever a Ready
     asset exists — keep it that way. Order by track id. Language as stored (case kept).
   - Empty for a non-Ready lesson. Empty for a lesson that is **locked for this viewer** — set where
     `Locked` is computed for the viewer (find it: `GetCourse`/domain), so no paid-content languages leak.
   - Domain stays framework-free (ADR rules in AGENTS.md "Architecture"). The course module may read
     `audio_tracks` in its own SQL as it already reads `video_assets`; do not import the media module.
2. HTTP: `learning_handler.go` emits `audioTracks: [{language, isDefault}]` (always an array, never null).
   Update `openapi_contract.go` so `make generate` (sqlc + `go run ./cmd/openapi-gen` + `cd frontend && npm run generate`)
   produces the new field in `backend/openapi/openapi.yaml` and `frontend/src/adapters/api/generated`.
   Run the generation; the generated files must be committed-ready (no hand edits).
3. Frontend strict schema `frontend/src/modules/playback/adapters/playback-schema.ts:42`: add
   `audioTracks: z.array(z.object({language: <same language rule as the media domain, see
   frontend/src/modules/media/domain/audio-track.ts validAudioLanguage>, isDefault: z.boolean()}).strict()).max(16)`
   (check the real per-asset maximum in the media domain and use it), and the matching domain type in
   `frontend/src/modules/playback/domain` (`LearningLesson`). Map it explicitly in the repository adapter.
   Update every unit/route-test fixture of a learning course (grep `clientSeq` / `isPreview` in
   `frontend/src`) and the e2e mock in `frontend/e2e/accessibility.spec.ts` if it builds a learning course body.
   No player/route behaviour change in this run.

## Tests (write them; they are part of this run)
- Integration (`//go:build integration`, real Postgres, existing course learning integration tests as the pattern):
  (i) old Ready asset + newer processing asset → tracks of the old Ready asset;
  (ii) newer asset evicted → previous Ready asset's tracks;
  (iii) failed/queued tracks excluded, order by id, `isDefault` true only for the default;
  (iv) locked lesson for a non-member viewer → empty; same lesson for a member → tracks;
  (v) lesson without video → empty.
- HTTP contract test: `audioTracks` present as `[]` and as a populated array.
- Frontend: schema test accepts the new field, rejects a missing field (strict) and an invalid language.

## Verify before finishing
`cd backend && go test ./internal/course/...`; integration for the course package
(`go test -tags integration ./internal/course/...` with the repo's Docker/testcontainers setup — if Docker
is unreachable in your sandbox, say so explicitly, do not claim it passed); `make generate` then
`git add -N` nothing — just report `git status --short` of generated paths; `cd frontend && npm run format && npm run typecheck && npm run lint && npx rstest run src/modules/playback src/routes`.
Report each command and its real result. Do not represent a skipped check as passed.

## Two channels
- If the contract is missing or contradicts the code (e.g. locked is not computed where you can clear
  tracks), STOP and write a QUESTION in the working file.
- If a plan statement is inaccurate in a way that would cause a bug, argue it in the working file
  (DECISION with evidence) instead of silently following it.
