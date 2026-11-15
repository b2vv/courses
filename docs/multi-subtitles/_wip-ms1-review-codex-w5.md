Reviewed `30c4031..71e17bf`, read-only. Opened the CSS, all four popover implementations, React Aria’s installed implementation, the before-fix F4 screenshot, and all nine new screenshots. No tests or browser interactions rerun; the reported ordered gate on `c34c0d0` is recorded evidence: exit 0, e2e 138, frontend 948.

**F4 disposition:** fixed in the supplied browser evidence. `390-text200-U4-dialog-open-selected-visible.jpg` shows the selected Undetermined option fully visible with its focus ring. `…-dialog-home.jpg` shows Ukrainian fully visible after Home; `390-text200-U1-end-after-F4.jpg` shows Spanish after End. Closure remains incomplete for the configured Firefox targets below.

**CSS assessment**

| Popover | Matches `:has(> .ui-listbox)`? |
|---|---|
| Select, [controls.tsx:183](/Users/strelia/projects/courses/frontend/src/presentation/ui/controls.tsx:183) | Yes |
| ComboBox, [controls.tsx:237](/Users/strelia/projects/courses/frontend/src/presentation/ui/controls.tsx:237) | Yes, including empty results |
| Menu, [overlays.tsx:74](/Users/strelia/projects/courses/frontend/src/presentation/ui/overlays.tsx:74) | Yes: its `Aria.Menu` carries `.ui-listbox`. Currently unused in product UI. |
| Player settings, [PlayerSettingsMenu.tsx:60](/Users/strelia/projects/courses/frontend/src/modules/playback/ui/PlayerSettingsMenu.tsx:60) | No: its direct content is a Dialog. |

Counter-examples examined statically: long lists under an inline popover `max-height`, first/last focused options, empty/loading ComboBox content, Menu, and the player Dialog.

In browsers supporting the selector, I found no additional defect:

- The list can shrink under the popover’s height limit because of `min-block-size:0`; excess content scrolls internally.
- Padding remains inside the list. Its 8px inset accommodates the option outline’s 3px width plus 3px offset.
- Positioning and flip styles remain controlled by React Aria; the change does not override them.
- Empty ComboBox content remains inside the padded ListBox; React Aria’s empty-state wrapper uses `display:contents`. This was inspected statically, without a new empty-state screenshot.

**Ranked findings**

- **Blocker:** None.

- **Major — F5: F4 remains unfixed for configured Firefox 104–120.** [tokens.css:279](/Users/strelia/projects/courses/frontend/src/app/styles/tokens.css:279). There is no project browserslist override; installed Rsbuild defaults include Firefox ≥104, matching its [documented defaults](https://v2.rsbuild.rs/guide/advanced/browserslist). Firefox supports `:has()` starting at 121. [MDN compatibility data](https://raw.githubusercontent.com/mdn/browser-compat-data/main/css/selectors/has.json)

  On those targets, the parent loses neither padding nor its block layout, while the unconditional child rule at line 284 adds padding. The list consequently lacks the flex constraint that makes this fix work, retaining the original popover-scrolling mechanism. This is a static compatibility finding, not a live Firefox reproduction.

  Use an explicit class on the affected popovers, or another compatible selector. Recheck F4 and shared controls, then repeat the required full gate.

- **Major — F6: final-revision reflow verification is missing.** [report.md:155](/Users/strelia/projects/courses/docs/multi-subtitles/report.md:155). The post-fix sweep names four baseline widths; the 320px evidence predates this shared scrolling change. [plan-ms1.md:146](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms1.md:146) requires reflow, and design-pass requires fresh evidence for affected scenarios after the final fix. Record a post-fix 320px check covering initial opening, keyboard navigation, visible focus, Escape and focus return.

- **Minor:** Existing F3 remains the previously accepted follow-up. No new Minor finding.

**VERDICT: BLOCK**