**VERDICT: ACCEPTED**

Round 3, stage: plan. Reviewed HEAD `6ad8d8d35700795a3c3e212569fb7d584ad13def`; untracked `docs/multi-subtitles/plan-f3.md`, SHA-256 `a03d2e35143c293e741a40c29d23c7d2d29ecdda62ab10f7b76a80ca7ce96117`.

Read-only review across the four requested lenses. No files changed, browser runs or tests performed. This accepts the plan, not implementation results.

**Round 1 dispositions**

| Finding | Disposition and reasoning |
|---|---|
| Major 1 — rename dialog outside the corrected chain | **Resolved.** F-D2 supplies its own spacing correction; F-R2 measures internal overflow, selected-option visibility and keyboard focus return. |
| Major 2 — 320px combined with 200% text | **Rebuttal accepted for this scoped fix.** Lines 86–94 explicitly exclude the combined case and require recording its limitation. The available width remains only **181px**. This acceptance does not establish general WCAG conformance; W3C discusses text enlargement and reflow together. [W3C explanation](https://www.w3.org/WAI/WCAG22/Understanding/reflow.html#resize-text) |
| Major 3 — larger intrinsic-width children | **Resolved in the plan.** F-D3/F-R6 address 64-character audio labels. F-D4/F-R9 now address the transcript fieldset and 200-character document labels that remained open in round 2. Module/lesson titles occupy shrinkable inputs; their legends contain ordinal labels. ASR coverage is explicitly limited to named rendered states. |
| Major 4 — simulation versus real text zoom | **Rebuttal accepted as a restricted evidence claim.** Lines 75–84 distinguish root-font simulation, browser font settings and page zoom, and mark Firefox text-only zoom **NOT_TESTED**. Real text-only zoom can enlarge `rem` dimensions, but the Chromium simulation does not establish identical browser behavior. [Mozilla analysis](https://bugzilla.mozilla.org/show_bug.cgi?id=1864445#c3) |
| Minor 5 — unspecified desktop baseline | **Resolved.** F-R4 pins `6ad8d8d`, identical fixture/locale/viewport, 100% text, computed paddings and bounding boxes. Font readiness and scrollbar conditions are now specified. |
| Minor 6 — cascade and RTL | **Resolved.** Both module declarations are included; physical shorthands become logical properties. The absence of an RTL locale/run is explicit. |
| Minor 7 — token range and compatibility wording | **Resolved.** The range is 12–24px; container-query compatibility is distinguished from F5’s `:has()` issue. Installed Rsbuild defaults include Firefox ≥104. |

**Round 2 findings**

- **Major 1, transcript fieldset/labels: resolved** by F-D4 and the populated F-R9 fixture.
- **Minor, F-R7 exactness: resolved in substance** by checking the actual `min(token, cap)` result across admin/client dialogs and both themes.
- **Minor, F-R1 selection rule: resolved** by selecting rendered in-flow descendants, checking both boundaries and internal overflow, and treating portals separately.

The sizing mechanism checks out against the source:

| Case | Calculation | Result |
|---|---|---|
| Existing 390/200% | `64 + 66 + 36 + 33 + 66 + 50` | **315px overhead** |
| Proposed 390/200% | `390 − 64 − (8 × 9.75 + 11)` | **237px available**, versus measured 233px trigger |
| Proposed rename dialog | `390 − 2×15.6 − 2×19.5 − 2` | **317.8px available** |
| Proposed 320/200%, excluded | `320 − 64 − (8×8 + 11)` | **181px available** |

The admin override is `--card-padding: var(--space-4)`—**16px**, doubling to 32px in the simulation—not the base 24px token. Track rows and the new-track fieldset are parallel paths; eight capped padding contributions is correct.

At 1024/1440 and 100%, the caps preserve existing padding. At 768, only module-start padding changes directly, from 24px to 19.2px. F-R4 correctly requires geometry comparisons as well as padding comparisons.

No AI-slop or unnecessary-complexity finding: existing selectors, logical properties and native `min()` are proportionate. F4’s listbox scrolling and F5’s compatible selectors remain untouched. The proposed changes preserve existing control minimum sizes.

**Remaining ranked findings**

- **Blocker:** none.
- **Major:** none.
- **Minor — allow numeric serialization tolerance in F-R7.** [plan-f3.md:114](/Users/strelia/projects/courses/docs/multi-subtitles/plan-f3.md:114): “exactly” should mean the specified formula, with a small numeric tolerance when comparing browser measurements. Layout quantization can serialize 15.6px slightly differently. Retain the exact expected formula; use, for example, ±0.1px for numeric comparisons.

**Open blocking findings: 0.** F-R1–F-R9 provide a checkable plan for the stated scope; real Firefox text-only zoom and combined 320px/200% remain explicit limitations.