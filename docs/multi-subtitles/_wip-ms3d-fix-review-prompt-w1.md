# Review: package `ms3d-fix-w1` (PR #109 review defects, frontend transcript recovery)

You are an independent reviewer. Read-only: no edits, no git writes, no Docker. You may run read-only frontend checks
(`cd frontend && npx rstest run src/modules/transcript src/modules/admin`, `npm run typecheck`) if useful.

- Base: branch `ms2-3-asr-per-track` @ `38c7915`; the delta is the **uncommitted** working-tree diff under `frontend/`
  (`git diff -- frontend/`; frozen copy `docs/multi-subtitles/_wip-ms3d-fix-review-diff.patch`, sha256 in the report). Other dirty
  files in the tree belong to other sessions: out of scope.
- Defects being fixed: `docs/multi-subtitles/_wip-pr109-review-codex-standards.md` finding 1 (capture version = `Date.now()`
  collision) and `docs/multi-subtitles/_wip-pr109-review-codex-spec.md` F1 (detached snapshot offered in an orphan `lang:*` tab;
  spec `docs/multi-subtitles/plan-ms2-3.md:398-415`). Checkpoint: the MS3d-fix entry at the end of `docs/multi-subtitles/report.md`.
- Author journal (claims to check, not proof): `docs/multi-subtitles/_wip-ms3d-fix.md`.

Check:
1. Is every identity comparison moved from `capturedAt` to `revision` (the class, not one place)? Any remaining `capturedAt`
   used as identity, or any path writing a snapshot without a fresh revision?
2. Does the destination now follow the spec exactly (matching *track* tab, else default track), including a lesson with zero
   tracks, a non-default track list, and case-insensitivity?
3. Do the new tests discriminate (equal `capturedAt`, different revisions; orphan-only)? Are any existing assertions weakened?
4. Regressions: snapshots without `revision` read as null — consequences for recovery, viewer sweep (`keys()`), e2e that plant
   snapshots (`frontend/e2e/transcript.spec.ts:567`, `frontend/e2e/asr.spec.ts`); `crypto.randomUUID` availability.
5. Standards: `AGENTS.md` TypeScript/Architecture, `.claude/standards.md`.

Finding format: `file:line`, impact, how verified, BLOCKER or non-blocking, confidence. End with
`VERDICT (ms3d-fix-w1): SHIP | BLOCK` and counts by severity.
