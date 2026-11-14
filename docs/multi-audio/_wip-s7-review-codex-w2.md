Reviewed `e86b733..9c1c641` on `multi-audio-s7`, including the wave-1 record. Read-only inspection; no tests or mutations executed. Reported test results and CSS-removal evidence remain author-supplied.

**Ranked findings**

- **Minor — the third track weakens the first-playlist proof.** [audio-tracks.spec.ts:355](/Users/strelia/projects/courses/frontend/e2e/audio-tracks.spec.ts:355) accepts any uploaded-track playlist. German and English both match that regex and contain the same 880 Hz audio ([setup:265](/Users/strelia/projects/courses/frontend/e2e/audio-tracks.spec.ts:265)). **Counter-example traced:** startup requests German first, then switches to English before the final radio assertion. Playlist, tone and final selection assertions all pass. Capture the English track ID and assert its exact playlist URI.

- **Minor — first-audible latency remains wait-relative.** [audio-tracks.spec.ts:334](/Users/strelia/projects/courses/frontend/e2e/audio-tracks.spec.ts:334) accurately logs completion of five samples, but `switched.firstAt` starts inside `waitForTone`, after clicking and checking the radio. **Counter-example traced:** 400 ms spent completing those actions plus a first tone 1,200 ms into the wait means 1,600 ms from the click. The log does not directly provide the plan’s click-to-first-tone measurement. Record the first sample timestamp against `clickedAt`.

No Blocker or Major finding.

**Wave-1 dispositions**

| Finding | Status | Opened and counter-example traced |
|---|---|---|
| **Major: stall followed by recovery passes** | **Resolved** | [Recorder and assertion:312–331](/Users/strelia/projects/courses/frontend/e2e/audio-tracks.spec.ts:312), `play`, `waitForTone`, and player media callbacks. A `waiting` event followed by recovery now leaves a nonempty array and fails. |
| **Minor: Tab reachability / visible focus** | **Resolved** | [Keyboard test:373](/Users/strelia/projects/courses/frontend/e2e/audio-tracks.spec.ts:373), `PlayerSettingsMenu`, RadioGroup and focus CSS. Removing the gear or selected radio from sequential navigation exhausts the 20-step loop and fails the explicit focus assertion. Removing the radio outline fails the computed-style assertion. |
| **Minor: deduplication creates another duplicate** | **Resolved** | [Label construction:121](/Users/strelia/projects/courses/frontend/src/modules/playback/ui/LessonPlayer.tsx:121), [regression test:905](/Users/strelia/projects/courses/frontend/src/modules/playback/ui/LessonPlayer.test.tsx:905), backend validation and uniqueness index. The original `Dub` / `Dub` / `Dub (uk)` counter-example now produces distinct final labels. |
| **Minor: latency / deadline** | **Partially resolved; latency detail open** | [Response listener:287](/Users/strelia/projects/courses/frontend/e2e/audio-tracks.spec.ts:287), [remaining budget:321](/Users/strelia/projects/courses/frontend/e2e/audio-tracks.spec.ts:321), and helper. Response timestamps are captured on arrival; the old sequential extra ten-second windows are gone. First-audible logging retains the Minor above. |

**Additional checks**

- **Dedupe uniqueness:** sound for backend-valid exact language identities. [audiotrack.go:267](/Users/strelia/projects/courses/backend/internal/media/domain/audiotrack.go:267) permits only `[A-Za-z0-9-]`; therefore a language contains no parentheses. Each appended final `(language)` suffix is unambiguous. Equal resulting strings require equal languages, which the per-asset unique index forbids among Ready tracks. I also traced labels containing nested suffix-shaped text.

- **CSS scope and effects:** [main.css:2521](/Users/strelia/projects/courses/frontend/src/app/styles/main.css:2521) applies only beneath `.lesson-player-settings-panel`. It intentionally affects speed/quality/audio radios and captions/autoplay switches. It permits wrapping and constrains width; handlers, fixed decorative marks, minimum control height and focus rules remain intact. Other forms are outside its scope. The reported removal failure supports discrimination; I did not reproduce it.

- **Waiting recorder:** installed before the click on the same video element, with no reset during measurement. Initial preload/resume seeking precedes three baseline tone samples. No explicit seek occurs inside the measured interval. A delayed startup event could conservatively fail the test; events before installation or after assertion remain outside its claim. This proves absence of recorded `waiting` during that interval, not continuous acoustic gaplessness.

- **R1 journey:** [lines 339–361](/Users/strelia/projects/courses/frontend/e2e/audio-tracks.spec.ts:339) now visit the single-track lesson, reach playback, read persisted `en`, then return and check audio plus selection. An automatic preference overwrite to the default is rejected. Exact first-English-playlist identity remains weakened as described above.

- **German fixture / cleanup / compatibility:** both relevant radio counts now expect three. Sequential upload and creation-order track sorting retain Original → English → German, so ArrowDown still expects English. Opened S5 media and S6 ASR scenarios: their counts concern their own courses/assets. Course registration before later failures and [afterAll cleanup:227](/Users/strelia/projects/courses/frontend/e2e/audio-tracks.spec.ts:227) remain intact. No further delta regression found.

This is approval of the delta with two nonblocking Minors. The required full ordered gate remains outstanding; targeted results do not establish release readiness.

**VERDICT: SHIP**