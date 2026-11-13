# S5-L2 frontend surface survey (read-only, 2026-11-13, HEAD 132833a on multi-audio-s5l1b-fixes)

ДАЛІ: routes under /admin and /teach for lesson video management; then tus client, Orval audio exports.

## Q1. Lesson video management UI — where it lives

ФАКТ: routes (flat file routing, `frontend/src/routes/`):
- `admin.courses.$courseId.tsx` — course editor; the only consumer of video wiring. Builds `videoActions: LessonVideoFieldActions` (`:133-164`) from `admin.videos.mint/retry/transfer/cancelUpload` mutations (`:95-98`) and `useQuery(admin.videos.list(course.id, activeTransfers > 0))` (`:100`); passes `video: { rows, isLoading, error, actions }` (`:176-181`) to `CourseEditorView` / `AdminOwnerEditor` (`:184-188`).
- `admin.courses.$courseId_.lessons.$lessonId.tsx` — lesson **document** editor only (lazy `LessonDocumentEditor`, `:13-17`, `:46`); no video.
- `teach.courses.$courseId.tsx` — instructor workspace: moderation, roster, `TeachLessonList` (`:11`, prefetches `teach.detail/courseComments/moderationQueue`, `enrollment.roster` `:38-43`); no video.
- `teach.courses.$courseId_.lessons.$lessonId.tsx` — same lesson document editor with `teach.documents` (`:46`); no video.
ФАКТ: component `frontend/src/modules/media/ui/LessonVideoField.tsx` rendered per lesson inside `frontend/src/modules/admin/ui/CourseEditorView.tsx:385-390` (`{video !== undefined && <LessonVideoField lessonId isSaved {...video}/>}`), import `:41`, prop `video?: CourseEditorVideo` `:80`.
ФАКТ: `grep -i video` over `src/modules/teach` and `src/app/composition/teach*.ts` → no matches. **Video upload exists only under `/admin`.**
ФАКТ: both layouts guard with the same predicate: `admin.tsx:10` and `teach.tsx:10` → `canAccessAdministration(user)` = `role === 'instructor' || 'admin'` (`modules/identity/domain/identity.ts:79-80`). So an instructor reaches `/admin/courses/$courseId` and the video field; `admin.courses.$courseId.tsx:184-187` branches only on owner options (`actor.role === 'admin'` → `AdminOwnerEditor`, else `CourseEditorView ownerOptions={null}`).

## Q2. Video upload client (tus)

