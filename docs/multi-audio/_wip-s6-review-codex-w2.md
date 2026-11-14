Reviewed `6aef653 → fc7421d` on `multi-audio-s6`, against the wave 1 record. Counter-examples below are source traces; tests, Whisper and ordered gates were **NOT RUN**. No files changed.

1. **Stale window survives cache eviction.** Opened `frontend/src/app/composition/admin-draft-queries.ts:42`, `:155` and `admin-draft-s6.test.ts:102`.

   Counter-example: observe stale for half the 15-minute window, unsubscribe, remove the cached query, then remount through undefined data. Result: undefined now preserves the original timestamp; polling stops at the original deadline. The added test exercises this path and would discriminate the previous reset behavior. An actual non-stale response still clears the timestamp.

2. **Stale-banner observation holds recognition.** Opened `frontend/e2e/asr.spec.ts:37`, `:53`, `:470`, `:479`.

   Counter-example: the ASR sweep enqueues replacement immediately after set-default, before the browser rereads the draft. Result: the pause is established before set-default; Asynq’s installed dequeue implementation checks that key atomically before taking work. Replacement cannot consume this observation window. A click or assertion failure enters `finally` and resumes recognition.

   **Other ASR tests:** `frontend/playwright.config.ts:4`, `:36`, `:49`, `:68` establish sequential execution with one worker and project dependencies. C06 uses the same pause/finally pattern (`asr.spec.ts:351`); **C18 does not pause** (`:389`). No overlapping pause ownership or normal failure-path leakage was found.

3. **Replacement content distinguishes the dub.** Opened `frontend/e2e/asr.spec.ts:174`, `:492`, `:496` and `frontend/e2e/media.setup.ts:74`, `:75`.

   Counter-example: replacement advances `updatedAt` but still recognizes embedded speech. Result: its text lacks both keywords, so the new positive assertion rejects it. The original’s negative assertion also checks that the discriminator separates the observed outputs.

   `/dub|languages/i` tolerates capitalization, punctuation, segment boundaries and “dubbed”; either keyword suffices. This is a reasonable fixture assertion, but actual Whisper robustness remains **unverified** here: a transcription such as “doubled … language” would fail despite correct audio selection. That possibility alone does not establish a defect.

The delta contains only these three files. No regression was found in wave 1’s verified items: backend claim/token/extraction behavior is unchanged; draft invalidation, manual pending markers, request scoping and terminal polling remain intact. The extended `readDraft` preserves its existing state/token/segment-count behavior.

Findings, ranked:

- **Blocker:** none.
- **Major:** none.
- **Minor:** none substantiated in this delta.

**VERDICT: SHIP**

Source-review verdict; runtime verification remains outstanding.