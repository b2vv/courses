# MS2a — run 2 brief for Codex: one fix + integration tests

Repository: current working directory. Branch `ms2-3-asr-per-track`, uncommitted MS2a diff from run 1
(see `docs/multi-subtitles/_wip-ms2a-impl.md`) plus Courses 1's overlay cleanup and `make generate`.
Shared working tree: no git checkout/reset/stash/clean/rm, no commits. Edit only the files named here.

WORKING FILE: docs/multi-subtitles/_wip-ms2a-tests.md
Append AFTER EVERY STEP (append, not rewrite). First entry before the first tool: a `NEXT:` line.
Then FACT / DECISION / NEXT / QUESTION. Final answer = concise summary of the file.

Authoritative: `docs/multi-subtitles/plan-ms2-3.md` (MS-D2, MS-D3, §4 rows M2-R1..R4, R7, R8, R9).

## 1. Fix (review finding of run 1)
`Recognize.Process` (`backend/internal/transcript/application/recognize.go:58-66`) returns any Claim
error to the asynq handler, which logs it as `dependency_unavailable` and counts a failure. Since
MS-D3 a Claim can answer `domain.ErrDraftSuperseded` (track of a non-latest Ready asset) and
`domain.ErrNotFound` (track/lesson gone). D8-7 says both are DONE, not failed: return a zero outcome
and nil error, the same as `complete()` does at `recognize.go:142`. Add unit tests in
`application/recognize_test.go`: Claim→Superseded and Claim→NotFound return nil error, no Extract call.
Keep `TestProcessClaimError` (generic error still propagates).

## 2. Integration tests (build tag `integration`, real Postgres via the existing helpers)
Codex cannot run Docker: write the tests so they compile (`go vet -tags integration ./...` must
pass — run it), Courses 1 runs them.

Adapt to the new ports (`ListLessonDrafts`, `Claim(..., trackID, requested, now)`, `LessonTrack`,
no `SaveDraft`/`GetLessonDraft`/`LessonAsset`/`source_leaf`) every integration test that references
the old API or `asr_drafts.source_leaf`:
- `backend/internal/transcript/adapters/postgres/draft_integration_test.go`
- `backend/internal/transcript/adapters/postgres/s6_integration_test.go`
- `backend/cmd/media-worker/regenerate_integration_test.go`
- `backend/internal/media/adapters/postgres/{audiotrack,backfill,language_shape}_integration_test.go`
  (check whether they touch asr_drafts or need `multilingual` in hand-written INSERTs)
- the course-module document tests appeared in a grep for these names; confirm and leave them alone
  if they only mention unrelated words.

Rewrite S6 default-leaf subtests into their MS2 meaning instead of deleting them; delete only what is
about removed behaviour (SaveDraft), and list each deleted subtest with the reason.

Required subtests (name them with the row id):
- **M2-R1** migration 00031 on populated data, precedent
  `internal/media/adapters/postgres/backfill_integration_test.go` (UpTo/DownTo helpers):
  `UpTo(30)` → seed: lesson A with an embedded-track draft (source_leaf NULL), lesson B with a draft
  whose source_leaf = `audio-<uploaded track id>`, lesson C with a draft whose leaf matches no
  non-failed track → `UpTo(31)` → A and B mapped to the right `audio_track_id` and
  `track_language = lower(track.language)`, C deleted; PK is `audio_track_id`; deleting a track
  cascades its draft; Down to 30 then Up to 31 again succeeds; `audio_tracks.multilingual` = false
  for existing rows.
- **M2-R2** pending: two `ready` tracks of the current asset without drafts → 2 rows with distinct
  TrackID; `awaiting_audio`/`failed` track, or a track of an older Ready asset → 0; `is_default`
  irrelevant; a track that has a draft, after rename (`UPDATE audio_tracks SET language`) or
  multilingual change → 0 (O7).
- **M2-R3** claim: writes `track_language` (lowercased) and `track_multilingual` from the track;
  foreign lesson/asset/track, a non-ready track, or wrong generation → `ErrNotFound` (C21).
- **M2-R4** C09 new: claim of a track of an older Ready asset → `ErrDraftSuperseded`;
  `ListLessonDrafts` returns only the latest Ready asset's drafts; a slow Complete with an old
  token does not overwrite a re-claimed row.
- **M2-R7** stale via `ListLessonDrafts` + `Draft.Stale(current...)`: `en`→`de` stale and not
  pending; `und`→`uk` not stale; `en-US`→`en-GB` stale; multilingual flip stale; automatic duplicate
  claim (`requested=false`) does not overwrite the stale draft (`Claimed=false`, text intact);
  `requested=true` re-claims it and snapshots the new language/mode.
- **M2-R8** Complete/Recover per track token; Recover × re-Claim → Superseded (adapt existing);
  two tracks of one asset recover/complete independently.
- **M2-R9** deleting a track removes only its draft; a published transcript of its language stays.
- Keep C08, C21, C22 subtests, adapted.

## Done = report
Exact changed files, subtests added/renamed/deleted (with reasons), `go build ./...`,
`go vet ./...`, `go vet -tags integration ./...`, `go test ./internal/transcript/...` results.
If the plan or the run-1 code is wrong for a case, write QUESTION and argue instead of bending the test.