ФАКТ: files under `frontend/src/modules/media/` (module barrel `index.ts:1-25`):
- `adapters/tus-client.ts` (288 lines) — hand-written tus 1.0.0, not Orval. `createTusTransport` (`:186-215`) does HEAD/PATCH/DELETE `fetch` with `credentials:'omit'`, `redirect:'error'`, `cache:'no-store'`, per-exchange deadline `chunkTimeoutMs = 120_000` (`:13`). `createTusClient(options, storage, transport)` (`:218-288`) → `transfer(target, file, onProgress, signal)` and `cancel(videoId, signal)`.
- Upload URL hard-coded `${apiOrigin()}/uploads/${videoId}` (`tus-client.ts:45-47`) — audio uses `/uploads/audio/{trackId}`, not covered by this function.
- Resume: localStorage fingerprint `courses.media.upload-fingerprint.v1.<videoId>` = `{originalName,size,lastModified}` (`:14`, `:49-111`); `new` stores it, `resume` requires a match else `upload_resume_unverified` (`:233-243`); HEAD → `parseTusHead` offset/Upload-Length (`:144-152`, `:245-259`); chunks via `nextTusChunk` with `videoChunkBytes` (`:170-183`, `:263-277`); `onProgress({receivedBytes, declaredBytes})` after HEAD and each PATCH (`:262`, `:275`).
- Size ceiling bound to video: `metadata.declaredBytes > maximumVideoBytes` → `invalid_request` (`:257`); `maximumVideoBytes = 2 GiB`, `videoChunkBytes = 8 MiB`, `acceptedVideoFileTypes = ['.mp4','.mov','.mkv']` (`domain/video.ts:71-73`).
- Cancel: DELETE on the tus URL, then remove fingerprint (`:281-286`). Abort via caller `AbortSignal` (`createRequestControl`, `:195`).
- Status mapping `tusStatusCode` 400→unsupported_media, 404→video_not_found, 413→upload_too_large, 503/507→dependency_unavailable (`:154-168`).
- `adapters/lesson-video-repository.ts` — the boundary where Orval JSON calls meet tus (`:70-137`): `mint` → `mintAdminVideoTicket(lessonId, {declaredBytes, originalName}, init)` expecting 201 (`:90-101`; URL `/api/v1/admin/lessons/{lessonId}/video`, `generated/courses.ts:2005-2017`); `listByCourse` → `listAdminCourseVideos(courseId)` 200 (`:102-120`; `/api/v1/admin/courses/{courseId}/videos`, `courses.ts:774-781`); `retry` → `retryAdminVideo` (`:121-129`); `transfer`/`cancelUpload` delegate to tus (`:130-135`). Wire-error whitelist per status `wireError` (`:23-57`). Zod schemas in `adapters/lesson-video-schema.ts`.
- `ports/LessonVideoRepository.ts:9-28` (mint/listByCourse/retry/transfer/cancelUpload); `application/lesson-video.ts:29-60` `createLessonVideos` validates ids/inputs before the port; `validTransferTarget` requires `ticket.uploadUrl === '/uploads/${videoId}'`, `chunkBytes === videoChunkBytes`, `maxBytes === maximumVideoBytes` (`:12-27`).
- Domain types `domain/video.ts`: `VideoState` (`:1-9`), `LessonVideo` (`:20-30`), `VideoUploadTicket` (`:32-38`), `VideoTransferTarget` new|resume (`:40-42`), `VideoErrorCode` union (`:49-65`), `VideoResult<T>` (`:67-69`), `validMediaId` (`:77-79`).
- UI flow `ui/LessonVideoField.tsx`: picker → `actions.upload` (mint then transfer) or `actions.resume` (`:167-177`); `ProgressBar` from progress (`:219-228`); interrupted (`created|uploading` without local progress) shows resume FileUpload + danger Cancel button (`:194-259`); retry button gated by `canRetry` age thresholds 10/40 min (`:12-13`, `:60-69`, `:260-273`); state shown as `<output data-video-state>` + i18n `administration.video.state.<state>` (`:214-218`); errors mapped to `administration.video.error.*` (`:75-108`).
ФАКТ: tests — `adapters/tus-client.test.ts` (372 l.), `adapters/lesson-video-repository.test.ts` (109), `adapters/lesson-video-schema.test.ts` (52), `application/lesson-video.test.ts` (77), `ui/LessonVideoField.test.tsx` (297).

## Q6. How video state is refreshed

ФАКТ: `frontend/src/app/composition/admin-video-queries.ts`:
- `videoPollMs = 3_000`, key `['admin','videos','list', courseId]` (`:13-14`, `:32-34`).
- `videoRefetchInterval(rows, transferInFlightInThisTab)` → 3000 ms while any row is `queued|processing` or a transfer runs in this tab, else `false` (`:36-44`); wired as `refetchInterval: (query) => videoRefetchInterval(query.state.data, transferInFlightInThisTab)` (`:99-100`). Route passes `activeTransfers > 0` (`admin.courses.$courseId.tsx:99-100`, counter `:120-131`).
- `invalidateAfterVideoWrite` = `invalidateQueries({queryKey: list(courseId), exact:true})` (`:46-48`); called `onSuccess` of mint/retry/cancelUpload and `onSettled` of transfer (`:102-124`).
- On a ready-generation change the list queryFn also invalidates `lessonDraftQueryKey(lessonId)` (`:76-96`).
- Errors thrown as `VideoQueryError(code)` via `unwrap` (`:18-30`); route maps them back with `videoErrorCode` (`admin.courses.$courseId.tsx:42-48`); `authentication_required` clears `sessionQueryKey` (`:102-107`, `:115`).
- Wiring: `app/composition/admin.ts:35-37` → `admin.videos = createAdminVideoQueries(createLessonVideos(createApiLessonVideoRepository(options)))`; `options` from `createIsomorphicFn` (cookie on server) `:25-27`.

## Q3. Orval audio-track exports

