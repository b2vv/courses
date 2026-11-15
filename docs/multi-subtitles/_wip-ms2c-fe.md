NEXT: Read repository instructions, briefing and MS-D5/MS-D10, then inspect generated DTO changes and the allowed frontend adapters before making the minimal temporary bridge.
FACT: Read repository instructions and briefing; user scope overrides broad gates/commits: frontend-only bridge, no generated edits, no e2e, no commits. Applying ponytail to reuse existing validation and query paths.
NEXT: Inspect MS-D5/MS-D10 and exact DTO/caller changes; identify every affected strict response validator before editing.
FACT: Initial working tree includes backend/generated/unrelated WIP; no hand-written frontend changes predate this run. Draft DTO is now a strict {drafts:[...]} envelope with non-null draft identity/state, per-track snapshots and no lessonId. Audio list adds multilingualDefault; shared track DTO adds multilingual (also PATCH response). Mint upload-ticket DTO is unchanged; transcript inventory is a new endpoint, not an existing consumed response.
DECISION: 2026-11-15, Codex under owner brief: Small frontend compatibility slice, existing domain empty-draft representation retained; carry nullable trackId for empty/pending placeholders, regenerate by shown track only. Disable Generate for no draft (O7 automatic recognition). MS3 tabs/polling redesign remains out of scope. Verify via adapter/query/component unit tests and requested frontend commands; independent review not claimed, delivery owner is Courses 1/user.
NEXT: Read remaining UI/tests, then implement the adapter/domain/port/query/UI bridge and shared audio schema mapping without generated edits.
FACT: Minimal caller glue is necessary in admin-draft-queries and admin course route to forward trackId instead of language, and in AudioTrackRepository/admin-audio-queries to carry the list metadata while retaining current array-based UI. No new UI or abstraction is needed.
DECISION: Keep audio query UI output as items for this bridge; repository list value carries {items,multilingualDefault} for MS3. Existing pending and stale polling behavior is preserved; it is not claimed as MS3 O7 polling implementation.
NEXT: Apply these narrowly dependent frontend changes and update affected fixtures/tests.

FACT: Pre-edit SHA-256 of dependent production files (read/rechecked before write):
- `frontend/src/modules/transcript/adapters/draft-repository.ts`: `b9df2bd32192ee994307304f498ef3b43a8a313e395e8cefad1bfc8a6a87750f`
- `frontend/src/modules/transcript/domain/draft.ts`: `d83e7fa1e49973feaddcc51b1355e7a49bb2bf78a63dead5af54e7844c3e29a3`
- `frontend/src/modules/transcript/ports/DraftRepository.ts`: `412f5a8435cec304711919585ea98c5dd04753e77bda2ffad074a13aeced899e`
- `frontend/src/modules/transcript/ui/LessonDraftField.tsx`: `143f655e8cde0af37c69f4ca5928787acdbe900b8a5516f8a45fa1bc3678918d`
- `frontend/src/app/composition/admin-draft-queries.ts`: `ea73945cb81f4fb8a2a8a7ae7ff015417eda5bcf95c159b06cf628847b9c9056`
- `frontend/src/routes/admin.courses.$courseId.tsx`: `ee8c7d5c661a5cd1f7856d530828601df4b44e872c219f63e29e469f30b7b01f`
- `frontend/src/modules/media/adapters/audio-track-schema.ts`: `afb13e95f048c9648cd282147b81117d141abb9eef5439d9302ab60499588b30`
- `frontend/src/modules/media/adapters/audio-track-repository.ts`: `478c9b0ec2d81f4f43e020d488e9fbb447d40ec02bdd68ff77310569c20cfd3d`
- `frontend/src/modules/media/domain/audio-track.ts`: `ded11ec1fe35b918567e3328e92a635ee36a6e078517aa376e9e159f4a0d43e6`
- `frontend/src/modules/media/ports/AudioTrackRepository.ts`: `7b4dc2544c461a4e218683ac7366ff0d792a9621647f85226defe5cf956606d0`
- `frontend/src/app/composition/admin-audio-queries.ts`: `e01d1d87f28fca6aec19f610548e0da3fcd222af315070f1588d139c027fdcf0`

