1. **MAJOR — Wave-1 #1 CLOSED**, `frontend/e2e/subtitles.spec.ts:88`. `resetProgress()` now invokes portable `docker compose` with the inherited environment, removing the machine-specific `ENOENT` failure.

2. **MAJOR — Wave-1 #2 CLOSED**, `frontend/src/modules/transcript/ui/TranscriptPanel.tsx:35`. Errors now use `alert`; loading and absence states use `status`. Scoped tests assert exactly one appropriate live region.

3. **MAJOR — Wave-1 #3 OPEN: PENDING**, `docs/multi-subtitles/report.md:948`. The ordered gate/M25 and design-pass/M26 remain scheduled owner work. They are **not code blockers for this review**.

4. **MINOR — Wave-1 #4 CLOSED**, `frontend/e2e/subtitles.spec.ts:1060`. E9 now polls advancing playback with a 15-second bound, removing its fixed readiness sleep.

5. **MAJOR — NEW: teardown retains globally counted fixtures**, [frontend/e2e/subtitles.spec.ts:646](/Users/strelia/projects/courses/frontend/e2e/subtitles.spec.ts:646). A successful run creates a course, student grant and three Ready video assets; teardown only archives the course. `adminstats/adapters/postgres/snapshot.go` still counts archived courses, active grants and Ready assets without excluding archived courses. Thus every repeat permanently increases those shared totals, violating AGENTS.md’s cleanup rule. The journal already records eleven retained archived fixtures. **Required change:** use the neighbouring `audio-tracks.spec.ts` run-scoped deletion pattern to remove the course and dependent grant/audit/media/progress records, including after partial setup failure.

No further standards findings in the fix delta. Static review only; no tests executed.

VERDICT: BLOCK