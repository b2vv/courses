Reviewed `71e17bf..3382e27` read-only. No tests or browser interactions rerun.

**F5 — CLOSED.** Opened the wave-5 record, CSS delta, and all four popover implementations.

- [tokens.css:262](/Users/strelia/projects/courses/frontend/src/app/styles/tokens.css:262) and `.ui-popover-list` at line 280 each have specificity `(0,1,0)`. The latter appears later, so `padding:0` wins. The child rule at line 285 puts padding and scrolling on the list.
- Select and ComboBox both opt in at [controls.tsx:183](/Users/strelia/projects/courses/frontend/src/presentation/ui/controls.tsx:183) and line 237. Menu opts in at [overlays.tsx:74](/Users/strelia/projects/courses/frontend/src/presentation/ui/overlays.tsx:74).
- A frontend source search found no other Popover containing a ListBox/Menu without the class. The remaining popover, [PlayerSettingsMenu.tsx:60](/Users/strelia/projects/courses/frontend/src/modules/playback/ui/PlayerSettingsMenu.tsx:60), contains a Dialog and correctly retains the base padding.
- Counter-examples checked statically: empty ComboBox, Menu carrying `.ui-listbox`, constrained-height lists, and the player Dialog. No regression found. Live Firefox 104–120 verification remains explicitly unperformed; the unsupported selector mechanism has been removed.

**F6 — CLOSED.** Opened [report.md:173](/Users/strelia/projects/courses/docs/multi-subtitles/report.md:173) and all five `final-*` screenshots:

- `final-320-reflow-U4-new-track-keyboard.jpg`
- `final-320-reflow-U4-dialog-open-selected.jpg`
- `final-390-text200-U4-dialog-open-selected-visible.jpg`
- `final-1440-combobox.jpg`
- `final-1440-newtrack-end.jpg`

The images show visible focused options, including initially selected Undetermined in the dialog. The revision-pinned record supplies the remaining interaction evidence: 320×640 navigation, Escape/focus return, ComboBox scrolling, and the four baseline widths. Screenshots alone cannot prove those keyboard sequences; closure rests on both the images and the recorded recheck.

**Tests and selectors.** No unit/e2e assertions referencing `.ui-popover`, `.ui-popover-list`, `.ui-listbox`, or the removed `:has()` selector were found. Existing tests use semantic roles: Select selection/focus return at [ui.test.tsx:192](/Users/strelia/projects/courses/frontend/src/presentation/ui/ui.test.tsx:192), ComboBox filtering at line 305, and Menu actions at line 382. Thus no stale selector assertions need updating; the scrolling/layout proof is the recorded browser verification.

**Gate — acceptable as recorded.** [report.md:43](/Users/strelia/projects/courses/docs/multi-subtitles/report.md:43) preserves the first run as red and separately records a complete ordered rerun: `GATE_EXIT=0`, frontend 948, e2e 138. AGENTS.md does not forbid rerunning a failed gate. This record does not turn skipped downstream checks from the first attempt into passes.

The failure location supports an infrastructure explanation:

- [sessions_integration_test.go:74](/Users/strelia/projects/courses/backend/internal/playback/application/sessions_integration_test.go:74) resolves the testcontainer connection string; line 76 fails on that error.
- Database opening follows at line 78; migration execution follows at line 88. Migration 00030 cannot directly cause this pre-connection port lookup failure.
- `media-masters` is a Compose service; this integration test starts its own containers.
- [audiotrack.go:299](/Users/strelia/projects/courses/backend/internal/media/adapters/postgres/audiotrack.go:299) implements `ListReadyAssetIDs` through an established database pool and is not called by this setup path.

The exact Docker/testcontainers root cause is unproven; the Ryuk comparison is an analogy, not a diagnosis. That qualification does not invalidate the separately green full rerun.

Ranked findings:

- **Blocker:** None.
- **Major:** None remaining; F5 and F6 closed.
- **Minor:** Existing accepted F3 remains at [report.md:137](/Users/strelia/projects/courses/docs/multi-subtitles/report.md:137): narrow nested-field wrapping at 200% text. No new Minor finding.

**VERDICT: SHIP**