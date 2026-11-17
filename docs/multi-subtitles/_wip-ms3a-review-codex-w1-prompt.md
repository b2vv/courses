# MS3a review, wave 1 — independent, read-only

You review an UNCOMMITTED frontend diff on branch `ms2-3-asr-per-track` @ `a985543` (repo = current directory). Read-only:
do not edit files, do not run git commands that change state. Shared working tree: ignore unrelated dirty files
(`docs/security/_wip-*`, `.gitignore`, `docs/development/ai-factory-package.json`, `backend/upload`, `.codex/`,
`docs/ux-review/`, other `_wip-*`). The diff under review: `git diff -- frontend/src` plus the untracked
`frontend/src/modules/admin/ui/LessonLanguageTabs.tsx`, `LessonLanguageTabs.test.tsx`,
`frontend/src/modules/transcript/domain/draft.test.ts`.

Authority: `docs/multi-subtitles/plan-ms2-3.md` (ACCEPTED r9) — MS-D6, MS-D7, MS-D9 (key part only), MS-D10 (frontend);
rows M3-R1, R2, R3, R5, R7, R10, R14, R15 and the save-path parts of R11 and R13; owner decisions O7, O8.
Scope split and interim rules: `docs/multi-subtitles/report.md` → "MS3a … checkpoint" (the `und` dialog, detached
snapshots / flush on unmount and `baseLanguage` are MS3b — do not report their absence as a finding; DO report if the
interim `und` rule "Save refused, 0 PUT" is violated). Project rules: `AGENTS.md` (strict TS, validate external data,
no fetch in presentation/routes, modules do not import each other, React Aria, >=44px, visible focus).
Implementation notes: `docs/multi-subtitles/_wip-ms3a-run1.md`, `_wip-ms3a-run2.md`.

Courses 1 already fixed two browser findings after run 2 (check they are correct, not that they exist):
tabs computed before the tracks list loaded (an orphan tab got selected and kept), now waiting for both lists with an
alert on a failed read; and two switches sharing the name "Multiple languages", the row one now named per track.

Look hardest for: a PUT that can be sent when the plan says 0 PUT (re-read failure, track gone/renamed, `und` tab); a
token silently advanced from the inventory; a draft loadable when `draftLoadable` says no (incl. mode mismatch under
the `und` exception); lines lost on a base change while dirty; React key/remount churn that drops edits; duplicate DOM
ids or wrong accessible names across mounted editors; stale-draft polling still present (O7); pending markers shared
between tracks; cache not invalidated after PATCH; `isDisabled` on a focused control dropping focus (the run-2 decision
disables line controls during an in-flight Save — check focus behaviour); strict zod schemas that reject a field the
backend now sends.

Output: a list of findings, each with severity (Blocker / Major / Minor), file:line, the concrete failure scenario
(inputs/state → wrong result), and the plan row it breaks. Then a verdict: SHIP or BLOCK. No findings → say so.
