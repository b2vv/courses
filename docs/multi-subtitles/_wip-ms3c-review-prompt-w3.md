# MS3c review w3 — independent read-only review of the delta after the gate (package `ms3c-w3`)

Repository `/Users/strelia/projects/courses`, branch `ms2-3-asr-per-track`, HEAD `7b20d2c` + uncommitted MS3c. Read-only: do not
modify files, do not run tests or touch Docker. You reviewed package `ms3c-w2` (verdict `_wip-ms3c-review-codex-w2.md`, SHIP).

Package `ms3c-w3` (SHA-256 of current bytes — verify them before and after your review):
- `frontend/e2e/asr.spec.ts` `f74e41d53c60e706e7b9ebaabc617f1fc65b6de78a4192c79a7f37ee94798a13`
- `frontend/src/modules/media/ui/LessonAudioTracksField.tsx` and `.test.tsx` — unchanged from `ms3c-w2` (check: `24e71a1e…`, `1c6698fc…`).
- w2-reviewed bytes of `asr.spec.ts` (`a3682a92…`): `docs/multi-subtitles/_wip-ms3c-w2-asr.spec.ts.txt`; delta w2 → w3:
  `docs/multi-subtitles/_wip-ms3c-w3.diff` (one oxfmt re-wrap + the `renameTrack` wait).

Raw evidence: `docs/multi-subtitles/report.md`, MS3c section, entries "Attempt 4" onwards (ordered `make check` exit 2: M3-R11 step (c)
Save refused with `track_changed`, request order recorded; instructor W07/W08 503). Product guard: `frontend/src/modules/admin/ui/LessonLanguageTabs.tsx`
`save` (~273–350), tab ids `lessonLanguageTabs` (~34–62); React Aria tab DOM id = `${baseId}-tab-${key}`
(`frontend/node_modules/react-aria/dist/private/tabs/utils.mjs`). Contract: `docs/multi-subtitles/plan-ms2-3.md` §4 rows M3-R10, M3-R11.

Questions, Standards and Spec axes separately:
1. Is the diagnosis right — is the R11 failure a test race (Save before the tabs re-render) and is the `track_changed` refusal the
   behaviour M3-R10 requires? Or is it a product defect a fast author hits (then say what the product should do)?
2. Does the new wait close the race for every `renameTrack`/`renameEmbedded` call site (lines ~492, 657, 1016–1052, 1147–1155, 1201,
   1261, 1326), including renaming a non-selected track and renaming into a language that already has an orphan `lang:` tab?
3. Does it weaken any row (e.g. hide a real M3-R10 refusal)?
4. The W07/W08 503 is outside this diff; say only whether anything in the package could cause it.

Each finding: file:line, consequence, how to check, blocking or not. Verdict SHIP or BLOCK for `ms3c-w3`. Name what you did not check.
