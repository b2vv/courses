NEXT: Read repository instructions, briefing, workflow adapter, and accepted MS2–3 plan before implementing MS2a within the authorized files.
FACT: Branch ms2-3-asr-per-track at 011ff68 confirmed. Existing unrelated dirty/untracked files inventoried and preserved.
DECISION: Implementation author = this Codex session; review/gates/commit = existing separate Courses 1 session (owner instruction; no subagents or cross-session sends). Author build/vet/scoped unit checks only are authorized here. Applying local ponytail skill for minimal existing-flow changes.
NEXT: Read exact accepted decisions and current media/transcript storage-to-HTTP flow; no generated OpenAPI or frontend edits.
FACT: Current draft repository is handwritten pgx (no sqlc); media audio reads share audioTrackColumns/scanAudioTrack. Migration README has no registry. Current Claim prechecks then inserts by lesson/default leaf, which must become an atomic track+asset INSERT SELECT.
QUESTION: Legacy default-only Regenerate needs a default track ID even when that track has no draft; ListLessonDrafts cannot return it, and LessonTrack requires a track ID.
DECISION: Preserve legacy endpoints with the smallest storage compatibility convention: LessonTrack(ctx, lessonID, "") selects the latest Ready asset's default ready track; a nonempty ID retains the strict per-track contract. Document this temporary convention in the port until MS2c replaces the default-only endpoint. No new port or hint source.
NEXT: Implement migration, media storage field, transcript domain/ports and atomic postgres adapter; then compile through application/queue/HTTP.
FACT: Added migration 00031 (backfill/drop unmatched, track PK/cascade/snapshots, deterministic test-only Down), media Multilingual persistence at all three INSERT sites and common read mapper, and transcript track identity/snapshot validation + new staleness predicate.
QUESTION: 00008's language CHECK combines tag grammar with the special detected-output sentinel ban (`auto`). Track tags already permit `auto` under 00030 and NewLanguage.
DECISION: Interpret “same grammar” as the shared tag grammar + lowercase, allowing und-family; do not copy the detected-output-only `auto` ban onto track snapshots. This preserves all valid existing track tags and matches NewLanguage validation; detected draft Language still refuses auto unchanged.
NEXT: Replace draft ports and postgres implementation, preserving WHY comments and token fencing.
FACT: Ports now expose ListLessonDrafts, track-aware pending/Claim and LessonTrack; SaveDraft removed while token/failure WHY comments retained and updated. Postgres uses one atomic track+asset INSERT SELECT, requested-only conflict replacement, post-write refusal diagnosis, current-asset draft listing, per-track pending and token-fenced completion/recovery. Removed dead readyAudioForLesson.
NEXT: Compile through legacy default-only Drafting, recognize jobs, asynq payload/TaskID and publish removal; update unit fakes/fixtures only to the new contracts.
FACT: Application now claims/enqueues/completes/recovers by TrackID while retaining job.Language as the existing hint source (MS2b deferred). Legacy GET selects IsDefault draft; legacy Regenerate selects default LessonTrack and uses that track's token. Removed Publish/TranscriptPublisher constructor parameter and API wiring, HTTP publish handler/route, and its Go OpenAPI contract entry (necessary to compile; no generation run).
NEXT: Adapt existing unit fakes/fixtures and obsolete staleness assertions; remove only tests of deleted Publish code. Then run authorized build/vet and scoped unit checks.
FACT: Removed tests only for deleted Publish code in backend/internal/transcript/application/draft_test.go: TestPublishReadyDraft, TestPublishNoDraft, TestPublishNonReadyDraft, TestPublishTokenConflict, TestPublishForbidden, TestPublishFirstTranscript.
FACT: Removed tests only for deleted Publish code in backend/internal/transcript/adapters/http/draft_handler_test.go: TestC05PublishRefusesStaleToken, TestPublishSuccessReturnsToken.
FACT: Updated unit test files: transcript/application/{draft_test.go,recognize_test.go,recognize_s6_test.go}, transcript/domain/{draft_test.go,draft_s6_test.go}, transcript/adapters/http/draft_handler_test.go, transcript/adapters/asynq/queue_test.go. Changes are contract fakes/signatures, track/snapshot fixtures, default selection, new staleness semantics and TaskID/legacy-payload expectations. Existing non-Publish scenarios retained; no new integration tests or edits to integration tests in run 1.
NEXT: Format only edited Go files; execute authorized go build ./..., go vet ./..., and untagged transcript/media unit suite, recording exact results and any fixes.
FACT: Formatted edited Go files. Initial `cd backend && go build ./... && go vet ./...` failed before compilation because the sandbox cannot access the default ~/Library/Caches/go-build; vet did not run.
DECISION: Retry with isolated GOCACHE under /private/tmp (no permission change or repository cache), preserving the shared session's default cache.
NEXT: Run build/vet with GOCACHE=/private/tmp/courses-ms2a-go-build, then scoped untagged tests.
FACT: Isolated-cache build/vet command is still running (cold cache); no output yet. Source inspection confirmed all media audio row reads use the expanded common mapper and all INSERT paths explicitly write multilingual; no sqlc regeneration applies. No shared services or integration tests started.
NEXT: Await build/vet, then run untagged transcript/media tests and fix only contract/fixture issues within scope.
FACT: `cd backend && GOCACHE=/private/tmp/courses-ms2a-go-build go build ./... && GOCACHE=/private/tmp/courses-ms2a-go-build go vet ./...` completed EXIT=0. `git diff --check -- backend` completed EXIT=0. Runtime non-integration sources no longer reference GetLessonDraft, LessonAsset, SaveDraft, TranscriptPublisher, PublishRequest, draft/publish or readyAudioForLesson.
NEXT: Execute `go test ./internal/transcript/... ./internal/media/...` without integration tag using the isolated cache; then finish source review and changed-test inventory for Courses 1.
FACT: `cd backend && GOCACHE=/private/tmp/courses-ms2a-go-build go test ./internal/transcript/... ./internal/media/...` completed EXIT=1. All transcript packages passed; media packages passed except media/adapters/http: TestMediaOpenAPIFidelity and TestMediaHandlersMatchGeneratedOpenAPI both failed with `orphan overlay key components/schemas/PublishLessonDraftRequest` after the authorized publish-contract removal. No tests skipped or weakened. This is not an aggregate PASS.
DECISION: Preserve scope: do not edit global OpenAPI overlay/generated artifacts here. Courses 1 owns make generate and can remove the obsolete publish overlay/source entry in its writer window. Integration tests, migration populated-data proof and ordered full gates remain run 2/Courses 1 responsibilities.
NEXT: Locate exact overlay dependency for handoff; finish local comment cleanup and changed-file/test inventory. Re-run only checks affected by subsequent edits.
FACT: Exact deferred OpenAPI dependencies are backend/openapi/openapi.overlay.yaml:655 (PublishLessonDraftRequest), :662 (publishResult) and :1320 (publish path); backend/openapi/openapi.yaml still contains the generated schema/path. Courses 1 must remove obsolete overlay keys/path before make generate. Neither file was edited here.
FACT: Correction to earlier sqlc shorthand: media has sqlc for video-assets queries, but audio_tracks CRUD is handwritten and there is no generated AudioTrack model/query (omit_unused_structs=true). No audio-track query regeneration is needed. Repository generation target is `make generate` (first stage: `cd backend && go run github.com/sqlc-dev/sqlc/cmd/sqlc@v1.30.0 generate`); not run because the same target also regenerates OpenAPI/frontend outside this run's authorized scope and is explicitly owned by Courses 1.
FACT: Cleaned obsolete pre-S6 payload wording, empty publish-test heading and updated track-token WHY comments; only comments/whitespace changed after the completed checks.
NEXT: Finish immutable acceptance/source hashes, exact changed files/test names and run-2 handoff in this append-only record. No commits or shared runtime gates.

