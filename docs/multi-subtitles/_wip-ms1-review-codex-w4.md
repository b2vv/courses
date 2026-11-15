Reviewed `8557d24..30c4031` on `multi-subtitles`. Read-only; no product tests or live browser interactions run. Opened all seven new screenshots. Unchanged wave-3 evidence remains accepted.

**Wave-3 dispositions**

| Finding | Disposition |
|---|---|
| U2 at 390 lacked a screenshot | **Resolved.** Opened `390-U2-und-only-open.jpg`: all six expected languages are visible, including English; Undetermined is absent. |
| U2/U4 at 200% marked N/A | **Evidence gap resolved; U4 fails on F4.** Opened `390-text200-U2U4-keyboard.jpg` and `390-text200-U4-end-scrolled.jpg`: English has visible focus, and the scrolled list exposes Spanish with focus. Opened both dialog images: the initial focused selection is hidden; the after-arrow image shows it revealed. |
| U2/U4 at 320 marked N/A | **Resolved.** Opened `320-reflow-U2U4-keyboard.jpg` and `320-reflow-U4-dialog-keyboard.jpg`: options fit and keyboard focus is visible in both surfaces. Escape and focus-return behavior remains recorded interaction evidence, not something these still images independently prove. |
| Diagnostic ignored election | **Resolved.** Candidate filtering and election now match migration step 2. |
| Previously resolved gate, provenance and other viewport findings | **Remain resolved.** This delta changes documentation/evidence only. F3 remains a recorded pre-existing Minor. |

**Diagnostic verification**

[diagnostic.sql:23](/Users/strelia/projects/courses/docs/multi-subtitles/ms1-diagnostic/diagnostic.sql:23) ranks candidates by default first, embedded second, then ascending track ID—the migration’s ordering.

I exercised the wave-3 counter-example using the exact diagnostic SQL in an in-memory SQLite database, and the migration’s selection predicate/order with `LIMIT 1` replacing `DISTINCT ON` for the single asset:

| Track | Diagnostic | Migration election |
|---|---|---|
| Default embedded `und` | `eligible (elected)` | Selected → `uk` |
| Non-default uploaded `und-x-original` | `candidate, not elected` | Unchanged |

This was a query-level check, **not a PostgreSQL integration test**. Static comparison also confirms matching Ready/non-evicted, non-failed, primary-transcript and existing-primary-language conditions. The supplied before/after captures remain valid: each captured asset has only one candidate.

**Ranked findings**

- **Blocker:** None.

- **Major — F4 violates the checked keyboard-focus requirement.** [report.md:139](/Users/strelia/projects/courses/docs/multi-subtitles/report.md:139), underlying scrolling CSS at [tokens.css:262](/Users/strelia/projects/courses/frontend/src/app/styles/tokens.css:262). At 390px with 200% text, opening the rename list leaves the focused selected option entirely outside the visible popover. The initial screenshot shows no visible option focus; the after-arrow screenshot demonstrates recovery.

  **Pre-existing: supported. Outside MS1’s implementation diff: supported. Outside its verification scope: no.** The complete `git diff 5cbb44b..HEAD -- frontend` changes language validation, filtering, display and tests; Select markup, scrolling CSS and dependencies are unchanged. However, [plan-ms1.md:144](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms1.md:144) explicitly requires visible keyboard focus, including at 200%.

  “Minor, follow-up” is not defensible here: complete focus disappearance is a functional accessibility failure in the checked control, rather than F3’s usable wrapping. Arrow recovery limits its severity to Major. The repository-pinned [design-pass §6](/Users/strelia/project-data/ai-factory-workspace/runtime-data/ai-factory-rollout/candidates/17a9527421e8fac3e43a16c287c404118699833c/skills/design-pass/SKILL.md) requires “blocker/major виправлені й переперевірені,” except for an explicitly accepted, dated owner limitation. The report’s proposed follow-up supplies no such acceptance.

  **Fix F4 before MS1 closes**, recheck initial opening and keyboard recovery, and replace [report.md:102](/Users/strelia/projects/courses/docs/multi-subtitles/report.md:102)’s “PASS, except F4” with FAIL until resolved. A shared Select fix also needs the repository-required full e2e verification.

- **Minor:** F3 remains accepted as a pre-existing follow-up; no new Minor finding.

**VERDICT: BLOCK**