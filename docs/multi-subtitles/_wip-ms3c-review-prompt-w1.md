You are an independent reviewer (read-only). Repository: current directory, branch `ms2-3-asr-per-track` @ `7b20d2c` + the
uncommitted MS3c delta: `frontend/e2e/asr.spec.ts`, `frontend/src/modules/media/ui/LessonAudioTracksField.tsx` (+ `.test.tsx`).
Do not modify files. Do not run git checkout/reset/stash/clean.

Contract: `docs/multi-subtitles/plan-ms2-3.md` §4 MS3 table (rows M3-R1..R15, e2e parts) and the MS3c section at the end of
`docs/multi-subtitles/report.md` (checkpoint decisions, defect 1). Author journals: `docs/multi-subtitles/_wip-ms3c-run{1,2}.md`.

Review questions:
1. For each row with an e2e part (R2, R4, R6, R8, R9, R10, R11 a–d, R12 a/b, R13 a/b, R14, R15): does the test actually prove the row
   through the production path, or does it prove a fake (stubbed response where the plan demands a real one, assertion that is true
   regardless, a step skipped, a selector that matches something else)? Name rows that are weaker than the plan, with file:line.
   Pay attention to timings that look too short for what they claim (e.g. M3-R12a "first processing" in ~5 s).
2. Flake risks: fixed sleeps, races between UI and background recognition, assertions on global state, leftovers (courses, IndexedDB,
   paused Redis queue not resumed in `finally`).
3. Defect 1 fix (`stopPropagation` in the track edit dialog submit): correct and complete? Any other portaled form/submit inside the
   course editor form that has the same class?
4. Anything else that would block merge.

Output: verdict SHIP or BLOCK, then findings ranked Blocker/Major/Minor with file:line, a concrete failure scenario and a fix.
