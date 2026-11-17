# Codex review: PR #109, Spec axis (package `pr109-final` @ 38c7915)

You are the independent **Spec-axis** reviewer. Read-only: no edits, no git writes, no tests, no Docker. Read the sources yourself
and do not trust author summaries.

1. Read the pinned package `docs/multi-subtitles/_wip-pr109-review-package.md`. It gives the scope, refs, sources, evidence and
   constraints.
2. Read the review contracts it lists:
   - `/Users/strelia/project-data/ai-factory-kit/skills/arch/code-review/SKILL.md`: the Spec brief;
   - its `references/behavior-coverage.md`;
   - `/Users/strelia/project-data/ai-factory-kit/ai-factory/review-handoff.md` §"Final-review coverage and ADR decisions".
3. The spec is `docs/multi-subtitles/plan-ms2-3.md` (ACCEPTED r9): §0 owner decisions, §3 decisions, §4 acceptance rows,
   §5 verification, §6 limits. Use `docs/multi-subtitles/plan.md` for phase context. Does `git diff 37ec4b0...38c7915` faithfully
   implement it? Report:
   - (a) missing or partial requirements and acceptance rows. Check each §4 row against the code **and** against a named test that
     would actually discriminate the behavior.
   - (b) scope creep;
   - (c) behavior that looks implemented but wrong;
   - (d) whether the §0/§3 owner decisions and the ADRs are honored.

   Quote the spec line for each finding. Mutation execution is not authorized: discrimination you cannot prove is NOT_TESTED. The
   author's evidence in `docs/multi-subtitles/report.md` is a claim to check, not proof.
4. Include a §4 table: row id → implemented? → test → PASS, PARTIAL, MISSING or NOT_TESTED.
5. Finding format: `file:line`, spec quote, impact, how verified, BLOCKER or non-blocking, and confidence. Mark anything unresolved
   as UNRESOLVED.
6. End with `VERDICT (Spec, pr109-final @ 38c7915): SHIP | BLOCK` and the counts by severity.
