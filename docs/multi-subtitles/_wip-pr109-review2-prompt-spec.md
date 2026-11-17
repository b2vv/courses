# Codex re-review: PR #109, Spec axis (package `pr109-final-2` @ ad3612c)

You are the independent Spec-axis reviewer. Read-only: no edits, no git writes, no Docker, no tests needing Docker. Author evidence is a
claim to check.

1. Read `docs/multi-subtitles/_wip-pr109-review-package-2.md` and the contracts it inherits (`arch/code-review/SKILL.md` Spec brief,
   `references/behavior-coverage.md`, `review-handoff.md` §"Final-review coverage and ADR decisions").
2. Re-check the previous Spec verdict (`docs/multi-subtitles/_wip-pr109-review-codex-spec.md`):
   - F1: does `detachedTranscriptSnapshots` now follow `docs/multi-subtitles/plan-ms2-3.md:406-407` exactly, including the dated author
     decision on "no default track" in `report.md` (MS3d-fix w1 entry) — is that decision within the spec or a change the owner must make?
   - F2: do `docs/multi-subtitles/ms3-design-pass/` (screenshots, `measurements.json`, harness) and the report table give revision-bound
     evidence for M3-R7 and the visual part of M3-R14 (390/768/1024/1440 + 200 % + 320, keyboard)? Inspect several PNGs and the JSON.
   - F3: does the executed mutation evidence + 208 closed gaps satisfy the behavior-coverage contract?
3. Produce the §4 table again for all rows: row id → implemented? → discriminating test → PASS / PARTIAL / MISSING / NOT_TESTED.
   Rows whose only proof is the final e2e stay NOT_TESTED until the gate on `ad3612c` runs — say which.
4. Review `5ef53d1` (copy) and any spec impact of the delta; scope creep.
5. End with `VERDICT (Spec, pr109-final-2 @ ad3612c): SHIP | BLOCK` and counts by severity.