FACT: Final implementation checkpoint — author Codex /root, 2026-11-15; root /Users/strelia/projects/courses; branch ms2-3-asr-per-track; base/HEAD 011ff68 (no commits). Independent review NOT_RUN here; delivery/review/gate owner = Courses 1, existing separate session per owner brief. Diff remains in the shared working tree.
FACT: Authoritative acceptance input docs/multi-subtitles/plan-ms2-3.md SHA256 5a783119da28ce4b5cbc4dbe4ca8b88886d447b0ddeca9e29f2103faffdbf73e. User implementation brief defines the narrower run-1 scope.
QUESTION: Test-only Down can encounter default tracks on several retained video assets of one lesson.
DECISION: Prefer a default-track draft, then newest updated_at and track id as deterministic tie-breakers, matching the stated default-first/newest rule. Verification belongs to populated-data migration tests in run 2.
FACT: Changed implementation files (SHA256; excludes pre-existing unrelated dirty files and the append-only log itself):
- backend/cmd/api/main.go 92354a0ccd8dbf67e2559cb25876cbb9fb6c24aa6861392a31d8c87b8e312482
- backend/internal/media/adapters/postgres/audiotrack.go 181d704bf1fab362aef2f2d0b9cb95b1da0cd1ba2e0fb30e37e45b61fe0a339b
- backend/internal/media/adapters/postgres/publish.go 6d4fe1243b7052f55763430c3d025c3f02331c9d8677510cfdcc81d97e22841d
- backend/internal/media/domain/audiotrack.go 2dcdf80524e1a2c6513b8a476c0177176a12d10eb1a7954df7db51a3ce33c970
- backend/internal/transcript/adapters/asynq/handler.go 8b7d523ad6f2ec883ae53fd506f90348e20a06d74f6999115773c5aa4690ae09
- backend/internal/transcript/adapters/asynq/queue.go ac84fb97eacc576fe2af6e71c7f36d1810627b5ee43ae0c46f6a9ff8dcaa2493
- backend/internal/transcript/adapters/asynq/queue_test.go 5079f34baf9ffd44d1974ab24c96978583ac091704d08c48b1b55b8241965cc0
- backend/internal/transcript/adapters/http/draft_handler.go 802c96ed4d222f1934a155f170c626ab0d847d7aecd8cdfd9c4313ce02e1d58e
- backend/internal/transcript/adapters/http/draft_handler_test.go d503a12d5b8ca40dd6489832da4ce4263a637ffeeb34169c7f1413b6960b3bd0
- backend/internal/transcript/adapters/http/openapi_contract.go d0bd113e7bda2566ee2d72541e703b2adfaa686d849c785c75154f6523825eaa
- backend/internal/transcript/adapters/postgres/draft.go 4dcc449abe69a546a4e3b870450e506207613adf71e217a37031475171f669ef
- backend/internal/transcript/adapters/postgres/repository.go c17fe6df1df1ac194b0540883b0e87ea335ef57e1621eb249159809174a29f9d
- backend/internal/transcript/application/draft.go d468e0b69b6ce3015ec300e910202fe7e4a9bd5028ef45f1003429530b1f8a24
- backend/internal/transcript/application/draft_test.go 8e4e6a85f02750203349f3a37e8ca745dacf925825a0e8ad74d2d1576c0977f5
- backend/internal/transcript/application/recognize.go 7fa179c23ba82d3d56ba4518bc6d0d293eb0650dde68e22fa82d052992a4da45
- backend/internal/transcript/application/recognize_s6_test.go ac1f58c01030c5fd42a337cbc3512a30691816d1229850ecb84bf6b414b2060f
- backend/internal/transcript/application/recognize_test.go 97187a2c93eb9e068ad2fff6d3ea3725fef7482f2898fdd0a53bba6adc86f5c9
- backend/internal/transcript/domain/draft.go 8193457b4e2027daec89ec3e50593257741d00c588abe4244e9a04efda2b5c47
- backend/internal/transcript/domain/draft_s6_test.go b2efd5e23835c986e9dd0dd7d00a5f8d4b3c07773d23dc170f0263a7ddfa8437
- backend/internal/transcript/domain/draft_test.go a74393b7e1b7965bef8f7df592408cb529a153b080c85f1a6dd82663337cdba0
- backend/internal/transcript/ports/draft.go 9a236a8d6a127aca118574abfb76e3ee93b5ad0ae218212d310a663693a9f05a
- backend/internal/transcript/ports/queue.go 83cad6e6035966344b00ed98d742634aceb57b8b9e9e7bb46f10e4623e1d1715
- backend/migrations/00031_asr_draft_per_track.sql ce664227d3de50560494697ac8a270f23e20be26cfbebcb9d474e8a9d0b5db27
FACT: Directly changed/renamed test functions by file (all other test cases retained; shared fakes and base fixtures updated as listed earlier):
- backend/internal/transcript/adapters/asynq/queue_test.go: TestQueueUsesAssetGenerationAndStableMarker, TestPayloadCarriesRequested.
- backend/internal/transcript/adapters/http/draft_handler_test.go: TestRegenerateAlreadyRunningReturnsConflict.
  Removed/replaced: TestC05PublishRefusesStaleToken, TestPublishSuccessReturnsToken.
