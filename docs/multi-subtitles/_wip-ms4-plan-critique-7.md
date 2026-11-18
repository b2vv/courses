1. **BLOCKER — Round-6 finding 1 OPEN: the specified exception can prevent E13 recovery.**

   **Location:** `docs/multi-subtitles/plan-ms4.md:258–261`; `frontend/node_modules/hls.js/src/controller/error-controller.ts:477–504,635–641`; `frontend/src/modules/playback/adapters/hls-media.ts:76–80`.

   `new DOMException('InvalidStateError')` produces **`name: "Error"`**, with `"InvalidStateError"` as its message. I verified this with a read-only Node expression; it also follows the [Web IDL constructor](https://webidl.spec.whatwg.org/#idl-DOMException).

   hls.js checks **`error.name`** to recognize recoverable append-state errors. With no suitable alternate, the specified exception becomes fatal, stops loading, and reaches the application’s terminal failure path instead of the required held retry.

   **Required plan change:** specify `new DOMException('injected media append failure', 'InvalidStateError')`; assert its name and the resulting nonfatal `bufferAppendError` before requiring recovery.

2. **BLOCKER — Round-6 finding 1 OPEN: the 880-Hz observation still does not identify the failed payload.**

   **Location:** `docs/multi-subtitles/plan-ms4.md:250–270`; `frontend/node_modules/hls.js/src/controller/audio-stream-controller.ts:485–508`; `src/controller/buffer-controller.ts:767–795,1138–1152`; `src/controller/buffer-operation-queue.ts:24–35,134–139`.

   Selecting `de` aborts loading but does not discard already queued audio appends. A queued `en` media operation can execute after the selection; flushing can itself unblock that queue. Therefore, “first media append after selecting `de`” does not establish that the failed bytes belong to `de`.

   Both languages can have the same `tfdt`. A later successful `de` append at that timestamp, followed by 880-Hz playback, identifies the **recovered** data but does not prove that the **failed** data was the same fragment/language.

   **Required plan change:** bind the failed and recovered coded payloads—for example, retain and compare their audio-payload fingerprints—and associate that identified payload with the observed 880-Hz playback. Alternatively, specify another independent, concrete correlation to the `de` rendition. Timestamp equality alone is insufficient.

3. **NONBLOCKING — Round-6 finding 1’s outcome, timeline and delayed-queue subproblems are CLOSED.**

   **Location:** `docs/multi-subtitles/plan-ms4.md:251–269`; `frontend/node_modules/hls.js/src/controller/base-stream-controller.ts:1363–1378`; `src/controller/buffer-controller.ts:906–912,2043–2084,2159–2168`; `src/config.ts:491–492`; `frontend/src/modules/playback/adapters/hls-media.ts:51`.

   The recipe now distinguishes synchronous throws from asynchronous error/completion, includes `timestampOffset`, parses duration, and requires three observations before releasing recovery. Delaying the native call leaves the audio operation at the queue head until `updateend`; **this stack has no enabled append timeout**, because the default is `Infinity` and the adapter does not override it.

   The observation predicate nevertheless lacks an upper bound: `currentTime >= start - 0.25` also accepts observations after the failed interval.

   **Required plan change:** retain these improvements; bound failed-state observations to the identified interval, with an explicit stalled-at-start exception. Require at least one actual post-recovery observation inside it.

4. **NONBLOCKING — Round-6 finding 2 OPEN: the universal SourceBuffer explanation remains inaccurate.**

   **Location:** `docs/multi-subtitles/plan-ms4.md:96–103`; `frontend/node_modules/hls.js/src/controller/buffer-controller.ts:1789–1808`; `src/remux/passthrough-remuxer.ts:163–169`; `backend/internal/media/adapters/ffmpeg/transcoder.go:234–244`.

   Using the dedicated audio SourceBuffer closes the production video-tail counterexample. However, `SourceBuffer.buffered` does perform EOS extension among tracks **within that SourceBuffer**. This buffer contains one selected audio track, so that extension changes nothing here. See [MSE SourceBuffer attributes](https://w3c.github.io/media-source/#dom-sourcebuffer-buffered).

   **Required plan change:** qualify the explanation with the single-audio-track condition.

5. **NONBLOCKING — Round-6 finding 19 OPEN: the event descriptions still overstate their guarantees.**

   **Location:** `docs/multi-subtitles/plan-ms4.md:84–95`; `frontend/node_modules/hls.js/src/controller/buffer-controller.ts:990–1007,2043–2084`; `src/controller/buffer-operation-queue.ts:84–89`.

   Quota recovery requeues an existing operation without another `BUFFER_APPENDING`. An asynchronous SourceBuffer error is followed by `updateend`, whose `onComplete` emits `BUFFER_APPENDED`. Thus neither “before every physical append” nor “APPENDED means successful append” is correct.

   **Required plan change:** distinguish logical attempts from physical executions; retain an attempt across quota requeue and keep error status sticky through APPENDED→FRAG_BUFFERED. Include nested listener ordering in adapter tests: BufferController registers before application listeners (`buffer-controller.ts:183–197`; `hls.ts:227–229`).

6. **NONBLOCKING — Round-6 findings 3, 6, 9 and 18 CLOSED: confirmation and suppression remain sound as planned.**

   **Location:** `docs/multi-subtitles/plan-ms4.md:73–120,167–175`; `frontend/node_modules/hls.js/src/controller/audio-track-controller.ts:182–199,254–257`; `src/controller/audio-stream-controller.ts:792–807,1114–1118`; `src/controller/base-stream-controller.ts:2413–2460,2692–2718`.

   Single/default startup and synchronous remembered non-default selection can emit initial SWITCHED. Ordinary switching can emit it; rapid cancellation need not, and its payload need not identify the currently playing fragment. Reload resets selection and fragment state (`audio-track-controller.ts:70–76`; `base-stream-controller.ts:375–384`; `hls.ts:537–562`).

   The revised plan correctly avoids trusting SWITCHED, suppresses captions during selection/request uncertainty, rejects empty/error-only coverage, and treats empty-manifest startup as indefinite pending.

   **Required plan change:** none beyond findings 1–5.

7. **NONBLOCKING — Round-6 findings 4 and 5 CLOSED: retry and generation cases remain concrete.**

   **Location:** `docs/multi-subtitles/plan-ms4.md:154–162,240–247,273`; `frontend/src/modules/playback/ui/LessonPlayer.tsx:295,349–360`; `frontend/src/modules/playback/application/recovery.ts:123–130`; `frontend/src/routes/learn.$.tsx:90,282–285`.

   E12 reserves its failure before Play and reaches the same-instance reload path. The reducer compares generations only within the current mount, so A’s generation 2 cannot reject B’s generation 1.

   **Required plan change:** none.

8. **NONBLOCKING — Round-6 findings 7, 12, 15 and 17 CLOSED: lesson-state and prediction boundaries use production facts.**

   **Location:** `docs/multi-subtitles/plan-ms4.md:128–151,175`; `frontend/src/routes/learn.$.tsx:82–90,276–307`; `frontend/src/modules/playback/ui/LessonPlayer.tsx:368–370`; `backend/internal/transcript/application/read.go:60–61`.

   Locked lessons bypass `OpenLesson`; email gating removes playback and transcript reading is refused. A mounted player does not establish Ready media, so the explicit primary/exact distinction is necessary and preserved.

   Ready publication creates an embedded Ready track or remains AwaitingAudio (`backend/internal/media/adapters/postgres/publish.go:70–91`); production deletion refuses the last Ready track (`domain/audiotrack.go:208–224`; `adapters/postgres/audiotrack.go:218–227`). The proposed lateral asset selection agrees with playback (`course/adapters/postgres/learning.go:55–63`; `playback/adapters/postgres/playback.go:17–23`).

   **Required plan change:** none.

9. **NONBLOCKING — Round-6 findings 8, 10 and 11 CLOSED: native invalidation, current-track control and transcript errors are covered.**

   **Location:** `docs/multi-subtitles/plan-ms4.md:122–126,174–184,195–201,278`; `frontend/src/modules/playback/adapters/hls-media.ts:113–140`; `frontend/src/modules/playback/ui/LessonPlayer.tsx:595–601,989–997`; `frontend/e2e/transcript.spec.ts:349–356`; `frontend/src/modules/transcript/ui/TranscriptPanel.tsx:22–24,40–46`.

   The plan supplies native valid→unavailable transitions, replaces index-zero caption control with the current element’s ref, updates the old e2e’s direct mode assignment, resets search/follow, and exposes transcript failure/retry without stopping playback.

   **Required plan change:** none.

10. **NONBLOCKING — Round-6 finding 13 CLOSED: fixture/API/CI/T0 paths remain feasible.**

    **Location:** `docs/multi-subtitles/plan-ms4.md:203–212,275–286`; `frontend/e2e/media.setup.ts:76–79,143–161`; `backend/internal/media/adapters/http/handler.go:129,367–370`.

    The third tone includes generation and export; PATCH requires both language and label. The new spec is CI-selected while ASR is conditionally excluded (`frontend/playwright.config.ts:49–60`; `.github/workflows/ci.yml:114–115`).

    T0 correctly covers the accessibility mock (`accessibility.spec.ts:294–370`), video-less learning fixture (`learning.spec.ts:95–104`), independent playback fixtures (`playback.spec.ts:56–85`), and seed without lessons/transcripts (`backend/sql/seed.sql:4–39`).

    **Required plan change:** none; E13 remains separately blocked.

11. **NONBLOCKING — Round-6 findings 14 and 16 CLOSED: casing and remembered startup remain covered.**

    **Location:** `docs/multi-subtitles/plan-ms4.md:178–193,218–220`; `backend/migrations/00030_audio_track_language_shape.sql:88–89`; `frontend/src/modules/playback/adapters/hls-media.ts:30–35,55–61,98–104`; `frontend/src/modules/playback/ui/LessonPlayer.tsx:887–891,1238–1240`.

    MS1 prevents case-only duplicates among non-failed tracks. M7 includes adapter selection and menu identity, preserves full-tag matching, and writes manifest spelling. E3 now presses Play and requires independent 880-Hz audio plus showing German cues, rather than proving prediction alone.

    **Required plan change:** none.

VERDICT: BLOCKED