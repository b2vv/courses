# Re-review of design-pass-auth

## Step 1 — start, before first tool inspection
- Scope: re-review HEAD fixes for the three findings in `docs/design-pass/_wip-review.md`, then inspect the fix diff for new defects, including the `import.meta.url` relative depth from `docs/design-pass/s0/`.
- Constraints: read-only source inspection through git show/diff; no checkout, Docker, or make. Only this working file may be written.
- Status: review started; no findings assessed yet.

## Step 2 — entry rules loaded
- FACT: Read the shared `agent-entry.md` before skills or substantive review.
- DECISION: The owner's explicit target and single writable review file define this task; no checkout, test execution, or source mutation is authorized.
- NEXT: Resolve repository/local review instructions and pin the fix commit, then compare the prior review with its diff.

## Step 3 — pinned scope and local rules
- FACT: Verified repository root `<repo>`; read `AGENTS.local.md`, the repository briefing, and review skill instructions.
- FACT: Fix commit pinned to `b912c54f4a6ef8975d64c38a6e1a3fa547d5d021` (HEAD), `docs(design-pass): repo-relative script paths, current status, post-port hash map (review)`.
- DECISION: Apply the code-review standards/specification lenses directly. The owner's single-writable-file rule overrides the skill's separate subagent working files; agents launched: 0, returned: 0.
- NEXT: Read the complete fix diff and previous findings, then verify the resulting scripts, plan/scope, and hash mapping against committed source.

## Step 4 — fix inventory and evidence recovery
- FACT: HEAD changes exactly seven files: `findings.md`, `plan.md`, and the five S0 scripts (`capture`, `discover`, `kbd`, `pairs`, `sheet`); no product code is changed by this fix.
- FACT: All five literal `<repo>` module roots are replaced with `new URL('../../../frontend…', import.meta.url)` forms. Plan now declares AUTH implemented and documents shared notice-form CSS effects; findings add `97642485` as the post-port closure commit.
- FACT: The prior review is untracked (not present in HEAD); read it using `git diff --no-index /dev/null` so its content remains read-only. The canonical skill is under `skills/arch/code-review/`, not the absent flat kit path.
- NEXT: Check full scripts and prior evidence for path semantics, current status/scope accuracy, and independently inspect the mapped port commits.

## Step 5 — initial dispositions
- FACT: Prior finding 1 concerns invalid `createRequire` inputs; from `docs/design-pass/s0/<script>.mjs`, `../../../` traverses `s0 → design-pass → docs → repository root`, so the new depth targets the repository's `frontend/`, not `docs/frontend/` or a parent checkout.
- FACT: Prior finding 2 is addressed in plan lines 9/102 (current status) and 129 (shared `.notice form`/non-field controls and primary-link CSS, including root/login retry forms explicitly marked NOT_TESTED).
- FACT: Prior finding 3 now has the original-to-port map at plan line 7 and the new closure references at findings lines 56–58. `97642485` commit metadata independently records cherry-pick source `5518cda0…`.
- NEXT: Finish bounded source reads (earlier combined output was truncated), confirm historical evidence has not been relabelled, and check affected CSS consumers for any new discrepancy.

## Step 6 — provenance and status cross-check
- FACT: All five old→new plan mappings match the port commit messages, including `5518cda0 → 97642485`. The original evidence file still identifies the measured revision as `5518cda`; it is not being represented as a new HEAD run.
- FACT: Plan §10 retains explicit AUTH-1…4 acceptance and requires an ordered gate before PR; the fix changes current status/scope descriptions without claiming a new runtime pass. Remaining Q3/Q4 are still explicit.
- FACT: A whole-frontend comparison of the historical and port trees includes unrelated rewritten-main differences and is unsuitable for proving the auth patch's equivalence; scope subsequent comparisons to the auth files/commit patches. No finding is inferred from that broad diff.
- NEXT: Finish exact script-path review and inspect current shared-CSS consumers and evidence limits.

