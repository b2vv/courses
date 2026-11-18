WORKING FILE: docs/multi-subtitles/_wip-ms4-dp.md
Append AFTER EVERY STEP (append, not rewrite). The first entry comes before the first tool:
a `NEXT:` line stating what you intend to check. Then use the labels FACT / DECISION / NEXT / QUESTION.
Final answer = a concise summary of the file.

# MS4 design-pass evidence (plan-ms4 row M26) on `ms4-player-captions` @ af483ce6

Repository `/Users/strelia/projects/courses`. Only writer in this tree. No commits/staging/branch switch, never
git checkout/reset/stash/clean/rm. Do NOT change product code: a UI defect you find is a FINDING (state, viewport,
screenshot, measurement) for the owner session; record it and continue.

Model: `docs/multi-subtitles/ms3-design-pass/harness.spec.ts.txt` + `harness.config.ts.txt` (a temporary Playwright
spec run outside the e2e suite; its measure() helper, viewports list and evidence layout). Build the same for MS4
under `docs/multi-subtitles/ms4-design-pass/`: run it from a temporary location, then copy the harness as
`harness.spec.ts.txt` / `harness.config.ts.txt`, the screenshots (PNG) and `measurements.json` there. Never commit
raw Playwright traces (they carry session cookies and media bytes). Delete every course you create (reuse
`frontend/e2e/subtitles-cleanup.ts`), and record the leftover count (must be 0).

Fixture: reuse `frontend/e2e/subtitles.spec.ts` setup helpers' approach (student storage state from
`frontend/e2e/auth.ts`; never print credentials): one Ready lesson with `en` (embedded, PATCH) + `de` + `fr` tracks
and `en`/`de` transcripts. Stack: live Kiri Compose stack already rebuilt from this revision — prefix commands with
`/Users/strelia/project-data/courses-kiri/courses-kiri`.

States (each a screenshot + measurement per viewport):
- S1 panel `transcript` (en) after Play, captions on;
- S2 panel `loading` (switch to de while `page.route` holds `?language=de`);
- S3 panel `none-for-track` (fr) AND the gear menu open showing the text row «Субтитри: немає для цієї доріжки»;
- S4 panel `error` (`?language=de` answered 503) with «Спробувати ще»;
- S5 gear menu open in the normal state (captions switch) for comparison.
Viewports: 390×844, 768×1024, 1024×768, 1440×900 at 100 %; 1440 and 390 at 200 % text (html font-size); 320×640.
Measure per state/viewport: horizontal overflow (`scrollWidth` vs viewport; elements of the player/panel past the
right edge), the size of the retry button and gear menu items (≥ 44 px), text clipping (`scrollWidth > clientWidth`
on the new texts), the live-region role present on the panel state (alert for error, status otherwise), and in
dark mode (`emulateMedia({colorScheme:'dark'})`) contrast of the new texts against their actual background
(compute from computed colours; if the background is not a solid colour, record NOT_TESTED, do not invent a ratio).
Keyboard: from the player, Tab to the gear button, Enter opens the menu, arrow/Tab through items — the «немає» row is
not focusable/actionable and is announced as text; Escape closes and focus returns to the gear button; Tab reaches
«Спробувати ще» in S4 with a visible focus outline (record outline width/colour); Enter on it re-requests
(`?language=de` seen in the request log).

Report: a table state × check with PASS / FAIL / NOT_TESTED and the evidence path, in the working file. Do not mark
anything PASS you did not measure.