ФАКТ: module `frontend/src/adapters/api/generated/courses.ts` (import path as used by adapters: `'../../../adapters/api/generated/courses'`); models in `frontend/src/adapters/api/generated/models/`.
| endpoint | export | line | declared statuses |
|---|---|---|---|
| GET `/api/v1/admin/videos/{videoId}/audio-tracks` | `listAdminAudioTracks(videoId, options?)`, url `getListAdminAudioTracksUrl` | 3072 / 3060 | 200 400 401 403 404 500 503 |
| POST same (mint ticket) | `postApiV1AdminVideosVideoIdAudioTracks(videoId, mediaAudioTicketRequest, options?)`, url `getPostApiV1AdminVideosVideoIdAudioTracksUrl` | 3149 / 3141 | 201 400 401 403 404 409 500 503 |
| DELETE `…/{trackId}` | `deleteAdminAudioTrack(videoId, trackId, options?)` | 3233 | 204 400 401 403 404 500 503 |
| PATCH `…/{trackId}` | `renameAdminAudioTrack(videoId, trackId, mediaAudioRename, options?)` | 3316 | 200 400 401 403 404 409 500 503 |
| POST `…/{trackId}/default` | `setDefaultAdminAudioTrack(videoId, trackId, options?)` | 3401 | 204 400 401 403 404 500 503 |
ФАКТ: the POST create has no `operationId` in `backend/openapi/openapi.overlay.yaml:1625-1690` (only list/rename/delete/setDefault have one), hence the generated name `postApiV1AdminVideosVideoIdAudioTracks`.
ФАКТ: models — `MediaAudioTicketRequest {declaredBytes, label, language, originalName}`; `MediaAudioUploadTicket {chunkBytes, expiresAt, maxBytes, trackId, uploadUrl}`; `MediaAudioTrack {isDefault, kind, label(1..64), language(1..35), state, trackId, updatedAt, videoId}`; `MediaAudioTracks {items}`; `MediaAudioRename {label, language}`; `MediaAudioTrackKind = embedded|uploaded`; `MediaAudioTrackState = created|uploading|queued|processing|awaiting_audio|ready|failed`.
ФАКТ: `grep -rn 'AudioTrack|audio-tracks|MediaAudio' frontend/src --exclude-dir=generated` → **no consumers, no Query hooks, no schema/adapter yet**.
ФАКТ (backend, for error mapping): `backend/internal/media/adapters/http/handler.go:505-536` `writeMediaError` — 409 codes are `duplicate_language`, `upload_in_progress`, `video_not_retryable`. Delete of the last Ready track (`domain.ValidateTrackDeletion`, `internal/media/domain/audiotrack.go` `:205`) and set-default on a non-Ready target (`ApplyDefault`, `:183`) return `ErrInvalidRequest` → **400 `invalid_request`**, consistent with no 409 in their declared statuses. Audio ceiling `MaximumAudioDeclaredBytes = 500 MiB` (`audiotrack.go:255`); language = 1..`MaximumLanguageLength` ASCII letters/digits/`-` (`validLanguage` `:267`); label rejects `"`, control chars (`validLabel` `:282`).

ДАЛІ: Q4 layering examples, Q5 UI primitives, Q7 e2e/fixtures, Q8 language constants.

## Q4. Layering conventions (one example per layer, from the media/video stack)

ФАКТ: feature module hexagon `frontend/src/modules/<module>/{domain,application,ports,adapters,ui}` + barrel `index.ts`; enforced by `frontend/scripts/check-architecture.mjs`: module domain may import only `contracts/` (`:54-55`), application/ports inward only (`:64-69`), cross-module only via `modules/<m>/index.ts` (`:76-86`, dynamic `import()` exempt), and `fetch` call in `routes/`, `presentation/` or any `ui` layer fails as "presentation calls fetch" (`:95-101`).
- domain: `frontend/src/modules/media/domain/video.ts` (types, limits, `validMediaId`).
- port: `frontend/src/modules/media/ports/LessonVideoRepository.ts`.
- application (validation before port): `frontend/src/modules/media/application/lesson-video.ts` `createLessonVideos`.
- adapter (Orval + zod + transport): `frontend/src/modules/media/adapters/lesson-video-repository.ts` (`createApiTransport` from `src/adapters/api/transport.ts`, zod in `lesson-video-schema.ts`).
- Query options factory (composition, not module): `frontend/src/app/composition/admin-video-queries.ts` (`queryOptions`/`mutationOptions`, error class, refetch policy).
- composition root: `frontend/src/app/composition/admin.ts:29-44` → router context `admin` in `frontend/src/router.tsx:42` (also `teach: createTeachDependencies()` `:46`, `app/composition/teach.ts:20-…`); context type in `src/routes/__root.tsx:30-33`.
- route = hooks + wiring only: `src/routes/admin.courses.$courseId.tsx` (`useMutation(admin.videos.*)`, `useQuery(admin.videos.list…)`) → passes plain data + callbacks (`CourseEditorVideo`) into ui.
- ui (props only, no Query/fetch): `frontend/src/modules/media/ui/LessonVideoField.tsx` (props `CourseEditorVideo & {lessonId,isSaved}`, `:39-42`).
- URL state: `frontend/src/app/router/admin-search.ts` (zod per-field parse `parseAdminSearch`, `adminListState`) used by `src/routes/admin.index.tsx:20` (`validateSearch: parseAdminSearch`) and `:35` (`adminListState(Route.useSearch())`); same for users `admin.users.index.tsx:24,39`. Course editor tab is not URL state in the video path (e2e clicks `role=tab` "Уроки", `frontend/e2e/playback.spec.ts:288`).
- i18n: keys `administration.video.*` in `frontend/src/app/providers/locales/uk.ts:773…` (and `en.ts`), used via `useTranslation()`.