FACT: Implemented drafts envelope validation of every item, first-item mapping/empty result, shown-track bodyless regeneration and track_not_found error bucket. Removed language Select/state. Added multilingual fields to shared list/PATCH schema and domain/list port; query exposes existing items array. No generated file written.
NEXT: Adjust fixtures and add focused wire/strict-validation regression tests, then run the requested frontend commands.
FACT: Editing `frontend/src/modules/transcript/adapters/draft-repository.ts`, pre-edit SHA-256 `6d41974a7dd739bc9d1de154ed865d2d33e88081b0d9cd5def8dc1fa8c124b69`.
FACT: Editing `frontend/src/app/composition/admin-draft-queries.test.ts`, pre-edit SHA-256 `7449f28f166fd3e73d1fa4eb65b588d650ca4f929238721eb8d034cab3d34a1c`.
FACT: Editing `frontend/src/app/composition/admin-draft-s6.test.ts`, pre-edit SHA-256 `712661fc6830a7ae9eb182e331a0f3ed59905b38c3fb6e53136b15d8f9fe6320`.
FACT: Editing `frontend/src/modules/transcript/ui/LessonDraftField.test.tsx`, pre-edit SHA-256 `47ce603906cb554644ac6e4dbaed227cebf31f15845177f0b5d0b6c0917364e5`.
FACT: Editing `frontend/src/routes/route-guards.test.tsx`, pre-edit SHA-256 `c8b83e49d14de3ff65192cd2c59dd2cd74d12713886d9dcaf34a64b54eb98ad5`.
FACT: Editing `frontend/src/modules/admin/ui/CourseEditorView.test.tsx`, pre-edit SHA-256 `32b8a4cfad1e1e8dca797ac948557194180ed8f828e43dd5bf1103814126fe45`.
FACT: Editing `frontend/src/modules/media/ui/LessonAudioTracksField.test.tsx`, pre-edit SHA-256 `3b43565fe33fd77b0db4dbce8f9eca79890bf807b81d3bdb0a60ef3ca17b549c`.
FACT: Editing `frontend/src/modules/media/adapters/audio-track-repository.test.ts`, pre-edit SHA-256 `2929ebd6a8928d2e82fd9135c0ae9ade4834c2ef2e4814a5e27f0269ea4033a9`.
FACT: Editing `frontend/src/app/composition/admin-audio-queries.test.ts`, pre-edit SHA-256 `af98210f246f989501371152b9f7bca3c54210a5c0c8d25fce96c360c25551f4`.
FACT: Editing `frontend/src/modules/media/domain/audio-track.test.ts`, pre-edit SHA-256 `b09a74b2c64dcea426b97a7543688d525be68a6e325c8a19311e16f8fcbbdaff`.
FACT: Editing `frontend/src/app/composition/admin-audio-queries.test.ts`, pre-edit SHA-256 `fc7fae73ca09cd2b31c1efd3774a7f93d0827cfcb36a8dcb857fa76156267291`.
FACT: Editing `frontend/src/modules/transcript/ui/LessonDraftField.test.tsx`, pre-edit SHA-256 `61c324b8948ec0a32010ac40a840d46d2f0e840e2d3791bdac1214a5a725b401`.
FACT: Editing `frontend/src/modules/media/adapters/audio-track-repository.test.ts`, pre-edit SHA-256 `6d47a7f53fdc9b7150ccf2ca19828ba325ac9bea15634bbe434fc1baf1868dbd`.
FACT: Updated query/UI/domain/route fixtures for required trackId and multilingual fields. Timestamp inspection found drafts use RFC3339Nano, so their validator accepts variable fractional precision (audio shared schema remains unchanged). Removed language-selection tests in favor of shown-track and disabled-empty-action assertions.
NEXT: Add draft wire tests, audio strict-field tests, and query forwarding assertion; then format/typecheck/lint/test in order.
FACT: Added draft adapter wire tests (first/empty, whole-list unknown/bad/missing fields, queued mapping, bodyless track regeneration and named errors), audio multilingual list/PATCH and strict missing-field tests, and query trackId forwarding.
NEXT: Run npm run format, npm run typecheck, npm run lint, npm test sequentially; inspect any failure before expanding scope.
FACT: Initial format invocation from repository root failed ENOENT (no root package.json), exit 254; no check passed. Correcting working directory to frontend.
NEXT: Execute requested ordered checks from frontend, watching formatter scope to preserve generated/shared work.
FACT: frontend npm run format PASS (exit 0); .oxfmtrc.json excludes generated clients/route tree. No generated formatting edits.
NEXT: Run frontend typecheck.
FACT: frontend npm run typecheck PASS (tsc --noEmit, exit 0). Formatter changed only intended hand-written frontend files/tests; generated directory excluded. Diff remains narrow forwarding/schema/fixture changes.
NEXT: Run frontend lint, then full npm test.
FACT: frontend npm run lint PASS (oxlint --deny-warnings and architecture check, 238 files, exit 0).
NEXT: Run the full frontend unit suite (npm test); no e2e or backend gate is authorized for this run.
FACT: Full npm test exited 1 immediately; raw output captured at /private/tmp/courses-ms2c-fe-tests.log. Not a passing gate.
NEXT: Inspect startup failure and rerun only after its cause is understood.
FACT: Test startup failure was sandbox listen EPERM on ::1:3000, before tests ran. Retrying the same authorized npm test with escalation for its local Rstest listener; no infrastructure/e2e operation added.
NEXT: Inspect the full-suite result and fix only failures caused by this bridge.
FACT: Completed OpenAPI diff/rg audit: (1) LessonDraft/ LessonDrafts handled by isGeneratedLessonDraft + strict envelope in draft-repository; (2) MediaAudioTrack handled by audioTrackSchema for list AND rename/PATCH; (3) MediaAudioTracks handled by audioTracksSchema; (4) MediaAudioUploadTicket/mint response unchanged, existing strict audioUploadTicketSchema retained; (5) MediaAudioRename/MediaAudioTicketRequest changes are optional request fields, omitted by bridge; (6) LessonTranscriptInventory/LessonTranscripts are new unconsumed MS3 endpoints, no existing parser; (7) existing transcript/video success schemas unchanged; (8) track_not_found is mapped for draft regeneration, generic error schema accepts code strings. No other changed consumed response schema found.
NEXT: Await full npm test and record final evidence and exact changed-file inventory.
FACT: Full frontend npm test completed with exit 0 after local-listener escalation; raw output /private/tmp/courses-ms2c-fe-tests.log. Ordered final source checks so far: format 0 → typecheck 0 → lint 0 → npm test 0. Initial root-format and sandbox-test failures are recorded above, not counted as passes.
NEXT: Inspect aggregate counts/final diff, record changed files and explicit bridge limitations, then hand back with no commit.

