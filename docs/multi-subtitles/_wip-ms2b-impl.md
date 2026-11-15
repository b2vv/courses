NEXT: Read repository instructions, briefing, workflow adapter, and accepted MS2–3 plan; inspect MS2b call sites before editing.
FACT: Read repository instructions and initial briefing/workflow/plan material; MS2b scope is application, queue, Whisper boundary and env default only. No commit or Docker gate is authorized for this run.
DECISION: Reuse existing ports/adapters and stdlib; retain the current HTTP contract while removing language from the application/job path.
NEXT: Inspect complete relevant plan decisions and implementation/test call sites, then update one logical slice at a time.
FACT: Root /Users/strelia/projects/courses, HEAD 4e789d2; initial tracked WIP is .gitignore, ai-factory-package.json and security WIP notes, plus unrelated untracked docs. These are preserved. Existing queue code already has track in TaskID, trackId payload validation, and MaxRetry(0); tests cover marker stability and decoder rejection but not two-track IDs, handler non-invocation or retry option.
FACT: Local AI Factory metadata is concurrently dirty and records commit=null/UNVERIFIED_FOR_CURRENT_PACKAGE, unlike adapter's historical pin. No shared-package verification/PASS is claimed; explicit owner implementation brief and accepted plan govern this bounded run.
DECISION: 2026-11-15, Codex: READY for the accepted MS2b implementation slice (resume/small execution within the existing phase). Acceptance: M2-R5/R6/R11 and scoped R12; author checks are the requested host Go commands only. Independent review/gates/commit owner: existing Courses 1 session (exact session reference pending owner handoff), manual delivery through this file; no reviewer spawn or shared heavy slot.
NEXT: Change transcript port/application/job payload together, preserve legacy HTTP decoding/error codes with an MS2c comment, then implement Whisper mapping and tests.
FACT: Updated transcript port to accept track language + multilingual value, Process to use claim snapshot, and removed job/payload/regenerate override. Legacy HTTP decoding still validates its existing body and codes; an MS2c comment marks its removal. Updated affected fakes/calls, including the integration-only transcriber signature.
FACT: Whisper now derives a supported primary-subtag hint or multilingual flag and logs unsupported tags at Warn. Python now forwards explicit multilingual and language=None in that mode.
NEXT: Add adapter argument/log and real-Python boundary tests, application claim-snapshot proof, and queue behavioral assertions.
FACT: Replaced the old override-oriented Process test with both-mode claim-snapshot assertions (engine inputs and completed row). Added queue tests observing actual enqueue TaskID options for two tracks, payload language absence, MaxRetry(0), and handler rejection without process invocation; existing marker/decoder tests remain.
NEXT: Exercise real transcribe.py with a stdlib stub, then implement explicit env-default plumbing in media constructors and repository mint.
FACT: Added M2-R6 subprocess-argv/log table for all requested tags and multilingual de, and a count/uniqueness check for the verbatim 100-code list. Added M2-R12 real transcribe.py test using exec.LookPath(python3) with fatal-on-absence, PYTHONPATH-first stdlib stub, isolated JSON captures, and three explicit kwargs cases. Stub bytecode writes are disabled.
NEXT: Add strict ASR_MULTILINGUAL_DEFAULT parser and pass its value through both composition roots, media constructors, embedded creation and postgres mint.
FACT: Added strict true/false/empty/unset config parsing with named ASR_MULTILINGUAL_DEFAULT errors and table coverage. API and media-worker parse before opening dependencies and pass an explicit bool into media application. Embedded creation carries the default; audio mint passes it through the port to postgres INSERT $8. Updated only compile-required fake/constructor/integration call sites; no media HTTP implementation changes.
DECISION: 2026-11-15, Codex: the regenerate-masters maintenance helper supplies false to NewAudioTracks because that command only rewrites existing manifests and never mints tracks (check: regenerateMasters → RegenerateAllMasters). Existing track values are not changed by it.
FACT: Added .env.example default/comment and compose api/media-worker environment entries using ${ASR_MULTILINGUAL_DEFAULT:-false}.
NEXT: Add both-value media application tests, format touched Go files and run the exact requested host build/vet/unit checks.
FACT: Added media application tests that execute Process and MintAudioTicket with both false/true defaults and assert the embedded published row and mint port/returned track values. Formatted the touched Go files only.
NEXT: Run go build ./..., go vet ./..., go vet -tags integration ./..., then scoped Go tests (including real Python boundary); record exact exits and any repairs.
FACT: First requested build chain exited 1 before compilation because sandbox denies the default /Users/strelia/Library/Caches/go-build cache; vet steps did not run. This is not a compile verdict.
DECISION: Use a task-local GOCACHE under writable /private/tmp without requesting broader permissions or modifying the repository environment.
NEXT: Retry the same ordered build/vet chain with GOCACHE=/private/tmp/courses-ms2b-go-cache.
FACT: Build/vet retry is running with the writable task-local cache. While compilation runs, reviewing the scoped diff and signature propagation; unrelated initial WIP remains outside the edit list.
NEXT: Inspect command completion and repair only concrete compile/test failures; run focused tests after build/vet succeeds.
FACT: Diff inspection found an obsolete auto-language test assertion referencing the removed job field and two new media tests referencing a nonexistent clock symbol. Corrected the first to assert the selected default TrackID and supplied explicit fixed UTC clocks; no acceptance requirement changed.
NEXT: Await compilation and rerun the ordered chain on the repaired snapshot if needed, then execute scoped tests.
FACT: Confirmed current branch ms2-3-asr-per-track. Scoped git diff --check passed. Hint code list is constant (no global mutable state); script stub uses stdlib only; real Python test has no skip path. No openapi/overlay/frontend or media HTTP implementation edits.
NEXT: Finish host verification and record explicit remaining Courses 1 integration/full-gate/review responsibilities.
FACT: Build/vet process remains running without diagnostic output on the fresh cache; no success claimed yet. A pre-existing/unrelated untracked backend/upload binary is left untouched (no removal or commit).
NEXT: Collect the final chain exit and run the scoped unit suite with the same task-local cache.
FACT: Ordered host chain completed exit 0 with GOCACHE=/private/tmp/courses-ms2b-go-cache: `cd backend && go build ./... && go vet ./... && go vet -tags integration ./...`. No compile/vet failures. The integration-tag command compiled/vetted tests; it did not execute database/Redis integration.
NEXT: Run `go test -v ./internal/transcript/... ./internal/media/... ./platform/config/...` and inspect M2-R12 plus package results explicitly.
FACT: Scoped unit suite is executing with verbose output captured in /private/tmp/courses-ms2b-tests.log. Integration execution, Docker/bookworm run and full ordered make/e2e gates remain NOT_RUN by this session; Courses 1 owns those checks and independent review.
NEXT: Inspect final unit exit and M2-R12 subtest results; finalize changed-file inventory and handoff.
FACT: Interim test output shows queue, legacy transcript HTTP and media-source adapter packages passing. Final suite status is still pending; no test failures reported in the inspected output.
NEXT: Capture final package results and explicit real-Python boundary execution proof.