## Q5. UI primitives (`frontend/src/presentation/ui/index.ts` barrel, React Aria)

ФАКТ exports `index.ts:1-24`: `Button, Link, TextField, TextArea, Select, ComboBox, Checkbox, Switch, FileUpload, SegmentedControl, RadioGroup` (controls), `Dialog, Menu, Tabs, Tooltip, Toast` (overlays), `Table, DataTable, ProgressBar, LoadMeter, Delta, Pagination` (data), `PageHeader, SummaryStrip`, `EmptyState, ErrorState, Skeleton`, `StatTile`, `TrendChart`. Storybook: `AdminKit.stories.tsx`.
- Language select: **no language-specific primitive**. Generic `Select({label, options: ChoiceOption[], description, errorMessage, …Aria.SelectProps})` `controls.tsx:170-200`; `ComboBox` with `filtering: 'client'|'server'` `controls.tsx:213-…`; `ChoiceOption = {id,label}` `:121`. `presentation/LanguageSwitcher.tsx` is the UI-locale (uk/en cookie) toggle, not reusable for content language (`:6-18`).
- Choosing a default item: `RadioGroup({label, value, onChange, options})` `controls.tsx:92-119` (label box `.ui-radio` is the click target, see gotcha "RAC Radio: клікай label.ui-radio"); `SegmentedControl` = ToggleButtonGroup `:134-…`. No existing "set default" pattern found.
- Confirm-before-delete: `Dialog({triggerLabel, trigger?, triggerAriaLabel?, title, closeLabel, children: node | (close)=>node})` `overlays.tsx:6-64`. Canonical usage: `modules/admin/ui/CourseListView.tsx:291-320` (danger icon Button trigger, body `administration.confirmDeleteBody`, danger confirm → `onConfirmDelete(course); close()`); also `CourseEditorView.tsx:865`, `:1016`, `RefundQueue.tsx:273`, `ParticipantReportQueue.tsx:224`.
- State badge: no Badge component; CSS `ui-pill ui-pill--{success|warning|danger|info|neutral|accent}` used as `<span className="ui-pill ui-pill--${variant}" data-status>` with a `Record<State,string>` map (`CourseListView.tsx:31-35`, `:141-148`; styles `app/styles/tokens.admin.css:156-180`). Video field instead uses `<output className="admin-video-state" data-video-state>` (`LessonVideoField.tsx:214-218`); draft field `<output className="admin-draft-state" data-draft-state>` (`modules/transcript/ui/LessonDraftField.tsx:116`).
- File picker / drop zone: `FileUpload({label, isDisabled, onDrop?, dropLabel?, …Aria.FileTriggerProps})` `controls.tsx:281-310` — trigger is a secondary Button; with `onDrop` wraps in `Aria.DropZone`, but the drop filter is **hard-coded to `image/jpeg|png|webp`** (`:300-303`). Video uses it without `onDrop` (`LessonVideoField.tsx:242-248`, `:279-289`, `acceptedFileTypes`, `allowsMultiple={false}`); image module uses it in `modules/document/ui/LessonDocumentEditor.tsx:728`.
- Progress: `ProgressBar({label, value | isIndeterminate})` `data.tsx:22-45`.

## Q7. e2e, fixtures, seed helpers