FACT: Final verification on ms2-3-asr-per-track @ 6d11e76d00d508e376910a9498426cb2bbe302df + existing shared backend/generated WIP + this frontend diff:
- `cd frontend && npm run format` — PASS, exit 0, 394 files scanned; generated paths excluded by existing formatter config.
- `cd frontend && npm run typecheck` — PASS, exit 0.
- `cd frontend && npm run lint` — PASS, exit 0; architecture boundaries checked, 238 files.
- `cd frontend && npm test` — PASS, exit 0 after local-listener escalation; 102 test files, 976 passed, 0 failed/skipped/todo, 9474 ms. Raw log: `/private/tmp/courses-ms2c-fe-tests.log`.
- `git diff --check -- frontend ':!frontend/src/adapters/api/generated'` — PASS, exit 0.
FACT: Changed files in this run (21 frontend files; final SHA-256 for exact handoff scope):
- `frontend/src/app/composition/admin-audio-queries.test.ts` — `fadc145811156aba2063eba8f0fca24b9fcec1b1432653e0d7f695af2f2933da`
- `frontend/src/app/composition/admin-audio-queries.ts` — `e4ea8abb94494305090c71cfe29ae9a8bea13deb01871737c019168607d9a598`
- `frontend/src/app/composition/admin-draft-queries.test.ts` — `6174b24436eaa849c15c65faac3803feeae5bdadee9368db3d6bc1dd15f654d6`
- `frontend/src/app/composition/admin-draft-queries.ts` — `d5ca84347a04157a6118ee224b5b963489a8c37e5a9889141532f3103661572d`
- `frontend/src/app/composition/admin-draft-s6.test.ts` — `b6ff5fd6c80de1380757c0f0084047ccd4c9797cd21ff9e8e7e5a20a14ed2c20`
- `frontend/src/modules/admin/ui/CourseEditorView.test.tsx` — `2155dae6692e850c497becedccdd9950a5ade40adf1e31059d31dea0385cdf71`
- `frontend/src/modules/media/adapters/audio-track-repository.test.ts` — `9921d7d7e780c5970539b3946441c75f1d758a2ffeacda0fadd9b0e1b37d5529`
- `frontend/src/modules/media/adapters/audio-track-repository.ts` — `ed4bdfc76697b0bb883db9237acd010c9afcd66813eca1857495ed62967219ee`
- `frontend/src/modules/media/adapters/audio-track-schema.ts` — `7a3e3e2328e568b28a22c25bf2f6ff5828a37ba8357476f55e2c028278c51b04`
- `frontend/src/modules/media/domain/audio-track.test.ts` — `190f41568f88cb0bc8cda76da835487d2fc312d27e0d3aec8c1d686db314de9f`
- `frontend/src/modules/media/domain/audio-track.ts` — `998f11e211b63f2b31651a25a7a7d22ac4ff317346ec2979495ea8c8b10bdd57`
- `frontend/src/modules/media/ports/AudioTrackRepository.ts` — `fc0dbb3a69dc5bd3006f3773ad204ffe4e89db10eb2ecb238db53af1c2c719fa`
- `frontend/src/modules/media/ui/LessonAudioTracksField.test.tsx` — `0187454c4ee0cdfa4108b55c8b073f7f7fa28cc6e6db24b7d859968223d6bb1a`
- `frontend/src/modules/transcript/adapters/draft-repository.test.ts` — `e559bbb8283f501f4b8028700335fbbeac6782e16a08dacacba7eac9baa03ada`
- `frontend/src/modules/transcript/adapters/draft-repository.ts` — `18507d9a85b31f48e2014bc8cd5f6835072ef2e2bd0f5d411ae246f984d907c7`
- `frontend/src/modules/transcript/domain/draft.ts` — `0bc104a37c1cebb043cdb199fa45ab4d1a86db600c21957185a65e6c34edcd3a`
- `frontend/src/modules/transcript/ports/DraftRepository.ts` — `ccb60af429fe043a55057bad4c0968530e9405d02ed055cb2948731f9eb6f48c`
- `frontend/src/modules/transcript/ui/LessonDraftField.test.tsx` — `0d717c1077375b8b6e0a73e27c774158f70ddfb3fd02a103a8f5f57fcf766f14`
- `frontend/src/modules/transcript/ui/LessonDraftField.tsx` — `2a4163606d52c28c1edd4cf373cc4ef62041c0e53c30421974bd2a1770606212`
- `frontend/src/routes/admin.courses.$courseId.tsx` — `56a423a2cacfe0a32544b9e9046524b4439391d1a9f44231880000b1d4b71b36`
- `frontend/src/routes/route-guards.test.tsx` — `50b412963f24db29fa929a9b443ef0f28ad850c5a6b63b5a32a8fd33b8b91ceb`

FACT: Working record is this append-only file. Existing backend, generated, configuration and unrelated docs WIP preserved; no generated/e2e edits, git checkout/reset/stash/clean/rm, commits or service mutations performed.
DECISION: Temporary bridge limitations, as scoped by owner brief: only drafts[0] is displayed after validating every item; no tabs or multilingual switch. The audio repository list retains both items and multilingualDefault; current UI query intentionally exposes only items. Existing stale banner copy and bounded stale polling (5 s interval, 15 min window) remain for MS3 to replace with O7-aware per-track UI. This run proves frontend compile/unit compatibility only; backend/integration/build/e2e/browser and independent review were not run or claimed.
QUESTION: None; the bridge required only existing caller forwarding and list-envelope propagation, not a broader feature implementation.
NEXT: Courses 1/user reviews this dirty frontend delta together with run 1/generated changes. MS3a replaces the default-first bridge and owns tabs, per-track polling/markers, O7 copy and multilingual controls; no commit was made.
