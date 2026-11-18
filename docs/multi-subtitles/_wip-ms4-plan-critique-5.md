1. **BLOCKER — R4-2 OPEN: `video.buffered` does not always prove actual audio coverage.**

   **Location:** `docs/multi-subtitles/plan-ms4.md:96–103,269`; `frontend/node_modules/hls.js/src/controller/buffer-controller.ts:1271–1311`.

   The revised intersection closes the original queued-append→removal race while MediaSource is open. Its universal claim is false at EOS, however. Chromium extends each buffer’s final range to the greatest end time before intersecting; this follows [MSE §10.2](https://w3c.github.io/media-source/#htmlmediaelement-extensions-buffered) and [Chromium’s `ComputeRangesIntersection`, lines 76–125](https://chromium.googlesource.com/chromium/src/+/refs/heads/main/media/filters/source_buffer_state.cc#76).

   Counterexample: retained labels cover `[0,30)`, actual audio now covers `[0,24)`, and video covers `[0,30)`. After `endOfStream()`, `video.buffered` can report `[0,30)`. Intersecting the retained labels with it resurrects “known” audio in the removed tail. M1 can confirm there instead of reporting pending. MediaSource reaching EOS means buffering finished; it does not require playback to have finished.

   **Required plan change:** reconcile against the actual audio SourceBuffer’s ranges, with appropriate reevaluation after removal/EOS. Add the unequal-audio/video-tail case to M01 and browser evidence. Do not describe the element’s buffered ranges as an unconditional audio-presence oracle.

2. **BLOCKER — R4-4 OPEN: E13 can fail an initialization append and never exercise the failed-media-append rule.**

   **Location:** `docs/multi-subtitles/plan-ms4.md:243–249,269`; `frontend/node_modules/hls.js/src/demux/transmuxer-interface.ts:205–234`; `src/demux/transmuxer.ts:172–180`; `src/remux/passthrough-remuxer.ts:94–107,268–271`; `src/controller/audio-stream-controller.ts:945–980,1041–1051`.

   Switching tracks resets the initialization data. The controller appends that initialization segment **before** the media data, using `frag.initSegment || frag`. Consequently, “the first audio append after selecting `de`” can be an init append. E13’s counter and subsequent-append assertions can pass when initialization fails but the subsequent media append succeeds. That does not prove M1 rejects a failed media fragment’s interval.

   The exception does reach the real error path: `buffer-operation-queue.ts:115–128` catches it and calls the operation’s `onError`; `buffer-controller.ts:1015–1071` emits `BUFFER_APPEND_ERROR`. But an ordinary throwing operation is shifted off the queue, not automatically retried as the same operation. Fragment recovery runs through `audio-stream-controller.ts:853–860` and `base-stream-controller.ts:2231–2270`.

   **Required plan change:** distinguish initialization bytes from media bytes and deterministically fail a specified `de` **media** append. Record enough passive identity to establish which media append failed and which later append recovered it. Assert pending through the failed attempt and recovery through the real controller path; “another append on this buffer happened” is insufficient.

3. **NONBLOCKING — R4-1 CLOSED for the original empty/no-progress/errored-fragment counterexamples.**

   **Location:** `docs/multi-subtitles/plan-ms4.md:84–95,269`; `frontend/node_modules/hls.js/src/controller/base-stream-controller.ts:2413–2460`; `src/controller/buffer-controller.ts:1182–1193,1425–1465`.

   A no-media fragment can still reach `FRAG_BUFFERED`, but now cannot create a label without an append. No-progress and fragment-associated errors invalidate the attempt before labeling. This closes the previously specified inference from `FRAG_BUFFERED` alone. The inaccurate event descriptions still need correction in finding 19; the browser evidence remains blocked by finding 2.

   **Required plan change:** retain these guards and correct the descriptions below.

4. **NONBLOCKING — R4-3 CLOSED: E12 now reserves its initiating failure.**

   **Location:** `docs/multi-subtitles/plan-ms4.md:235–242,250`; `backend/internal/media/adapters/ffmpeg/transcoder.go:167,181–186`; `frontend/node_modules/hls.js/src/config.ts:455`; `frontend/src/modules/playback/adapters/hls-media.ts:76–80`; `frontend/src/modules/playback/ui/LessonPlayer.tsx:349–360`.

   Starting at zero and installing the fourth-main-fragment hold before Play removes the earlier dependence on a request first issued after confirmation. Earlier fragments can establish confirmation while the reserved response remains held. Its single 503 reaches the existing same-instance reload path.

   **Required plan change:** none. This is a feasible recipe, not a claim that the unimplemented test has passed.

5. **NONBLOCKING — R4-5 CLOSED: generation comparison is scoped to mount identity.**

   **Location:** `docs/multi-subtitles/plan-ms4.md:149–157,270`; `frontend/src/routes/learn.$.tsx:90,282–285`; `frontend/src/modules/playback/ui/LessonPlayer.tsx:295,359`.

   The explicit different-mount branch accepts B-generation-1 after A-generation-2. Comparisons apply only within the current mount.

   **Required plan change:** none.

6. **NONBLOCKING — R4-6 CLOSED: switching and request-in-flight states agree with D7.**

   **Location:** `docs/multi-subtitles/plan-ms4.md:107–108,165,210–212`; parent `docs/multi-subtitles/plan.md:238–241`; `frontend/src/modules/playback/adapters/hls-media.ts:63–65,98–104`.

   The plan now removes captions and shows loading from the switch while the selected language’s request is held. It no longer deliberately displays the previous language during that transition.

   **Required plan change:** none.

7. **NONBLOCKING — R4-7 CLOSED: the lesson-state boundary uses existing facts.**

   **Location:** `docs/multi-subtitles/plan-ms4.md:134–146`; `frontend/src/routes/learn.$.tsx:82–90,276–299`; `frontend/src/modules/playback/ui/LessonPlayer.tsx:370,853–855`; `backend/internal/transcript/application/read.go:60–61`.

   Locked lessons bypass the surface; email-gated reads are disabled; non-ready lessons retain primary reads; Ready lessons use prediction and subsequent selection. A mounted player component is no longer confused with an available media instance.

   **Required plan change:** none.

8. **NONBLOCKING — R4-8 CLOSED: native confirmation can be invalidated.**

   **Location:** `docs/multi-subtitles/plan-ms4.md:117–121,271`; `frontend/src/modules/playback/adapters/hls-media.ts:113–140`; `frontend/src/modules/playback/ports/PlaybackMedia.ts:20–26`.

   Initial selection, load invalidation, valid→missing and valid→empty transitions are specified. The unit-only native coverage matches the parent’s explicit limitation.

   **Required plan change:** none.

9. **NONBLOCKING — R4-9 CLOSED: initial selection/reload facts support the revised boundary.**

   **Location:** `docs/multi-subtitles/plan-ms4.md:45–52,73–108,208–209,227–228`; `frontend/node_modules/hls.js/src/controller/audio-track-controller.ts:182–199,254–257`; `src/controller/audio-stream-controller.ts:485–506,792–807,1114–1118`.

   Single/default selection and a remembered non-default selected synchronously in `AUDIO_TRACKS_UPDATED` can produce initial `AUDIO_TRACK_SWITCHED` when the first playing fragment changes from no previous fragment. An ordinary `nextAudioTrack` switch can produce it when the playing level changes; rapid cancellation need not.

   Reload resets selection and playing-fragment state (`audio-track-controller.ts:70–76`; `base-stream-controller.ts:375–384`; `hls.ts:537–562`). The plan correctly avoids treating SWITCHED as sufficient audible-language confirmation.

   **Required plan change:** none.

10. **NONBLOCKING — R4-10 CLOSED: re-keying includes current-track mode control and follow reset.**

    **Location:** `docs/multi-subtitles/plan-ms4.md:173–179,190–194,255,276–278`; `frontend/src/modules/playback/ui/LessonPlayer.tsx:595–601,992–997`; `frontend/src/modules/transcript/ui/TranscriptPanel.tsx:22–24`; `frontend/e2e/transcript.spec.ts:349–356`.

    Re-keying alone would leave the existing index-zero effect unsafe. The specified current-element ref, language dependency and removal of test-side mode assignment address that problem. Search and follow reset are both present in acceptance.

    **Required plan change:** none.

11. **NONBLOCKING — R4-11 CLOSED: transcript errors have a concrete recovery path.**

    **Location:** `docs/multi-subtitles/plan-ms4.md:169,190–195,229–230`; `frontend/src/routes/learn.$.tsx:180`; `frontend/src/modules/transcript/ui/TranscriptPanel.tsx:12–17,40–46`.

    The proposed error variant, retry callback and playback-continuation journey replace the current discarded query error and absent panel error state.

    **Required plan change:** none.

12. **NONBLOCKING — R4-12 CLOSED: zero Ready tracks is outside the ordinary lifecycle.**

    **Location:** `docs/multi-subtitles/plan-ms4.md:55–58,114–115,294`; `backend/internal/media/domain/audiotrack.go:208–224`; `backend/internal/media/adapters/postgres/audiotrack.go:218–227`; `backend/internal/media/adapters/postgres/publish.go:70–91`.

    Production deletion refuses removal of the last Ready track. Video publication enters Ready together with an embedded track, otherwise AwaitingAudio. The unexpected-empty-manifest limitation is therefore properly distinguished from a normal non-ready lesson.

    **Required plan change:** none.

13. **NONBLOCKING — R4-13 CLOSED: fixture/PATCH/CI/T0 routes are concrete.**

    **Location:** `docs/multi-subtitles/plan-ms4.md:198–207,252–263`; `frontend/e2e/media.setup.ts:76–79,143–161`; `backend/internal/media/adapters/http/handler.go:129,367–370`.

    The third tone includes generation and export; PATCH renames an existing track and supplies both required fields. Chromium selects the new spec, while CI excludes ASR conditionally (`frontend/playwright.config.ts:49–60`; `.github/workflows/ci.yml:114–115`).

    T0 correctly distinguishes seed data without lessons (`backend/sql/seed.sql:4–39`), separate playback lessons (`frontend/e2e/playback.spec.ts:56–85`), mocked accessibility media (`accessibility.spec.ts:294–370`) and video-less learning lessons (`learning.spec.ts:95–104`).

    **Required plan change:** none to this inventory; E13 remains separately blocked.

14. **NONBLOCKING — R4-14 CLOSED: case-insensitive identity covers the remaining consumers.**

    **Location:** `docs/multi-subtitles/plan-ms4.md:177–188`; `backend/migrations/00030_audio_track_language_shape.sql:88–89`; `frontend/src/modules/playback/adapters/hls-media.ts:30–35,98–104`; `frontend/src/modules/playback/ui/LessonPlayer.tsx:887–891,1238–1240`.

    The index prevents case-only duplicates among non-failed tracks. M7 includes `shownAudio` and selection, preserves full-tag matching, and stores manifest spelling. M6 preserves the existing `captionsOn` value through whole-preference writes (`application/preferences.ts:78–85`).

    **Required plan change:** none.

15. **NONBLOCKING — R4-15 CLOSED: primary reads remain enabled independently.**

    **Location:** `docs/multi-subtitles/plan-ms4.md:139–145,170`; `frontend/src/app/composition/transcript-queries.ts:14–15,30–36`; `frontend/src/routes/learn.$.tsx:180`.

    Explicit off/primary/exact modes preserve the language-less request for non-ready lessons. The language-required condition is now scoped to exact mode.

    **Required plan change:** none.

16. **NONBLOCKING — R4-16 CLOSED: E3 proves remembered playback beyond prediction.**

    **Location:** `docs/multi-subtitles/plan-ms4.md:213–215`; `frontend/src/modules/playback/adapters/hls-media.ts:55–61`; `frontend/src/modules/playback/ui/LessonPlayer.tsx:548–550`; `frontend/e2e/audio-tracks.spec.ts:164–201`.

    E3 presses Play and requires confirmation, independently measured 880-Hz audio and showing `de` cues, while checking for unwanted `en` requests/text from navigation onward.

    **Required plan change:** none.

17. **NONBLOCKING — R4-17 CLOSED: the backend selector contradiction is removed.**

    **Location:** `docs/multi-subtitles/plan-ms4.md:123–132,272`; `backend/internal/course/adapters/postgres/learning.go:55–63`; `backend/internal/playback/adapters/postgres/playback.go:17–23`.

    Extending `current_video` reuses the selector that already agrees with playback whenever a Ready asset exists. Acceptance covers newer processing/evicted assets and actor-dependent locked output.

    **Required plan change:** none.

18. **NONBLOCKING — R4-18 CLOSED: empty-manifest evidence remains correctly narrowed.**

    **Location:** `docs/multi-subtitles/plan-ms4.md:114–115,294`; `frontend/node_modules/hls.js/src/controller/audio-track-controller.ts:160–162`.

    Startup with no audio tracks emits no `AUDIO_TRACKS_UPDATED`. The plan explicitly retains indefinite pending for this out-of-invariant state instead of claiming a fabricated empty-list event reproduces startup.

    **Required plan change:** none.

19. **NONBLOCKING — Correct two event claims without weakening the error guard.**

    **Location:** `docs/multi-subtitles/plan-ms4.md:84–95`; `frontend/node_modules/hls.js/src/controller/buffer-controller.ts:990–1007,2043–2084`; `src/controller/buffer-operation-queue.ts:84–89,115–139`.

    `BUFFER_APPENDING` is not emitted before every physical retry: quota recovery reinserts the existing operation directly. Also, `BUFFER_APPENDED` does not fire exclusively on success: asynchronous SourceBuffer error handling invokes `onError`, then the following `updateend` invokes `onComplete`, which emits APPENDED.

    These facts do not independently defeat the combined rule if the original quota attempt remains open and an errored attempt cannot be reopened by APPENDED. They do defeat the descriptions currently used to justify it.

    **Required plan change:** distinguish a logical fragment attempt from queued append executions; explicitly retain the attempt across quota requeue and reject error→APPENDED→FRAG_BUFFERED as success. Include the real listener ordering in the tests rather than assuming a flat APPENDING→ERROR sequence.

VERDICT: BLOCKED