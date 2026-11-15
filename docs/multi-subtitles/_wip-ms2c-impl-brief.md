# MS2c — implementation brief for Codex (run 1: backend HTTP + openapi overlay)

Repository: current working directory (`courses`). Branch `ms2-3-asr-per-track` @ `6d11e76` (MS2a per-track
drafts + MS2b hint/multilingual/env committed). Shared working tree with another session: do NOT run git
checkout/reset/stash/clean/rm, do not commit, touch only files this brief needs. Courses 1 runs
`make generate` (Docker), integration, gates, review and commits.

WORKING FILE: docs/multi-subtitles/_wip-ms2c-impl.md
Append AFTER EVERY STEP (append, not rewrite). First entry before the first tool: a `NEXT:` line.
Then FACT / DECISION / NEXT / QUESTION. Final answer = concise summary of the file.

## Read first
- `AGENTS.md`; `docs/multi-subtitles/plan-ms2-3.md` (ACCEPTED) — authoritative: **MS-D5** (transcript
  HTTP), **MS-D10 HTTP part** (media `multilingual`, `multilingualDefault`), rows **M2-R10** and the HTTP /
  media parts of **M2-R12**; phase `docs/multi-subtitles/plan.md` **D3** (inventory).
- Precedents: `backend/internal/transcript/adapters/http/{draft_handler,admin_handler,openapi_contract}.go`,
  `backend/internal/media/adapters/http/handler.go`, `backend/openapi/openapi.overlay.yaml`, and the
  openapi fidelity tests in both modules.

## Scope
1. **Transcript drafts HTTP (MS-D5).** Remove `GET …/lessons/{lessonId}/draft` and
   `POST …/draft/regenerate` (handlers, routes, overlay entries, contract entries, tests) — they must
   answer 404 afterwards (test it). Add:
   - `GET /api/v1/admin/lessons/{lessonId}/drafts` → `{drafts: [{trackId, trackLanguage,
     trackMultilingual, state, language, failure, stale, updatedAt, assetId, generation, segments}]}`
     from `ListLessonDrafts`; `stale` = `Draft.Stale(current track language, current multilingual)`.
     Order: default track first, then embedded, then the rest by track language, then trackId (the
     MS-D6 tab order; the next slice's temporary frontend takes `drafts[0]`). Empty list when no
     Ready asset or no drafts. `Cache-Control: no-store`. Access = course editor, as `GetDraft` today.
   - `POST /api/v1/admin/lessons/{lessonId}/drafts/{trackId}/regenerate`: no body (a non-empty body →
     `invalid_request`, same strictness as neighbours), same-origin, `202`. Track not a `ready` track
     of the lesson's latest Ready asset → `404 track_not_found` (new code; add a domain error, e.g.
     `ErrTrackNotFound`, mapped in `writeTranscriptError`; lesson missing stays `lesson_not_found`).
     Draft of that track `running` → `409 already_running`. Uses `LessonTrack(lessonID, trackID)` with
     the real trackID — remove the temporary "empty trackID = default track" convention from MS2a
     (`adapters/postgres/draft.go`), and `Drafting.GetDraft` if nothing else uses it.
2. **Inventory (D3).** `GET /api/v1/admin/lessons/{lessonId}/transcripts` →
   `{transcripts: [{language, updatedAt, stale, isPrimary}]}` ordered by `language`; `stale` uses the
   existing `Transcript.Stale(readyGeneration)` rule. Access: course editor (the plan says "same as
   `GET …/admin/lessons/{id}/transcript`" — that admin GET does not exist; the editor today reads the
   learner GET. Use the editor rule of the admin `PUT …/transcript` and record this as a DECISION).
   Repository method in the transcript postgres adapter; integration test (lesson with 0, 1, 2 languages;
   stale after a newer Ready asset; foreign course → forbidden).
3. **Media HTTP (MS-D10).**
   - `POST …/audio-tracks` (`mintAudioRequest`) accepts optional `multilingual`; absent → the env default
     already held by `AudioTracks`. Thread an explicit `*bool`/value through the application.
   - `PATCH …/audio-tracks/{trackId}` keeps `language` and `label` required, adds optional `multilingual`
     (absent = unchanged). `RegenerateMaster` runs only if `language` or `label` changed; a
     multilingual-only change updates the row without touching the master. Unit test with a counting fake:
     PATCH multilingual-only → 0 master calls; PATCH language → 1. Integration test for the repository write.
   - `GET …/audio-tracks` returns `multilingual` on every item and `multilingualDefault` at top level.
4. **OpenAPI.** Edit `backend/openapi/openapi.overlay.yaml` and the `openapi_contract.go` registries so
   the fidelity tests pass; do NOT run `make generate` (Docker) — Courses 1 runs it and the frontend
   follow-up. You may run `cd backend && go run ./cmd/openapi-gen` if it works without Docker, and say
   so; never hand-edit `backend/openapi/openapi.yaml` or `frontend/src/adapters/api/generated`.
5. **Tests (name with the row id):** M2-R10 HTTP contract tests for every new/removed endpoint and code
   (202 / 404 `track_not_found` / 404 `lesson_not_found` / 409 / 403 / 401 / removed-route 404);
   M2-R12 HTTP: mint without field → env default, with field → field; PATCH flag-only → no master
   rewrite; GET carries both fields. Integration tests compile (`go vet -tags integration ./...`);
   Courses 1 runs them.

## Out of scope
Frontend (next run, after generate), e2e, migrations (none needed — if you think one is, QUESTION).

## Done = report
Files changed; `cd backend && go build ./... && go vet ./... && go vet -tags integration ./...` and
`go test ./internal/transcript/... ./internal/media/...` results. If the plan is ambiguous or would cause a
bug, write QUESTION and argue instead of implementing around it.
