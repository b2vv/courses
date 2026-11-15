**VERDICT: BLOCKED**

Reviewed HEAD `6ad8d8d35700795a3c3e212569fb7d584ad13def`, untracked plan SHA-256 `07b40dadc48662ca402a94b67aacc1ed52a7db26a8af8a8d2f5d4912d2eb0179`. Four independent lenses launched / four returned. Read-only review; no browser runs or tests.

The core calculation is correct **for the measured CSS root-font simulation**, provided the mobile override is also updated:

- Admin `--card-padding` is **16px**, overriding the base 24px.
- Current chain: `64 + 66 + 36 + 33 + 66 + 50 = 315px`. The mobile module override explains its 36px contribution.
- Proposed at 390/200%: `64 + 8×9.75 + 11 = 153px`, leaving **237px** against the measured **233px** trigger minimum.
- At 1024/1440 and 100%, every affected token remains below the cap. At 768, module start padding changes from 24px to 19.2px.

No unnecessary-complexity finding: existing selectors plus native `min()` are a suitably small solution.

1. **Major — acceptance includes a dialog the solution cannot resize.** [plan-f3.md:46](/Users/strelia/projects/courses/docs/multi-subtitles/plan-f3.md:46)  
   The rename dialog is outside the targeted container chain. Existing overlay padding, dialog padding and borders leave approximately `390 − 64 − 96 − 2 = 228px`, below the measured 233px trigger minimum. Its `overflow:auto` can hide horizontal overflow from the document-level check. Add a scoped dialog solution or demonstrate its existing fit; F-R2 must measure the dialog’s own `scrollWidth/clientWidth`, selected option visibility and focus return.

2. **Major — 320px at 200% fails the proposed fit budget.** [plan-f3.md:47](/Users/strelia/projects/courses/docs/multi-subtitles/plan-f3.md:47)  
   Available width becomes `320 − 64 − (8×8 + 11) = 181px`, below 233px. F-R3 tests only 100%, so it cannot expose this. Add the combined boundary and revise the spacing/layout strategy. Removing every nested padding would leave only 245px, making the fixed shell boundary consequential.

3. **Major — the measured trigger is not the largest possible child.** [plan-f3.md:45](/Users/strelia/projects/courses/docs/multi-subtitles/plan-f3.md:45)  
   Valid audio labels can contain 64 unbroken characters. They render in an unwrapped `<strong>` within the track header; flex wrapping does not split that item. Padding reductions cannot bound its min-content width. Add a long-label fixture and explicit handling for user-provided text. Also exercise long module/lesson titles, legends, populated transcript document-selection fieldsets and ASR states. Inputs already have `min-inline-size:0`; buttons retain their minimum size—these should be verified, not shrunk to obtain a pass.

4. **Major — simulated resizing does not establish browser text-zoom behavior.** [plan-f3.md:9](/Users/strelia/projects/courses/docs/multi-subtitles/plan-f3.md:9)  
   Root-font resizing doubles these rem paddings while leaving `vw` unchanged. That is a valid reproduction, but it is not interchangeable with every browser zoom mechanism. Page zoom changes the initial viewport; Firefox separately offers text-only zoom. Record browser/version, zoom mode, actual viewport width and computed paddings, and add a real text-only zoom check. Scope “vw does not grow” to the tested mechanism. [CSSOM View](https://www.w3.org/TR/cssom-view-1/), [Firefox zoom documentation](https://support.mozilla.org/en-US/kb/font-size-and-zoom-increase-size-of-web-pages).

5. **Minor — “desktop unchanged” needs a pinned comparison.** [plan-f3.md:48](/Users/strelia/projects/courses/docs/multi-subtitles/plan-f3.md:48)  
   Specify the baseline revision, identical fixture, locale, fonts, viewport and scrollbar conditions. Compare both inline paddings and relevant bounding boxes; computed padding alone does not prove unchanged layout. Limit the claim explicitly to **100%**. F-R1/F-R2 should check nested overflow as well as document overflow; F-R6 supplies regression coverage, not those measurements.

6. **Minor — make cascade and direction handling explicit.** [plan-f3.md:27](/Users/strelia/projects/courses/docs/multi-subtitles/plan-f3.md:27)  
   Update the later mobile override at [main.css:1455](/Users/strelia/projects/courses/frontend/src/app/styles/main.css:1455). Existing module/lesson shorthands set physical left padding; adding only `padding-inline-start` can leave that padding on the opposite side under RTL. Specify logical padding consistently and verify both directions.

7. **Minor — correct supporting wording.** [plan-f3.md:35](/Users/strelia/projects/courses/docs/multi-subtitles/plan-f3.md:35)  
   Affected tokens span **12–24px**, not 16–24px. W5 F5 establishes the `:has()` compatibility problem, not the separate container-query claim.

**Four open blocking findings: items 1–4.** F-R1–F-R6 are a useful starting set, but do not yet establish the claimed fit across the required states.