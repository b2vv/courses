WORKING FILE: docs/multi-subtitles/_wip-ms4-fix1.md
Append AFTER EVERY STEP (append, not rewrite). The first entry comes before the first tool:
a `NEXT:` line stating what you intend to check. Then use the labels FACT / DECISION / NEXT / QUESTION.
Final answer = a concise summary of the file.

# MS4 fix 1 — review findings (both axes BLOCK) on `ms4-player-captions` @ fc37aa2b

Repository `/Users/strelia/projects/courses`. Only writer in this tree. No commits, no staging, no branch
switch, never git checkout/reset/stash/clean/rm. Reviews: `docs/multi-subtitles/_wip-ms4-review-codex-spec.md`
and `_wip-ms4-review-codex-standards.md` — read both. Contract: `docs/multi-subtitles/plan-ms4.md` (M1, M5,
M8, E4b/E6 dated decisions). Fix exactly these, each with a test that fails before the fix:

1. **Replacement history across failed/overlapping attempts** (spec #1, `audio-coverage.ts:50-67`). When an
   append overlaps spans, its `replaced` set must include: languages of overlapping completed, appended,
   non-failed spans, PLUS the `replaced` history of overlapping spans that are failed or not completed; and the
   new attempt inherits the latest outstanding `settleUntil` of overlapping spans. Unit: buffered `de` [0,6);
   `en` append at playhead 2.5 fails; retry succeeds at 2.6 → pending until ≥ 2.5+1.5 (inherited) or the
   retry's own window, whichever is later; never confirmed `en` at 2.6.
2. **Portable e2e Docker call** (spec #2 / standards #1, `subtitles.spec.ts:16,89`). Use the neighbouring
   specs' convention (`execute('docker', ['compose', …])`, inheriting the environment: on Kiri the caller
   prefixes the whole Playwright run with the courses-kiri wrapper, which sets DOCKER_HOST/COMPOSE_FILE).
   No machine-specific path anywhere in `frontend/e2e`.
3. **Request the new language at SWITCHING** (spec #3, `caption-state.ts:43`, plan M1 "транскрипт нової мови
   запитується тоді ж"). Extend `CaptionSelection` pending with the selected language when one is selected
   (`{state:'pending'; requested?: string}`), set by the hls adapter from SWITCHING/initial selection and by the
   native adapter where known. The route then runs the `exact` query for `requested` immediately, but the panel
   stays `loading` and no `<track>` renders until `confirmed` with the same language (render gate unchanged;
   und family still never requests). Unit + route test: after SWITCHING to `de` the `?language=de` request is
   issued before confirmation, nothing `de` renders before confirmation.
4. **Live-region semantics of panel states** (standards #2, `TranscriptPanel.tsx:35`): `error` → `role="alert"`;
   `loading`, `none-for-track`, `none` → `role="status"` (one live element per panel at a time — check that no
   existing e2e counts `role=status` on the learn page unscoped; scope any such selector you find, AGENTS.md
   "Decisions" second bullet). Unit asserts the roles.
5. **No fixed sleep in E9** (standards #4, `subtitles.spec.ts:1061`): bounded `expect.poll` on advancing
   `currentTime`.
Also update the existing `panel loading` copy decision only if needed; do not touch anything else.

## Verify
`cd frontend && npm run format && npm run format:check && npm run typecheck && npm run lint && npx rstest run src --testTimeout 30000`.
Then (stack on Kiri is live and rebuilt by the owner session after you finish — do NOT rebuild images yourself):
stop here and report; the owner session rebuilds `frontend` and runs the browser checks. Report each command's
real result.

## Two channels
Missing/contradictory contract → STOP, QUESTION. A review finding you believe is wrong → argue it with evidence
(DECISION), do not silently skip it.
