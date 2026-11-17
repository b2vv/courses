# Codex review: PR #109, Standards axis (package `pr109-final` @ 38c7915)

You are the independent **Standards-axis** reviewer. Read-only: no edits, no git writes, no tests, no Docker. Read the sources
yourself and do not trust author summaries.

1. Read the pinned package `docs/multi-subtitles/_wip-pr109-review-package.md`. It gives the scope, refs, sources, evidence and
   constraints.
2. Read the review contracts it lists:
   - `/Users/strelia/project-data/ai-factory-kit/skills/arch/code-review/SKILL.md`: the Standards brief and the Fowler smell baseline
     in step 3;
   - its `references/behavior-coverage.md`;
   - `/Users/strelia/project-data/ai-factory-kit/ai-factory/review-handoff.md` §"Final-review coverage and ADR decisions".
3. Review `git diff 37ec4b0...38c7915` against the documented standards: `.claude/standards.md`; `AGENTS.md` Architecture,
   TypeScript, Decisions and Verification; `docs/adr/*`. Report:
   - (a) violations of documented standards, citing the file and the rule;
   - (b) baseline smells. These are always judgement calls, and a repo standard overrides them.
   - (c) behavior-coverage for changed guards, filters, serializers and invariants: do tests exist that would discriminate the
     behavior? Mutation execution is not authorized, so mark it NOT_TESTED.
   - (d) existing-ADR compliance and freshness, and whether new decision material needed an ADR. Give evidence or a justified N/A.

   Skip what tooling enforces. Prioritise:
   - the backend domain/application/ports/adapters boundaries;
   - the migration;
   - cross-module contracts;
   - frontend editor and tab state;
   - the Makefile `e2e-reset`.
4. Finding format: `file:line`, impact, how verified, BLOCKER or non-blocking, and confidence. Mark anything unresolved as
   UNRESOLVED.
5. End with `VERDICT (Standards, pr109-final @ 38c7915): SHIP | BLOCK` and the counts by severity.
