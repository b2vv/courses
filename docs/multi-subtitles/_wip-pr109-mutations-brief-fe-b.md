# PR #109 mutation run — brief for Codex (part fe-b)

Owner decision 2026-11-17 ("Мутації в окремій копії"): per-construct mutations are executed in a disposable copy.
Contract: `/Users/strelia/project-data/ai-factory-kit/skills/arch/code-review/references/behavior-coverage.md` — read it
first and follow it exactly (one construct at a time, unchanged tests, restore baseline bytes, attributable failures only,
infrastructure failures are inconclusive, survivors are classified gap vs equivalent with evidence).

- **Disposable copy (the ONLY place you edit):** `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut` — a `git clone --local` of the repository at `5a56965`. Never edit,
  stage or run git writes in `/Users/strelia/projects/courses` (shared working tree, other sessions). Read it only if needed.
- **Scope (this part):** frontend/src/modules/admin, frontend/src/modules/media, frontend/src/app/composition/admin-*-queries.ts, frontend/src/routes/admin.courses.$courseId.tsx, frontend/src/app/providers/locales are excluded (copy only)
- **Pinned diff:** `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff 37ec4b0...5a56965 -- frontend/src/modules/admin frontend/src/modules/media frontend/src/app/composition frontend/src/routes` (production files only; tests, generated code, docs excluded
  from mutation but used as the detectors).
- **Runtime:** frontend tests `cd /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut/frontend && npx rstest run <files>` (check `package.json` for the exact runner);
  Go unit `cd /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut/backend && go test ./internal/<pkg>/...`; Go integration (only for SQL/repository constructs)
  `. /Users/strelia/project-data/courses-kiri/environment.sh && cd /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut/backend && go test -tags integration ./internal/<pkg>/adapters/postgres/...`
  (testcontainers on the Kiri engine; do NOT run `make`, `docker compose`, `make e2e-reset` or touch Courses compose volumes).
  e2e is out of scope: a construct only an e2e could detect is recorded as `E2E_ONLY` with the e2e row name, not killed.

WORKING FILE: `/private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mutlogs/_wip-pr109-mutations-fe-b.md` (append only; outside the copy; the only file you write outside it). First entry before the first tool: a `NEXT:` line. Then FACT / DECISION / NEXT / QUESTION.

## Steps
1. Inventory: enumerate every changed guard, filter clause, serializer and invariant in scope, including deleted constructs
   (reintroduce the old behavior as an inverse mutation). Write the inventory table first: id, `file:line`, construct,
   expected observable behavior, intended detector test.
2. Baseline: run the relevant tests unmutated; record command + pass counts. A red baseline → STOP + QUESTION.
3. For each construct: apply one mutation, run the relevant tests, record a row:
   `id | file:line | mutation diff (1–3 lines) | command | exit | failing test + assertion line (verbatim) | KILLED/SURVIVED/INCONCLUSIVE/E2E_ONLY`.
   Restore and verify with `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut diff --quiet -- <file>` before the next one.
4. Survivors: classify `GAP` (no test discriminates; propose the test at the agreed seam, do not write it) or `EQUIVALENT`
   (evidence why the behavior cannot change). Do not label them killed.
5. End: `git -C /private/tmp/claude-501/-Users-strelia-projects-courses/9307c360-464c-45bf-bb0f-3980a0267658/scratchpad/mut status --short` must be clean; counts KILLED / SURVIVED(GAP) / SURVIVED(EQUIVALENT) / INCONCLUSIVE /
   E2E_ONLY. Final answer = summary of the file.

Long scope: if you approach your time limit, stop at a construct boundary with the copy clean and write `NEXT: resume at <id>`.

## Survivor confirmation (added after part fe-a)
Before classifying a survivor, rerun that one mutation against the **whole** frontend unit suite (`cd frontend && npx rstest run`);
a survivor is a GAP only if the whole suite stays green. Record that second command and result in the row.
