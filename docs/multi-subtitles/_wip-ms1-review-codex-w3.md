Reviewed `f415643..8557d24` on `multi-subtitles`. Read-only; no tests run. Opened all 25 screenshots.

**Wave-2 dispositions**

| Finding | Disposition | Evidence opened and conclusion |
|---|---|---|
| Final-code ordered gate | **Resolved** | `report.md:43` records `GATE_EXIT=0`, 138 e2e and 948 frontend tests on `f415643`. The subsequent diff contains only documentation/evidence; accepting that recorded gate for unchanged code. |
| U1 at 200% | **Resolved** | Opened `390-text200-U1-open-top.jpg` and `…-scrolled.jpg`: list opens above the trigger, Polish has visible focus, and scrolling exposes Undetermined. F3 remains visible. |
| U2/U4 across baseline viewports | **Open, partly resolved** | U2 screenshots at 768/1024/1440 show English present and Undetermined absent. U4 screenshots at 390/768/1024/1440 show visible option focus; the new 390 dialog shows English selected. **No U2 screenshot at 390 exists.** |
| Retained-volume diagnostic | **Resolved for captured data; query qualification remains** | Opened `diagnostic.sql`, `before.txt`, `after.txt`, migration 00030 and the third-run report. Seven distinct assets before, six afterward; the sole candidate with primary `uk` disappears from the diagnostic as expected. |
| Screenshot provenance | **Resolved** | Opened both `before-fix-*` images and replacements. Old images show raw `EN`; replacements show English. The report now distinguishes them. |

**Ranked findings**

- **Major — required UI evidence remains incomplete.** [report.md:98](/Users/strelia/projects/courses/docs/multi-subtitles/report.md:98), [report.md:101](/Users/strelia/projects/courses/docs/multi-subtitles/report.md:101). U2 at 390 is marked PASS without a backing screenshot. U2/U4 at 200% and 320px remain applicable under [plan-ms1.md:146](/Users/strelia/projects/courses/docs/multi-subtitles/plan-ms1.md:146); the N/A rationale does not satisfy “each row.” Width-independent option generation establishes membership, but does not establish rendered fit or keyboard focus, scrolling and Escape return at constrained sizes. Complete these cells or obtain an explicit scoped acceptance change.

- **Minor — diagnostic “eligible” means candidate, not elected relabel target.** [diagnostic.sql:10](/Users/strelia/projects/courses/docs/multi-subtitles/ms1-diagnostic/diagnostic.sql:10). Its filtering matches the migration’s **candidate** predicate, but omits election at [00030:61](/Users/strelia/projects/courses/backend/migrations/00030_audio_track_language_shape.sql:61). Counter-example: one Ready, non-evicted asset with primary `uk`, no `uk` track, default embedded `und` and non-default uploaded `und-x-original`, both non-failed. The query calls both eligible; migration relabels only default `und`. Mirror the election or label candidates and distinguish non-elected rows. This does **not** invalidate the supplied before/after counts: every captured asset has only one candidate.

No Blocker finding.

F3’s **pre-existing, Minor, outside-MS1 classification is supported**. I inspected the complete `git diff 5cbb44b..HEAD -- frontend`: generated annotations, language validation/filtering/display and tests changed; no JSX layout or CSS changed. The screenshots support the narrow wrapping and usable list scrolling. Recording F3 separately is reasonable within §4a’s control scope; it does not excuse the missing U2/U4 checks.

**VERDICT: BLOCK** — remaining required UI evidence, with no new production-code defect established.