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

ROUND 2. Your round-1 critique is `docs/multi-subtitles/_wip-ms4-plan-critique-1.md` (VERDICT: BLOCKED, 7 blockers, 3 nonblocking). The plan was rewritten as r2; each fix is tagged `[r1-N]`. A new owner decision O7 (plan-ms4.md §0, also added to plan.md §1) adds a backend field so the panel shows the predicted track's transcript before Play. For every round-1 finding state CLOSED or OPEN with reason, then critique the new material (O7/M3 backend field, M5 generation binding, M9 e2e feasibility: holding audio fragments via page.route, the `data-caption-selection` attribute) with the same rigor. Same output format and final VERDICT line.
