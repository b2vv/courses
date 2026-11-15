**VERDICT: BLOCKED**

Reviewed HEAD `6ad8d8d35700795a3c3e212569fb7d584ad13def`; untracked plan SHA-256 `7cfe32aad2a9b71223b1701944d378a2d218ec0ac3751175631765708397e4c8`. Four independent lenses launched / four returned. Read-only review; no browser runs, tests or file changes.

The revised spacing calculation is correct for the documented CSS simulation:

- Admin `--card-padding` overrides the base token to **1rem**, so it becomes 32px at 200%.
- Current chain: `64 + 66 + 36 + 33 + 66 + 50 = 315px`.
- Proposed chain: `64 + 8×9.75 + 11 = 153px`, leaving **237px** for the measured **233px** trigger.
- Track rows and the new-track fieldset are parallel paths; counting both together would be wrong.
- Proposed dialog content width: `390 − 2×15.6 − 2×19.5 − 2 = 317.8px`.

These calculations establish the measured trigger’s budget, not every possible lesson descendant’s fit.

**Round 1 dispositions**

| Finding | Disposition |
|---|---|
| **1 — Dialog outside the corrected chain** | **Resolved in the plan.** F-D2 supplies an independent dialog correction; F-R2 checks internal overflow, selected-option visibility and focus return. |
| **2 — 320px combined with 200% text** | **Rebuttal accepted for this scoped fix.** The inherited acceptance matrix separates 320px reflow from 390px/200% text. The combined case still fails arithmetically—181px available—but is now explicitly excluded. The blanket WCAG explanation should be softened: W3C discusses Resize Text and Reflow operating together, so separate test columns alone do not establish general conformance. [W3C explanation](https://www.w3.org/WAI/WCAG22/Understanding/reflow.html#resize-text) |
| **3 — Trigger is not the largest possible child** | **Open, partially addressed.** F-D3/F-R6 handle long audio labels. Populated transcript document fields remain unaddressed; see Major below. |
| **4 — Simulation does not establish real text zoom** | **Rebuttal accepted as a restricted evidence claim.** Lines 67–76 distinguish font settings, page zoom and Firefox text-only zoom, and explicitly record the latter as NOT_TESTED. This resolves the overclaim; it does not supply browser evidence. |
| **5 — Desktop comparison unspecified** | **Resolved in substance.** F-R4 pins revision, fixture, locale, viewport, 100% text, paddings and bounding boxes. Record font readiness and scrollbar conditions when executing the comparison. |
| **6 — Cascade and RTL** | **Resolved in the plan.** Both module rules are included, physical shorthands become logical properties, and absence of an RTL product run is explicit. |
| **7 — Token range and compatibility wording** | **Resolved.** The range is correctly 12–24px; container-query compatibility is distinguished from the F5 `:has()` problem. |

Real text-only zoom can also enlarge rem-based layout dimensions; Mozilla documents that mechanism. At a fixed CSS viewport, the viewport cap addresses that dependency, but the Chromium root-font experiment does not prove identical browser behavior. Page zoom changes the CSS viewport and therefore the `vw` budget. The revised limitation is appropriate. [Mozilla implementation analysis](https://bugzilla.mozilla.org/show_bug.cgi?id=1864445#c3)

**Ranked findings**

1. **Major — Whole-lesson acceptance still omits a known larger intrinsic-width child.**  
   Where: [plan-f3.md:89](/Users/strelia/projects/courses/docs/multi-subtitles/plan-f3.md:89), [plan-f3.md:95](/Users/strelia/projects/courses/docs/multi-subtitles/plan-f3.md:95).  
   F-R1 promises that every lesson descendant fits, but the fixture specifies only an `und` audio track. [TranscriptEditor.tsx:262](/Users/strelia/projects/courses/frontend/src/modules/transcript/ui/TranscriptEditor.tsx:262) renders another plain fieldset, retaining UA min-content sizing, with document labels rendered directly at line 277. Valid labels permit **200 unbroken characters** ([transcript-schema.ts:42](/Users/strelia/projects/courses/frontend/src/modules/transcript/adapters/transcript-schema.ts:42)). Neither that fieldset nor its labels receive the proposed audio-specific correction.  
   **Required plan edit:** add populated transcript/document fixtures at 390/200% and 320/100%, and specify how their fieldset and label fit—or explicitly narrow the whole-lesson claim and record this unresolved case. Include representative ASR states in the fixture matrix rather than inferring their coverage.

2. **Minor — F-R7 permits the old, incorrect dialog padding.**  
   Where: [plan-f3.md:101](/Users/strelia/projects/courses/docs/multi-subtitles/plan-f3.md:101).  
   “As baseline **or** ≤ …vw” accepts unchanged 24px dialog padding at 390/100%, although F-D2 specifies 19.5px. Compare each primitive against its exact `min(token, cap)` result. Since these declarations are global, include a client dialog in light/dark as well as the admin examples.

3. **Minor — Nested-overflow acceptance needs a precise selection rule.**  
   Where: [plan-f3.md:95](/Users/strelia/projects/courses/docs/multi-subtitles/plan-f3.md:95).  
   A right-edge comparison alone misses left-side overflow and content clipped inside a correctly sized child. Specify rendered, in-flow descendants; check both inline boundaries and relevant containers’ `scrollWidth/clientWidth`. Inspect portalled popovers separately through F-R2/F-R5.

No unnecessary-complexity finding: existing selectors and native `min()` remain a small solution. Module/lesson titles are input values with shrinkable inputs; legends contain ordinal labels. The proposed padding changes preserve control minimum sizes. F4’s scrolling correction and F5’s compatible selectors remain intact.

**One open blocking finding: the unaddressed portion of r1 Major 3.**