- backend/internal/transcript/application/draft_test.go: TestGetDraftFound, TestGetDraftNotFound, TestGetDraftForbidden, TestGetDraftStale, TestRegenerateEnqueues, TestRegenerateRunningCurrentDraftIsVisibleAndDoesNotEnqueue, TestRegenerateNoAsset, TestRegenerateForbidden, TestRegenerateNilDuration, TestRegenerateAutoLanguage.
  Removed/replaced: TestPublishReadyDraft, TestPublishNoDraft, TestPublishNonReadyDraft, TestPublishTokenConflict, TestPublishForbidden, TestPublishFirstTranscript.
- backend/internal/transcript/application/recognize_s6_test.go: TestProcessExtractsTheClaimedDefaultLeaf, TestProcessRefusedClaimExtractsNothing, TestEnqueuePendingJobsAreAutomatic.
- backend/internal/transcript/application/recognize_test.go: TestProcessHappyPath, TestProcessNotClaimedReturnsNil, TestProcessNoSpeech, TestProcessSourceNotFound, TestProcessEngineError, TestProcessContextCancelledDuringTranscribe, TestProcessContextCancelledDuringExtract, TestProcessDeadlineDuringExtractFailsDraft, TestProcessDeadlineDuringTranscribeFailsDraft, TestProcessSegmentLimitPermanent, TestProcessSupersededOnComplete, TestProcessNotFoundOnComplete, TestProcessReleasesAudioOnSuccess, TestProcessReleasesAudioOnEngineError, TestProcessPassesLanguageHint, TestProcessNoSpeechOutcomeFlag, TestProcessBoundaryAtExactlyOneSecond, TestProcessClaimError, TestProcessInvalidOutput, TestEnqueuePendingEnqueuesJobs, TestEnqueuePendingFailsNilDuration, TestEnqueuePendingContinuesOnQueueError, TestEnqueuePendingSkipsAlreadyRunningAndContinues, TestRecoverFailsTimedOutDraft, TestRecoverSupersededIsSilent, TestRecoverNotFoundIsSilent, TestRecoverContinuesOnCompleteError.
- backend/internal/transcript/domain/draft_s6_test.go: TestDraftStaleFollowsTrackLanguageAndMode.
  Removed/replaced: TestDraftStaleFollowsGenerationAndHeardLeaf.
