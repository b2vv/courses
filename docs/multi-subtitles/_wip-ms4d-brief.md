WORKING FILE: docs/multi-subtitles/_wip-ms4d.md
Append AFTER EVERY STEP (append, not rewrite). The first entry comes before the first tool:
a `NEXT:` line stating what you intend to check. Then use the labels FACT / DECISION / NEXT / QUESTION.
Final answer = a concise summary of the file.

# MS4 run (d) — e2e `frontend/e2e/subtitles.spec.ts` (E1–E14) and the existing-e2e inventory (M9/M10)

Repository `/Users/strelia/projects/courses`, branch `ms4-player-captions` @ `092ff1b7` (runs a–c
committed). You are the only writer in this tree. Do not commit, stage, switch branches, or git
checkout/reset/stash/clean/rm. Product code under `frontend/src`/`backend` only if an e2e proves a defect —
then STOP that test, write a QUESTION with evidence (request log, screenshot path) and continue the others;
the owner session decides the fix.

Contract: `docs/multi-subtitles/plan-ms4.md` (ACCEPTED r9) §3 M9 and M10 and §4 rows M10–M25 — read them
in full; every E-row's recipe there is binding (fixtures, page.route barriers, the audio tap, the
appendBuffer wrapper with ISO-BMFF classification, FNV-1a mdat fingerprints registered before
`route.fulfill`, bounded observation windows, `DOMException(msg, 'InvalidStateError')`). Patterns to reuse:
`frontend/e2e/audio-tracks.spec.ts` (upload/Ready poll helpers, tapAudio/peak/waitForTone at :160-205, dub
response path regex :289-325), `frontend/e2e/transcript.spec.ts` (transcript PUT, enrollment, openStudent),
`frontend/e2e/auth.ts` (storage states; never print credentials). UI names: the e2e runs with
`locale: 'uk-UA'`; read the real strings in `frontend/src/app/providers/locales/uk.ts` and the components
(`LessonPlayer.tsx`, `PlayerSettingsMenu.tsx`, `TranscriptPanel.tsx`); `data-caption-selection` is on
`section.lesson-player`.

## Runtime
Live Compose stack on Docker context `kiri`, rebuilt from this branch by the owner session. Prefix every
docker/npx command with `/Users/strelia/project-data/courses-kiri/courses-kiri` (never change the global
Docker context). Targeted run:
`cd frontend && /Users/strelia/project-data/courses-kiri/courses-kiri npx playwright test e2e/subtitles.spec.ts --project=chromium --output=test-results/ms4d`
(the setup project runs first and regenerates fixtures — needed for the new 1320 Hz fixture; add it to
BOTH the ffmpeg recipe at `media.setup.ts:76-79` and the export list at `:143-153`). For single tests use
`-g '<title>'`; remember a single test with dependencies may need `--no-deps` plus a reseed (see existing
notes in the file headers). Host load is very high (load average ~60): use generous per-test timeouts like
the existing media tests; never assert tight upper bounds on durations — E7 logs measurements.

## Scope
1. New `e2e/subtitles.spec.ts` with E1–E14 exactly as the plan specifies, one `test` per row, titles
   prefixed with the row id (`E1 …`). Serial describe where rows share the prepared course.
2. M10 inventory: update `transcript.spec.ts:300-365` (PATCH embedded to `uk` with a label before opening;
   before Play the panel shows `uk`; Play; read the CURRENT track's mode/cues without assigning `mode`;
   turn captions on through the gear switch). Check each other row of the M10 table and record the
   disposition you verified (file:line) in the working file.
3. Run the new spec and the touched `transcript.spec.ts` targeted until green; then run each once more
   in full (`--project=chromium`, no `-g`) and record pass counts. Do NOT run the whole e2e suite or
   `make check` (the owner session runs the ordered gate). Copy E7/E4b measurement lines into the working file.

## Two channels
- Missing or contradictory contract (e.g. a recipe step that cannot be built on the real stack) → STOP
  that row, QUESTION with evidence.
- An inaccurate plan statement that would make a row vacuous or wrong → argue it (DECISION with evidence).
A row that cannot be made to pass is a finding, not something to weaken: never relax an assertion to green.