FACT: Final host scoped unit command `cd backend && GOCACHE=/private/tmp/courses-ms2b-go-cache go test -v ./internal/transcript/... ./internal/media/... ./platform/config/...` exited 0: 14 tested packages passed. Raw verbose evidence: `/private/tmp/courses-ms2b-tests.log`. No failed/skipped M2-R12 test.
FACT: M2-R12 `TestRealTranscribePythonModelBoundary` PASS (0.19s): multilingual PASS (0.07s), regional en-US PASS (0.06s), undetermined PASS (0.06s). Actual kwargs prove true/null, false/"en", false/null. M2-R6 argv/log table and 100 unique codes passed; M2-R5 two enqueue IDs and M2-R11 invalid-task/non-invocation/MaxRetry(0) passed; claim-snapshot and both media-default/config tables passed. Legacy HTTP package tests passed.
FACT: Final scoped git diff --check and gofmt -l passed; current HEAD 4e789d29352119918e54de2c05b25701f41dedbc. All three build/vet checks passed on this implementation; no product code changed after those checks completed.
FACT: Changed files in this implementation (32 product/test/config files, plus this append-only working file); SHA-256 pins the reviewed handoff snapshot:

- `.env.example` — `ca0673c2b0fda810bab0600b7f0f78b0271890e23f4267b329057be469cbc689`
- `backend/cmd/api/main.go` — `00684bc8f22da258fd1ee3c51a60267a7d75074b63cf81fbd27436b3033dec1f`
- `backend/cmd/media-worker/main.go` — `aadd554127c05dd93d9d05fcc6b2fa221283e6ecf6b665f1d55a56f396d9b32b`
- `backend/cmd/media-worker/regenerate.go` — `096f8f39ac1cad7b120131ba7f82cf5d2dc4293c85fc44708f5e801c91bcc35b`
- `backend/internal/media/adapters/http/handler_test.go` — `9764cfb05336790a497c3d5f98f1afbc4f936abd16066218694a7b453e237009`
- `backend/internal/media/adapters/postgres/audiotrack.go` — `463a853d7ad1549389a2332d5d483628d3311f5d18871b4d9d066bee7a4c34dd`
- `backend/internal/media/adapters/postgres/repository_integration_test.go` — `4bb60520cfcd50db65706d3902a960e2feacc403212ebc17d92599593587b1de`
- `backend/internal/media/adapters/tus/server_test.go` — `b5d3b26b99a44cbc59cda9ca8c5199ae9532246278b07083941dd81b2c66ee87`
- `backend/internal/media/application/audiotrack.go` — `d23cea6f1cfa4bde668fa4124a6e3ca42e5c96c95a1905b549d16367edb6eaba`
- `backend/internal/media/application/regenerate_test.go` — `51510b4c3fe8ea2e788be623f1e7a758b58a8c53a8dbd3497c877d13dd117834`
- `backend/internal/media/application/transcode.go` — `53227ce6c0fb0d412262b90fd827c7c3b5dc0869c7c7069af8f636b59ea2c2cd`
- `backend/internal/media/application/transcode_test.go` — `9829bd84cd33841ef1470671f955073b2d65ddacd64a226fcb2f324e5291abfc`
- `backend/internal/media/application/upload_test.go` — `e8cab16dec3f93b832ab544edafdf3d1234a47f2ce145d684807d110a4399a87`
- `backend/internal/media/ports/repository.go` — `d7c8801073b84ed36823e14062893ebee2254c4f06bb96951bff2c87c38a0079`
- `backend/internal/transcript/adapters/asynq/handler.go` — `5f7d18ea6f483f669d1636cc1ae4956805bb6636150d6094bb2e681d717b8072`
- `backend/internal/transcript/adapters/asynq/queue.go` — `6e87a51b720609beadad261f0539ab3dde3e43ee5c71c349b9502491110f9c79`
- `backend/internal/transcript/adapters/asynq/queue_test.go` — `711c5e086642a9f7659698590ed8549a026dbd38249aa5bed8d5332e29b22d96`
- `backend/internal/transcript/adapters/http/draft_handler.go` — `42702cab1526cc3960a0c4d74023347b37c21766d841365befcc69fc63c7fc94`
- `backend/internal/transcript/adapters/postgres/draft_integration_test.go` — `60fd5ef81abc143c41248591bea69eee1ff0ec215765f11686686df503905e71`
- `backend/internal/transcript/adapters/whisper/engine.go` — `f679d55f701a2a34491a621c52e5b5184a5ca2f039a6d529686d8f1625e98513`
- `backend/internal/transcript/adapters/whisper/engine_test.go` — `3dba50a938dc70f09efbf8372a673f4210a1943be7f942dbb9d6e3cc015a116c`
- `backend/internal/transcript/adapters/whisper/testdata/faster_whisper/__init__.py` — `b073e6d10a2d8bb34b637409f50077d5569866173f1d1a3af6814b8bab8fce2f`
- `backend/internal/transcript/adapters/whisper/transcribe.py` — `7ed1fe7a362defd894adf7a9556d6abc18d51e7a2f67dadf289422c17b4b8824`
- `backend/internal/transcript/application/draft.go` — `765e9e917aafb61402c2ae99eb5cf0bb5d07af7a202833ee8ef2b8b5f3bf32f2`
- `backend/internal/transcript/application/draft_test.go` — `28071904550537050fd29c2b2875f63173935a798ec2cc121a69020be5fcb278`
- `backend/internal/transcript/application/recognize.go` — `ecdf9f48333e3037e87522d533844b89b5f95cb770027bf08b6a9ef975e45c75`
- `backend/internal/transcript/application/recognize_test.go` — `b0fbf23a1c8d96cddc0f3f9b12503344f38b6260da40a0bdf99a8a847c19f97a`
- `backend/internal/transcript/ports/queue.go` — `b4d489edb7becb025071212e113ee43afba14be7a2f792126bc429b246a3e442`
- `backend/internal/transcript/ports/transcriber.go` — `7f61d3d0c84af7ca5863e8d6ccb9347266f14de5ae4c4ebc8360ba5aadf53c1a`
- `backend/platform/config/asr.go` — `589083cd17faa280f56e46d695eb970a13bdc4d69fcf4ffea64d338b6f410467`
- `backend/platform/config/asr_test.go` — `d469ace3a0968e6f2c24dfb2d673e66e7d495f54cc5c51324e084a01c9694a2f`
- `docker-compose.yml` — `668412d415b1fc947a4bb0e69bd5756cc70e5434bac136447fd4857c3f4babb2`

FACT: Authoritative plan SHA-256: `e27efa6a519ba33074e80f6b3a4844f679ab04a05a1984bf1dbb4821e13209b4`. Existing TaskID shape/MaxRetry/decode validation were reused; new queue tests add only the missing behavioral proofs. No commit, checkout/reset/stash/clean/rm, Docker, integration execution, frontend, generated contracts, media HTTP implementation or unrelated WIP changes were performed.
FACT: Limits: host Go checks are PASS; golang:1.27.3-bookworm execution NOT_RUN, real-model inference NOT_RUN (M2-R12 deliberately stubs only the external Python model), database/Redis integration/full make check/e2e NOT_RUN, independent review NOT_TESTED here. These remain Courses 1 responsibilities, not phase closure claims.
NEXT: Courses 1 reviews this exact diff (including new config files and testdata stub), runs the authorized Docker/full ordered gates, records phase evidence and commits; MS2c replaces legacy transcript HTTP and adds media HTTP multilingual/default fields. No unresolved MS2b plan QUESTION.
