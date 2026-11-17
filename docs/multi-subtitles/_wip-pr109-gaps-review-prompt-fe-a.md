# Review: package `gaps-fe-a` (tests closing 74 mutation GAPs, frontend transcript)

Independent reviewer, read-only: no edits, no git writes, no Docker. (Rstest may not run in your sandbox; `npm run typecheck` does.)

- Delta: uncommitted test files under `frontend/src/modules/transcript/` in the working tree (frozen
  `docs/multi-subtitles/_wip-pr109-gaps-fe-a.patch`, sha256 `a461ad43…`). Branch `ms2-3-asr-per-track` @ `46c773b`. Other dirty files
  belong to other sessions — out of scope.
- Inputs: the GAP rows and proposed tests in the fe-a mutation journal (outside the repo:
  `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/_wip-pr109-mutations-fe-a.md`,
  large — grep row ids) and the closure journal `docs/multi-subtitles/_wip-pr109-gaps-fe-a-journal.md` (per row: test, baseline PASS,
  mutated FAIL line). Contract: `/Users/strelia/project-data/ai-factory-kit/skills/arch/code-review/references/behavior-coverage.md`.

Check:
1. Only test files changed; no production change; no existing assertion weakened (6 removed lines).
2. Each closing test asserts **observable behavior** with independently specified expected values, not implementation echoes or
   mocks of the unit under test; races use events/deferred promises, not sleeps.
3. Sample at least 15 rows across all gap classes (schema fields, accepted-source cleanup fields, hook lifecycle, detached hook,
   draft repository): does the named test actually reach the mutated path, so the quoted failure is attributable to that mutation?
4. Flakiness risk (timers, ordering, global state such as fake IndexedDB or MSW handlers leaking between tests).
5. Standards: `AGENTS.md` TypeScript, `.claude/standards.md`.

Finding format: `file:line`, impact, how verified, BLOCKER or non-blocking, confidence. End with `VERDICT (gaps-fe-a): SHIP | BLOCK`
and counts.
