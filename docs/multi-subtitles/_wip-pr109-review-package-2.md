# Review package `pr109-final-2` — PR #109 (multi-subtitles MS2+MS3), re-review after `pr109-final` BLOCK

- **Task:** multi-subtitles MS2+MS3. Writer: Courses 1 session (Claude, Codex as implementer). Reviewers: two fresh independent
  read-only Codex runs, one per axis (Standards, Spec) of the shared `arch/code-review` contract — the owner's route (2026-11-17, "Codex").
- **Root and refs:** `/Users/strelia/projects/courses`, branch `ms2-3-asr-per-track`, base `main@37ec4b0`, HEAD **`ad3612c`**.
  Previous package `pr109-final` @ `38c7915` (`_wip-pr109-review-package.md`) and its verdicts `_wip-pr109-review-codex-{standards,spec}.md`.
- **Pinned scope:** `git diff 37ec4b0...ad3612c`. Delta since the previous review: `git log --oneline 38c7915..ad3612c` (8 commits):
  - `5a56965` MS3d-fix — Standards finding 1 (capture `revision`) and Spec F1 (detached destination = matching track tab, else default
    track); reviewed `ms3d-fix-w1` BLOCK → `w2` SHIP.
  - `5ef53d1` MS3d-copy — `uk` dialog strings name «Аудіодоріжки» (found by design-pass; not separately reviewed — review it here).
  - `7845300` design-pass evidence — Spec F2: `docs/multi-subtitles/ms3-design-pass/` (42 PNG, `measurements.json`, harness copy) and
    the "design-pass MS3" table in `report.md`.
  - `46c773b`, `bc22b82`, `16489a5`, `cc52ba4`, `ad3612c` — Standards F2 / Spec F3: mutation execution (owner: "Мутації в окремій копії")
    and gap closure (owner: "Закрити всі тестами"): 771 constructs, 500 KILLED, 208 GAP → **208 CLOSED** by tests-only packages
    `gaps-{fe-a,be-a,fe-b,be-b}`, each with an independent Codex SHIP (`_wip-pr109-gaps-review-codex-*.md`).
- **Your extra input:** `docs/multi-subtitles/_wip-pr109-mutation-dispositions.md` — 38 SURVIVED(EQUIVALENT), 24 E2E_ONLY, 1 INCONCLUSIVE,
  with the runner's evidence. The contract requires the reviewer to accept or reject each EQUIVALENT disposition.
- **Spec / standards sources:** as in `pr109-final` (`plan-ms2-3.md` ACCEPTED r9, `plan.md`; `.claude/standards.md`, `AGENTS.md`, `docs/adr/*`).
- **Evidence (author claims; verify by reading):** `docs/multi-subtitles/report.md` (MS3d entries onward); journals `_wip-ms3d-fix*.md`,
  `_wip-pr109-gaps-*-journal.md`. Large mutation journals live outside the repo:
  `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/_wip-pr109-mutations-*.md`.
  The final ordered `make check` has **not** run on `ad3612c` yet (needs the owner's consent; it resets the dev volumes).
- **Constraints:** read-only; no git writes, no tests that need Docker, no Docker. Do not repeat credential values from `frontend/e2e/auth.ts`.
  The working tree has unrelated dirty/untracked files from other sessions — out of scope.
- **Finding format:** `file:line`, impact, how verified, BLOCKER or non-blocking, confidence; UNRESOLVED where unresolved; missing proof =
  NOT_TESTED, not PASS.
