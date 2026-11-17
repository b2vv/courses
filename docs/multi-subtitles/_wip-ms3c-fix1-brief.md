# MS3c fix run 1 brief for Codex (review w1: 3 Major, 1 Minor) — e2e only

Repository: current directory, branch `ms2-3-asr-per-track` @ `7b20d2c` + uncommitted MS3c delta (package `ms3c-w1`, hashes in the
MS3c record block at the end of `docs/multi-subtitles/report.md`). Shared working tree: do NOT run git checkout/reset/stash/clean/rm,
do not commit, do not touch generated files, backend or `frontend/src` (a proven product defect → STOP that test + QUESTION).

WORKING FILE: docs/multi-subtitles/_wip-ms3c-fix1.md
Append AFTER EVERY STEP (append, not rewrite). First entry before the first tool: a `NEXT:` line. Then FACT / DECISION / NEXT / QUESTION.
Final answer = concise summary of the file.

Review: `docs/multi-subtitles/_wip-ms3c-review-codex-w1.md`. Runtime/cleanup as in your runs 1–2 (`courses-kiri` prefix,
`--project=chromium -g`, own `--output=test-results/ms3c-fix1-*`).

1. **Major 1 (R8, R11b–d): expected values independent of the PUT.** R8: each tab's PUT segment texts equal that track's ASR draft
   segments (read from `GET drafts`, per trackId) — und→en PUT = the und draft, en-gb PUT = the en-gb draft; then GET equals those.
   R11b/c/d: the PUT texts equal the text the test typed/restored (constants in the test), then GET equals them. For each, record a
   mutation check in the journal: temporarily change the expectation (e.g. swap the two drafts' texts in R8, or expect the pre-edit
   text in R11b) → the assertion fails; revert. Copy both reporter lines.
2. **Major 2 (R12b): discriminate placement.** Replacement video with a default track of one language AND a non-default track whose
   language matches the snapshot → the offer appears in the matching non-default tab and not in the default tab. Keep a second variant
   where no tab matches → default tab. Both continue: Restore → Save → 200 → new transcript has the restored lines, the original
   language transcript unchanged. Capture the old/new asset identity from the media read (upload/inventory), not from an optional draft.
   (Mint a non-`und` uploaded track the same way as run 1's en-GB mint, or via the S5 list if its language is offered.)
3. **Major 3 (R15 `en`): remove the race.** Wait for the initial recognition to finish, rename the track to `en`, Regenerate explicitly,
   wait for a draft with `trackLanguage: 'en'`, then toggle the switch.
4. **Minor (cleanup):** per-course try/catch in `afterAll`, record failures, continue; fail at the end if any failed.

Then run the whole `e2e/asr.spec.ts` once (no `-g`) and copy the summary + every reporter line.

## Done = report
`cd frontend && npm run format && npm run typecheck && npm run lint` exit codes; per fixed test reporter line; mutation checks; the
full-file summary; changed files + SHA-256 of `frontend/e2e/asr.spec.ts`.
