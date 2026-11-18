1. **BLOCKER — New E13 race: fingerprint availability is not synchronized with injection.**

   **Location:** `docs/multi-subtitles/plan-ms4.md:258–265`; `frontend/node_modules/playwright-core/lib/coreBundle.js:13544–13554,59640–59641,61927–61930`; `frontend/node_modules/hls.js/src/utils/xhr-loader.ts:193–210`; `src/controller/buffer-operation-queue.ts:115–128`.

   `page.on('response')` does not pause delivery. Reading its body, hashing it, and transferring the fingerprint to the browser are asynchronous. Meanwhile, hls.js can receive, transmux and append that response. A valid execution therefore reaches every `de` append before its fingerprint enters the wrapper’s set: nothing throws, no recovery is held, and E13 fails despite correct application behavior. An asynchronous append wrapper cannot solve this: hls.js must catch the exception synchronously.

   **Required plan change:** establish an explicit barrier. For example, intercept `de` media responses, fetch and fingerprint their real bytes, register the payload/fingerprint with the browser wrapper and await acknowledgement, then fulfil the response. Specify synchronous payload recognition inside `appendBuffer`.

2. **NONBLOCKING — Round-7 finding 1 CLOSED: exception and observable recovery.**

   **Location:** `docs/multi-subtitles/plan-ms4.md:263–274`; `frontend/node_modules/hls.js/src/controller/error-controller.ts:477–504,635–641`; `src/controller/buffer-controller.ts:1020–1055`; `frontend/src/modules/playback/adapters/hls-media.ts:76–80`; `ui/LessonPlayer.tsx:251–255,857–874`.

   The two-argument constructor supplies `name === 'InvalidStateError'`, which hls.js recognizes as an append-state error. The first injected failure remains below the configured append-error threshold. Positive observation of the held retry, followed by successful recovery without the application’s terminal error view, provides observable behavioral evidence without accessing the private Hls instance.

   **Required plan change:** none beyond finding 1’s synchronization.

3. **NONBLOCKING — Round-7 finding 2 CLOSED: the fingerprint establishes payload identity.**

   **Location:** `docs/multi-subtitles/plan-ms4.md:258–281`; `frontend/node_modules/hls.js/src/demux/mp4demuxer.ts:124–140`; `src/remux/passthrough-remuxer.ts:215–217,338–339,342–347,532–548`; `src/controller/base-stream-controller.ts:1353–1378`.

   For this ordinary audio-only fMP4 path, demuxing retains the segment bytes; the passthrough remuxer returns those bytes as `data1`, with no `data2`. Its rewriting branch concerns I-frame video. The audio `mdat` payload therefore reaches append unchanged. Matching the failed and recovered payload hashes to a response from the identified `de` path closes the previous queued-`en` counterexample.

   **Required plan change:** retain this identity check; make its prerequisites deterministic as finding 1 requires.

4. **NONBLOCKING — Round-7 finding 3 remains OPEN for observation bounds; its queue/timeline fixes remain CLOSED.**

   **Location:** `docs/multi-subtitles/plan-ms4.md:272–281`; `frontend/node_modules/hls.js/src/controller/buffer-operation-queue.ts:115–139`; `src/controller/buffer-controller.ts:2043–2062,2159–2168`.

   The held native call correctly delays queue completion. However, `currentTime >= start − 0.25` also accepts three samples **after** the failed interval. Likewise, “while playhead is in this interval” needs an explicit nonzero observation count to prevent vacuous recovery evidence.

   **Required plan change:** bound failure samples by the interval’s end, preserve an explicit stalled-at-start allowance, and require at least one post-recovery sample inside the identified interval.

5. **NONBLOCKING — Round-7 finding 4 remains OPEN: the SourceBuffer explanation is too broad.**

   **Location:** `docs/multi-subtitles/plan-ms4.md:99–103`; `frontend/node_modules/hls.js/src/remux/passthrough-remuxer.ts:163–169`; `backend/internal/media/adapters/ffmpeg/transcoder.go:234–244`.

   `SourceBuffer.buffered` also applies EOS extension across tracks managed by that buffer. Here, the dedicated buffer contains one audio track, so that extension does not introduce the video-tail counterexample. The chosen mechanism remains sound. [MSE SourceBuffer attributes](https://w3c.github.io/media-source/#dom-sourcebuffer-buffered).

   **Required plan change:** qualify the explanation with this single-audio-track condition.

6. **NONBLOCKING — Round-7 finding 5 remains OPEN: event descriptions still overstate success.**

   **Location:** `docs/multi-subtitles/plan-ms4.md:84–95`; `frontend/node_modules/hls.js/src/controller/buffer-controller.ts:990–1007,2043–2084`; `src/controller/buffer-operation-queue.ts:84–89`.

   Quota recovery requeues an operation without another `BUFFER_APPENDING`. An asynchronous SourceBuffer error is followed by `updateend`, whose completion handler still emits `BUFFER_APPENDED`. Thus APPENDED alone is not proof of success.

   **Required plan change:** describe logical attempts versus physical executions, retain error status through APPENDED→FRAG_BUFFERED, and explicitly test nested dispatch where the internal BufferController runs before application listeners.

7. **NONBLOCKING — Round-7 findings 6–11 remain CLOSED.**

   **Location:** `docs/multi-subtitles/plan-ms4.md:73–212,218–247,284–297`.

   The previous counterexamples remain addressed:

   - Initial/default/remembered selection can reach SWITCHED; cancellation need not. M1 correctly avoids trusting it (`audio-track-controller.ts:182–199,254–257`; `audio-stream-controller.ts:792–807,1114–1118`).
   - Ready eligibility uses `videoState`, not component presence. Production publication/deletion preserves a Ready audio track (`learn.$.tsx:82–90,276–307`; `LessonPlayer.tsx:370`; backend `publish.go:70–91`; `audiotrack.go:218–227`).
   - Current-track refs, language dependencies, case-insensitive selection/menu identity, native invalidation, and transcript error handling cover their earlier gaps (`plan-ms4.md:122–126,178–201`; migration `00030:88–89`).
   - PATCH requires both language and label; fixture generation/export and T0 cover the named sources. The new spec is CI-selected while ASR is conditionally excluded (`handler.go:129,367–370`; `media.setup.ts:76–79,143–161`; `playwright.config.ts:49–60`; `.github/workflows/ci.yml:114–115`).

   **Required plan change:** none for these closed findings. One blocking finding remains: E13 fingerprint synchronization.

VERDICT: BLOCKED