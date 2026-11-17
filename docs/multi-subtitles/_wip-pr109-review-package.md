# Review package `pr109-final` — PR #109 (multi-subtitles MS2+MS3)

- **Task:** multi-subtitles MS2+MS3. The author and writer is the Courses 1 session (Claude, with Codex as implementer). The
  reviewers are two fresh independent subagents (owner, 2026-11-17: "Ревʼю саб агентом за правилами ai factory"), one per axis
  (Standards, Spec). This is the route in `docs/development/ai-factory.md` §Skills/review, using the shared
  `arch/code-review` two-axis contract.
- **Root and refs:** root `/Users/strelia/projects/courses`, branch `ms2-3-asr-per-track`, base `main@37ec4b0` (merge-base
  `37ec4b0aa69b8fb2d6856dd83301b78fe508ba7b`), HEAD `38c7915`. PR: https://github.com/b2vv/courses/pull/109.
- **Pinned scope:** `git diff 37ec4b0...38c7915` (211 files, +16002/−2616) and `git log --oneline 37ec4b0..38c7915` (10 commits).
  The working tree also has unrelated dirty and untracked files from other sessions (`docs/security/_wip-*`, `docs/multi-audio/*`,
  `.gitignore`, `docs/development/ai-factory-package.json`, …). They are **out of scope**; do not review them.
- **Spec sources:**
  - `docs/multi-subtitles/plan-ms2-3.md` (sha256 prefix `e27efa6a519ba330`, ACCEPTED r9): §0 owner decisions, §3 decisions,
    §4 acceptance rows, §5 verification, §6 limits;
  - the phase plan `docs/multi-subtitles/plan.md`.
- **Standards sources:**
  - `.claude/standards.md` (sha256 prefix `a16c3508897ab7d0`);
  - `AGENTS.md`, sections Architecture, TypeScript, Decisions and Verification;
  - the ADRs in `docs/adr/` (`013-one-access-rule-as-a-contract.md` is the cross-module rule).
- **Evidence (raw, author-produced; verify independently where you can by reading):**
  - `docs/multi-subtitles/report.md` (sha256 prefix `bc209d8ab7ffc99d`), the MS2a…MS3c-reset entries and the gate attempts 1–7;
  - the final ordered `make check`, attempt 7, on the Kiri engine, 2026-11-17 10:34–11:06: `GATE_EXIT=0`; frontend tests 1186
    passed; 127 Go `ok` lines and 0 FAIL; **e2e 151 passed, 0 failed, 0 did not run**. The raw log is outside the repo, at
    `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/ms3c-make-check-7.log`;
  - the per-slice Codex reviews are in `docs/multi-subtitles/_wip-*review-codex*.md`.
- **Contracts to read before reviewing:**
  - `/Users/strelia/project-data/ai-factory-kit/skills/arch/code-review/SKILL.md`;
  - its `references/behavior-coverage.md`, including the Anti-patterns section;
  - `/Users/strelia/project-data/ai-factory-kit/ai-factory/review-handoff.md` §"Reviewer contract" and §"Final-review coverage and
    ADR decisions";
  - `/Users/strelia/project-data/ai-factory-kit/ai-factory/reference-anti-patterns.md`.
- **Constraints:**
  - **Read-only.** Do not edit any repository file, and run no `git` writes (checkout, reset, stash, add, commit).
  - Do not run tests or Docker; the shared Kiri stand belongs to the owner.
  - Mutation and behavior-coverage *execution* is **not authorized** in this round, so report it as NOT_TESTED with what would be
    needed, never as PASS.
  - Do not repeat any credential values you may see in `frontend/e2e/auth.ts`.
- **Finding format:** `file:line`, impact, how you verified it, BLOCKER or non-blocking, and confidence. Mark anything unresolved
  as UNRESOLVED. A missing proof means NOT_TESTED, not PASS.
