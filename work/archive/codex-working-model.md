# Working model: briefing Codex

Codex writes code; the reviewing session plans, briefs, verifies and commits. Codex's
sandbox has no Docker, so it can never run integration, e2e or the ordered gate — every
gate is re-run in the reviewing session before any commit, without exception.

## The brief contract

Every `work/brief-<task>.md` carries, in this order: a read-first list (plan sections,
sibling code, the exact acceptance rows the task owns), the task statement, a **NOT list**
(what belongs to other tasks), the environment limits, and the report path.

Standing rules in every brief:

- **Name the denial, never work around it.** A check that could not run is reported as not
  run, never as passed.
- **Never touch git**, never edit `work/tasks.md` (task status is the reviewing session's
  call, after the gates actually ran), never write the phase report (`docs/<phase>/report.md`
  belongs to the final task), never write inside `generated/`.
- **Missing contract → stop and report.** If a domain rule or port method the task needs
  does not exist, do not add it in the wrong layer; stop and say what is missing.

## Raising problems: stop, or discuss

Stopping is for a missing contract. It is not the only channel, and silence is never right.

**If Codex sees an inaccuracy that would produce a bug — in the plan, in the brief, or in
existing code — it opens a discussion instead of quietly working around it or quietly
complying.** This holds even when the instruction *could* be followed as written: following
a wrong instruction accurately is still shipping a bug.

The discussion is asynchronous, because Codex runs detached:

1. Codex writes a **`## Discussion`** section in its report (`work/report-<task>.md`) —
   what is wrong, the evidence (file:line, the plan clause, what would actually happen at
   runtime), the proposed correction, and what it did in the meantime (stopped, or
   implemented with the divergence flagged).
2. The reviewing session reads it, decides, and answers by resuming the same Codex session:
   `codex exec -s workspace-write --skip-git-repo-check resume --last "<answer>"`.
3. The outcome is recorded where it belongs: a plan defect becomes a dated plan amendment;
   a brief defect is corrected in the brief and named as the reviewer's error; a code defect
   becomes a fix with a test.

Codex may argue. If the answer is unconvincing it should say so once more with evidence
rather than comply against its judgement — a second refusal is cheaper than a defect.

## Why this shape

Both stops in Phase 5 paid for themselves. M04 stopped because six port signatures could
not express the atomicity the plan required; had it "helped" by adding the missing shapes in
the application layer, four later tasks would have been built on a bent seam. M08 stopped
because the thumbnail branch needed a duration no contract carried, and separately because
the brief contradicted the plan on process cancellation — the brief was wrong, and following
it would have shipped `Setpgid` behaviour the plan explicitly refuses.

The reviewing session is wrong often enough that this channel has to exist: of the problems
raised so far, one was a plan defect, one was a briefing error, and one was a field a gate
round had removed as unused while a consumer sat in the same document.
