# MS2b — implementation brief for Codex (application, queue, whisper, env default)

Repository: current working directory (`courses`). Branch `ms2-3-asr-per-track` @ `4e789d2` (MS2a
committed: per-track drafts, migration 00031, `audio_tracks.multilingual`, claim snapshots
`track_language`/`track_multilingual`). Shared working tree with another session: do NOT run git
checkout/reset/stash/clean/rm, do not commit, touch only files this brief needs. Courses 1 reviews,
runs gates and commits.

WORKING FILE: docs/multi-subtitles/_wip-ms2b-impl.md
Append AFTER EVERY STEP (append, not rewrite). First entry before the first tool: a `NEXT:` line.
Then FACT / DECISION / NEXT / QUESTION. Final answer = concise summary of the file.

## Read first
- `AGENTS.md` (domain stdlib-only; modules never import each other; adapters map explicitly; no
  global mutable state; config from env in composition roots, named startup errors).
- `docs/multi-subtitles/plan-ms2-3.md` — ACCEPTED. Authoritative for this run: **MS-D4** (hint rule),
  **MS-D10** (only: env default + media application/constructors; NOT media HTTP), rows
  **M2-R5, M2-R6, M2-R11, M2-R12** (the parts listed below). MS-D5 (HTTP) is run MS2c — not here.

## Scope
1. **Transcriber port carries the track, not a hint** (`internal/transcript/ports/transcriber.go`).
   `Transcribe(ctx, audioPath, TrackLanguage domain.Language, multilingual bool)` or a small request
   struct — your choice, keep it domain types only. Rewrite the doc comment (the hint is no longer the
   author's override; it is derived from the track by the adapter, MS-D4).
2. **Application** (`internal/transcript/application/recognize.go`): `Process` passes
   `claimed.Draft.TrackLanguage` and `claimed.Draft.TrackMultilingual` (the claim snapshot — the same
   values the draft row records, so Stale is consistent with what was heard). Remove
   `RecognitionJob.Language`, the asynq payload `language` field and the `language` parameter of
   `Drafting.Regenerate` (MS-D4: Regenerate takes no body language). The existing
   `POST …/draft/regenerate` handler keeps compiling and still returns its codes, but no longer passes
   a language (the endpoint is replaced in MS2c — do not touch openapi/overlay in this run; if the
   handler's request decoding becomes dead, leave a one-line comment pointing to MS2c rather than
   changing the contract).
3. **Whisper adapter** (`internal/transcript/adapters/whisper/engine.go`):
   - hint rule MS-D4: multilingual → `--multilingual`, no `--language`; `und` or `und-*` → no hint;
     else first subtag of the track language, lowercased, if it is in the faster-whisper 1.2.1 code
     list → `--language <code>`; else no hint and log `hint_unsupported` (Warn, with the tag — a tag
     is not personal data; no audio path needed).
   - The 100-code list (faster-whisper 1.2.1 `faster_whisper.tokenizer._LANGUAGE_CODES`), verbatim:
     `af am ar as az ba be bg bn bo br bs ca cs cy da de el en es et eu fa fi fo fr gl gu ha haw he hi
     hr ht hu hy id is it ja jw ka kk km kn ko la lb ln lo lt lv mg mi mk ml mn mr ms mt my ne nl nn no
     oc pa pl ps pt ro ru sa sd si sk sl sn so sq sr su sv sw ta te tg th tk tl tr tt uk ur uz vi yi yo
     yue zh` — verify the count is exactly 100 in a unit test; comment where it came from (version +
     method: `/opt/whisper/bin/python3 -c "from faster_whisper.tokenizer import _LANGUAGE_CODES"` in the
     asr-worker image) and that an engine upgrade must refresh it.
   - M2-R6 unit (adapter): `en-US`→`en`, `de`→`de`, `haw`→`haw`, `und`/`und-x-original`→ none,
     `tlh`→ none + `hint_unsupported` logged, multilingual + `de` → `--multilingual` and no
     `--language`.
4. **`transcribe.py`**: add `--multilingual` (store_true). Call
   `model.transcribe(..., language=None if args.multilingual else args.language, multilingual=args.multilingual, …)`.
   Update the docstring contract. faster-whisper 1.2.1 `WhisperModel.transcribe` has
   `multilingual: bool = False` (it detects language per 30 s window, `transcribe.py:1191-1197`).
5. **M2-R12 — the Python→faster-whisper boundary, tested for real.** A Go test in the whisper package
   runs the REAL `transcribe.py` through `python3` (look it up with `exec.LookPath`; absent → `t.Fatal`,
   NOT skip) with a stub `faster_whisper` package under `testdata/` put first on `PYTHONPATH`
   (`t.Setenv`). The stub `WhisperModel(...).transcribe(audio, **kwargs)` writes the kwargs as JSON to a
   file named by an env var the test sets, and returns `([], info)` where `info` has `language`,
   `duration`, `duration_after_vad`. Through `Engine.Transcribe`, assert: multilingual track →
   `multilingual=True, language=None`; `en-US` track → `multilingual=False, language="en"`; `und` →
   `multilingual=False, language=None`. The test must pass both on the host and inside
   `golang:1.27.3-bookworm` (`make test` runs Go there; it has Python 3.11). Keep the stub minimal and
   do not import anything outside the stdlib.
6. **Queue units**: M2-R5 — two tracks of one asset (same generation/marker) → two distinct TaskIDs
   (`asset:generation:track:marker`); M2-R11 — a task payload without `trackId` → handler returns the
   invalid-payload error, `process` is not called, and the task is enqueued with `MaxRetry(0)` (assert
   the option you can observe). Add only what is not already covered; say what existed.
7. **`ASR_MULTILINGUAL_DEFAULT`** (MS-D10): parse in `platform/config` next to its siblings
   (`true`/`false`; unset or empty = `false`; anything else → named error, process refuses to start),
   with a table unit test. Read it in the composition roots of **api** (`cmd/api/main.go`) and
   **media-worker** (`cmd/media-worker/main.go`) and pass the bool as a value into the media
   application:
   - `NewTranscode(...)` → embedded track built in `transcode.go` `embeddedTrack` gets
     `Multilingual: <default>`;
   - `NewAudioTracks(...)` → mint passes `multilingual` down; `ports.MintAudioTrack` and the postgres
     `MintAudioTrack` take a `multilingual bool` and INSERT it instead of the literal `false`. The HTTP
     request field is MS2c: for now the application always passes the env default.
   - Add `ASR_MULTILINGUAL_DEFAULT=false` to `.env.example` (with a one-line comment) and to the
     `environment` of the `api` and `media-worker` services in `docker-compose.yml`
     (`${ASR_MULTILINGUAL_DEFAULT:-false}`, matching how neighbours are written).
   - Unit tests: transcode embedded track and mint carry the configured value (both `true` and
     `false`). Update fakes/constructors elsewhere only as far as needed to compile.

## Out of scope
Media HTTP (POST/PATCH/GET `multilingual`, `multilingualDefault`), transcript HTTP, openapi, frontend,
integration tests (Codex cannot run Docker — but `go vet -tags integration ./...` must still pass;
fix compile breaks in integration tests caused by changed signatures, e.g. `MintAudioTrack`).

## Done = report
Files changed; `cd backend && go build ./... && go vet ./... && go vet -tags integration ./...`;
`go test ./internal/transcript/... ./internal/media/... ./platform/config/...` results (M2-R12 test
must run, not skip). If the plan is ambiguous or would cause a bug, write QUESTION and argue instead
of implementing around it.
