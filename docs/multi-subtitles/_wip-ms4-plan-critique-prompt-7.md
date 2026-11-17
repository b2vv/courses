You are an independent plan critic (stage: plan). Read-only: do not edit any file except that you return your critique as your final answer.

Repository: /Users/strelia/projects/courses (TanStack Start frontend in `frontend/`, Go backend in `backend/`). Code on `origin/main` = `f50f6ec`; the working tree matches it for `frontend/src`, `frontend/e2e` and `backend/`.

Plan under critique: `docs/multi-subtitles/plan-ms4.md` (slice MS4). Parent phase plan, ACCEPTED and binding: `docs/multi-subtitles/plan.md` — especially D1, D2, D7, §4 MS4 row, §6 limits, §7 open question "MS4: SWITCHING or SWITCHED". Owner decisions O1–O6 in §1 are fixed; do not reopen them.

Task: find what in plan-ms4.md would produce a bug, a gap against the parent plan, an unproven claim, or an acceptance row whose evidence would not take the production path. Verify every cited fact against the code (file:line) and against hls.js 1.7.2 sources in `frontend/node_modules/hls.js/src`. In particular check:
1. M1: does hls.js 1.7.2 fire AUDIO_TRACK_SWITCHED for the INITIAL selection (one track; remembered non-default selected synchronously in AUDIO_TRACKS_UPDATED; default), for a nextAudioTrack switch, and on a new loadSource / retry? Is "hide on SWITCHING, ask on SWITCHED" consistent with D7 ("before confirmation and while the request is in flight — no captions; another language never shown")?
2. M3/M4: is the boundary "player not rendered → read without language" well defined from existing facts in `frontend/src/routes/learn.$.tsx` and `LessonPlayer.tsx`? What happens for a lesson whose video is not ready, locked, email-gated, or a Ready asset with zero audio tracks (check backend media: can the last track of a Ready asset be removed?).
3. M5: does re-keying `<track>` work with the captionsOn effect and with existing e2e reading `video.textTracks[0]`?
4. M7: is case-insensitive matching safe given MS1's index, and does anything else (menu value `shownAudio`, `selectAudioTrack`, preference write) still compare exactly?
5. M8/M9: are the e2e rows feasible on the real stack with the named fixtures and existing PATCH API (find it), CI-selected (CI skips asr.spec), and is the T0 inventory mechanism complete (seeded lessons, playback.spec, accessibility.spec, learning.spec)?
6. Anything in D7 of the parent plan that plan-ms4 does not cover.

Output format: numbered findings, each `BLOCKER` or `NONBLOCKING`, with location (plan line and code file:line), the traced counterexample, and the required plan change. End with exactly one line: `VERDICT: ACCEPTED` or `VERDICT: BLOCKED`.

ROUND 7. Prior critiques: `_wip-ms4-plan-critique-{1..6}.md` (round 6 left 1 blocker, E13 evidence). The plan is now r7 with the E13 fix tagged `[r6-1]`: per-call outcome (threw/error/ok via SourceBuffer error and updateend), timestampOffset and parsed duration for the real interval, a held recovery until >=3 observations of the failed state, and language association by the 880 Hz tone. State CLOSED or OPEN for round-6 findings, verify the E13 recipe for determinism and vacuity (e.g. does delaying the real appendBuffer call break hls.js queue or trigger its append timeout), and report anything else that blocks. Same format and final VERDICT line.
