# Review: package `gaps-fe-b` (tests closing 60 mutation GAPs, frontend admin/media/composition)

Independent reviewer, read-only: no edits, no git writes, no Docker. (Rstest may not run in your sandbox; `npm run typecheck` does.)

- Delta: uncommitted test files in the working tree: `frontend/src/app/composition/admin-{audio,draft}-queries.test.ts`,
  `frontend/src/modules/admin/ui/{CourseEditorView.test.tsx,LessonLanguageTabs.test.tsx,LessonLanguageTabs.lifecycle.test.tsx}`,
  `frontend/src/modules/media/adapters/audio-track-repository.test.ts`, `frontend/src/modules/media/ui/LessonAudioTracksField.test.tsx`,
  `frontend/src/routes/route-guards.test.tsx`. Frozen patch `docs/multi-subtitles/_wip-pr109-gaps-fe-b.patch` (sha256 `dbfd3ad0…`).
  Branch `ms2-3-asr-per-track` @ `16489a5`. The backend test files and `frontend/src/adapters/api/e2e-reset.test.ts` (+ fixture) belong to
  package `gaps-be-b`, reviewed separately; other dirty files belong to other sessions — out of scope.
- Inputs: GAP rows + proposed tests in the fe-b mutation journal (outside the repo:
  `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/_wip-pr109-mutations-fe-b.md`, grep row ids) and the closure journal `docs/multi-subtitles/_wip-pr109-gaps-fe-b-journal.md`.
  Contract: `/Users/strelia/project-data/ai-factory-kit/skills/arch/code-review/references/behavior-coverage.md`.
- Author's timing note: under host load ≈ 29 the full suite had 3–5 timeouts (5 s / 20 s) in old `CourseEditorView` tests and twice in two
  new `route-guards` tests; at load ≈ 20 the new route-guards tests take 331–415 ms and runs 4–5 were green (1306 passed). Judge whether any
  new test adds avoidable cost or timing sensitivity.

Check: 1) only test files changed, no assertion weakened (9 removed lines); 2) observable behavior, independently specified expectations,
races via events/deferred promises; 3) sample ≥ 15 rows across composition queries, CourseEditorView, LessonLanguageTabs (+ lifecycle),
media repository, LessonAudioTracksField, route-guards — the named test reaches the mutated path and the failure is attributable;
4) flakiness/leaks (timers, MSW handlers, QueryClient, global stubs); 5) standards (`AGENTS.md` TypeScript, `.claude/standards.md`),
file naming (`LessonLanguageTabs.lifecycle.test.tsx`).

Finding format: `file:line`, impact, how verified, BLOCKER or non-blocking, confidence. End with `VERDICT (gaps-fe-b): SHIP | BLOCK` and counts.