ФАКТ: specs in `frontend/e2e/`. Admin video upload: `media.spec.ts` (433 l.) describe "the first real video through the whole chain" `:220`, serial, admin storageState `:222`; tests P04/P08/B05/B25/R13 upload one-rung `:295`, B26/B27/B30 ladders + "one shared audio rendition" `:347`, R26 student direct mint refused `:391`, P05/R35 processing→ready `:420`. Helper `upload(page, lesson, file)` `:115-128`: locate `getByRole('group', {name: 'Урок N'})`, expect button "Завантажити відео", wait for `POST /api/v1/admin/lessons/{id}/video`, `locator('input[type="file"]').setInputFiles(file)`, expect 201. Progress recorder via MutationObserver on `[role=progressbar][aria-valuenow]` `:88-113`. `playback.spec.ts:276-327` `uploadReady` (goes to `/admin/courses/{id}`, clicks tab "Lessons|Уроки", uploads, polls `GET /api/v1/admin/courses/{id}/videos` for that videoId → `ready`, 240 s).
ФАКТ: fixtures are **generated, not committed** — `git ls-files` has no `.mp4/.m4a/.mp3/.aac/.wav/...`. `frontend/e2e/media.setup.ts:52-143` runs ffmpeg inside `media-worker` (lavfi `testsrc2` + `sine=frequency=440`, `-c:a aac`) producing `e2e-{640x360,1280x720,320x240,speech-640x360,1280x720-5min,1280x720-100s}.mp4` into `frontend/playwright/.fixtures/` (`:10`, `:134`). **No audio-only fixture** in e2e. Backend analogue for audio-only: `generateAACFixture` (`backend/internal/media/adapters/ffmpeg/transcoder_audio_integration_test.go:78-88`: `sine=frequency=440:duration=1`, `-c:a aac -b:a 128k -movflags +faststart -f mp4`). `.asr-assets/audio/*.wav` exists locally but is git-ignored.
ФАКТ: Playwright projects `frontend/playwright.config.ts:44-105`: `setup` (`auth.setup.ts`, `media.setup.ts`) → `chromium` (excludes media spec) → `media` (`media.spec.ts`, deps setup+chromium) → `playback-profile` → `playback` → `recovery`.
ФАКТ seed helpers: per-spec, no shared module. API seeding in `media.spec.ts:224-265` (POST `/api/v1/admin/courses`, PUT tree with lesson ids, `Origin` header; afterAll DELETE + restart media-worker `:274-292`); `compose`/`worker` exec helpers `:40-51`; SQL via `docker compose exec -T postgres psql -v ON_ERROR_STOP=1 -U courses -d courses -c` in `teach-editor.spec.ts:32-48` (similar `psql` use in accounts/learning/lesson-builder/playback/ratings specs). Accounts in `e2e/auth.ts` (`instructor@courseforge.local` `:11-13`; admin id `40000000-0000-4000-8000-000000000003` used in `media.spec.ts:27`). Contract-checked field reads `e2e/contract.ts` `contractField<T>`. Base data `backend/sql/seed.sql`, `seed-identities.sql`.

## Q8. Language list / constants

ФАКТ: no content-language list or constant in `frontend/src` (grep `language` across modules; no `Intl.DisplayNames`).
- Transcript language: `string | null` in `modules/transcript/domain/transcript.ts:20` and `domain/draft.ts:6`; wire validation zod `.min(2).max(35).regex(/^[a-z][a-z0-9]{1,7}(?:-[a-z0-9]{1,8})*$/).nullable()` in `modules/transcript/adapters/transcript-schema.ts:14-19`; default `'uk'` when absent `modules/transcript/ui/LessonTranscriptField.tsx:82`. No language input in the transcript editor (only start/end/text fields, `TranscriptEditor.tsx:247-262`); language changes only from draft/recovery offers (`:191-192`, `:227`).
- UI locale: `'uk' | 'en'` in `app/composition/locale.ts:7-11` (`normalizeLocale`), `presentation/LanguageSwitcher.tsx`.
- Backend audio language rule is looser than the transcript regex: 1..35 chars, `[A-Za-z0-9-]` (`backend/internal/media/domain/audiotrack.go:267`); generated model only says minLength 1 / maxLength 35 (`models/mediaAudioTrack.ts`).

## Гаразд / підсумок
ФАКТ: усі 8 питань покрито; жоден файл, крім цього, не змінено; команд гейту не запускала.