- backend/internal/transcript/domain/draft_test.go: TestStalenessIsDerivedFromTrackMetadata.
  Removed/replaced: TestStalenessIsDerivedAndAgreesWithTheTranscriptRule.
FACT: Existing integration-tag tests still target the old per-lesson API and are intentionally not changed or executed in this implementation-only run. Run 2 must adapt Claim/ListLessonDrafts/LessonTrack, remove SaveDraft/Publish-only cases, and prove MS-D2/MS-D3 migration/track semantics on real Postgres. No claim of migration execution, HTTP new endpoints, full ordered gate or product acceptance PASS.
NEXT: Courses 1 reviews the exact diff and the two recorded interpretation choices (temporary default selector; tag grammar versus detected-output auto sentinel), removes stale global publish overlay keys/path, runs make generate, then run 2 adds/adapts integration proofs and executes the required gates. MS2b owns hint/multilingual engine/env wiring; MS2c owns new per-track HTTP.
FACT: Final `git diff --check -- backend docs/multi-subtitles/_wip-ms2a-impl.md` EXIT=0. MS2a implementation handoff contains 23 changed/new backend files, exact test names and hashes. No commits, branch changes, generated artifacts, integration runs or unrelated-file edits performed.
NEXT: Hand off this working file and shared diff to Courses 1 for review, overlay/generation cleanup and run-2 verification.