## Step 7 — script regression check
- FACT: Read all five affected scripts in full. `kbd.mjs:2`, `pairs.mjs:3`, and `sheet.mjs:3` pass a file URL directly to `createRequire`; their three-level relative paths are correct. `frontend/package.json` exists in the committed tree.
- FACT: `capture.mjs:3–6` and `discover.mjs:3–6` resolve the correct absolute frontend path at this review root, so the original literal-placeholder startup failure is removed.
- OPEN [P3, non-blocking]: `docs/design-pass/s0/capture.mjs:3` and `docs/design-pass/s0/discover.mjs:3` use `URL.pathname` as a filesystem filename. A checkout path containing a space (e.g. `/tmp/course review/`) becomes `/tmp/course%20review/frontend`, and `createRequire`/`readFileSync` then target the wrong directory. Use `fileURLToPath(new URL(...))` (or keep file URLs for the Node APIs). This does not affect the verified `<repo>` path and does not invalidate the depth fix. Static reasoning only; no script executed.
- NEXT: Finish product/CSS and historical-evidence cross-check, then record final CLOSED/OPEN dispositions and scoped verdict.

## Step 8 — final verification and verdict
- FACT: Inspected the AUTH product diff from `bdc7048` and the current root/login retry forms. The forms at `frontend/src/routes/__root.tsx:78` and `frontend/src/routes/login.tsx:34` match the extra CSS consumers now named in the plan. HEAD introduces no further product/CSS change.
- FACT: Scoped `git diff --name-only 5518cda0 97642485` for the four auth views and their two test files is empty: those files remain byte-identical across the port. No runtime equivalence is inferred for the entire rewritten tree.
- FACT: `git diff --check HEAD^ HEAD` completed successfully (exit 0); final HEAD remains `b912c54f4a6ef8975d64c38a6e1a3fa547d5d021`. The fix inventory is still the same seven files.

### Original findings
1. **CLOSED — P2 script startup failure.** `docs/design-pass/s0/capture.mjs:3`, `docs/design-pass/s0/discover.mjs:3`, `docs/design-pass/s0/kbd.mjs:2`, `docs/design-pass/s0/pairs.mjs:3`, `docs/design-pass/s0/sheet.mjs:3`. No literal `<repo>` reaches `createRequire`; three parent traversals from `docs/design-pass/s0/` correctly reach the root and then `frontend`. Valid absolute filenames at this repository root, or file URLs, now reach `createRequire`.
2. **CLOSED — P3 stale status and undeclared shared CSS impact.** `docs/design-pass/plan.md:9`, `docs/design-pass/plan.md:102`, `docs/design-pass/plan.md:129`. Current status distinguishes completed S0/implemented AUTH from unstarted slices; scope names notice-form layout, non-field alignment, auth alternatives, primary-link override, and the retry consumers. Extra consumers are explicitly visually NOT_TESTED, not silently claimed covered.
3. **CLOSED — P3 missing post-port provenance.** `docs/design-pass/plan.md:7`, `docs/design-pass/findings.md:56` (also rows at 57/58). Explicit `5518cda0 → 97642485` mapping agrees with cherry-pick metadata. `docs/design-pass/evidence/AUTH.md:5` and `work/tasks.md:496` keep the original measured hash; the central map makes it traceable without falsely relabelling measurements.

### Standards lens
- **OPEN — P3, non-blocking:** `docs/design-pass/s0/capture.mjs:3` and `docs/design-pass/s0/discover.mjs:3` use URL `.pathname` instead of decoding with `fileURLToPath`. Space/non-ASCII/percent-encoded checkout paths can fail dependency or file lookup. The verified repository root is unaffected. Recommended correction: `fileURLToPath(new URL('../../../frontend', import.meta.url))`, or URL inputs throughout.
- No other new standards finding in this seven-file delta. Existing-ADR impact and new ADR material: N/A; the fix changes documentation and local evidence-script resolution, with no architectural, API, persistence, or business-policy decision.

### Specification lens
- All three requested corrections are present. No new blocking specification finding. Depth verified by source/path arithmetic; all five mapped commits independently matched their cherry-pick metadata.
- Behavior/runtime verification: NOT_RUN by explicit owner instruction. No scripts, unit tests, browser, Docker, make, checkout, or gate were executed. Historical AUTH evidence remains historical; `docs/design-pass/evidence/AUTH.md:40` still records full ordered gate NOT_RUN. This verdict is the requested static re-review, not merge/gate certification.

**VERDICT: PASS** for the fix re-review at `b912c54`. Original findings: 3 CLOSED / 0 OPEN. New findings: 0 blocking / 1 non-blocking portability issue. Standards: 1 P3; specification: 0 open findings. Agents: 0 launched / 0 returned. Only `docs/design-pass/_wip-review-2.md` was written.
