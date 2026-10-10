# Repository standards manifest

Describes the repository as it is, not as it was at bootstrap. Where this file and the
code disagree, the code wins and this file is corrected.

## Мірник
| Док | Що вирішує | Чекліст | Діє на |
|---|---|---|---|
| AGENTS.md | Architecture, scope, TypeScript rules, verification | all sections | entire repository |
| docs/architecture/frontend.md | Frontend dependency direction | dependency rules | frontend/** |
| docs/architecture/backend.md | Backend dependency direction | dependency rules | backend/** |
| docs/adr/*.md | Decisions already taken; a change that contradicts one needs a new ADR | the specific decision | as scoped by each ADR |

## Не мірник
`docs/<phase>/` (foundation, identity), `docs/security/`, `docs/reference/` and `work/` are
decisions, plans, evidence, briefings and tracked work. They record; they do not rule.

## Машинні гейти
Make targets in order: `format-check` → `lint` → `typecheck` → `test` → `integration` →
`check-generated` → `build` → `up` → `seed-identities` → `seed` → `smoke` → `e2e`.
CI runs the same order in `.github/workflows/ci.yml`. `lint` includes the frontend
architecture check; `check-generated` fails on sqlc/Orval drift.
Measured results belong in the current phase report (`docs/<phase>/report.md`).
Unexecuted gates are not proof, and a gate that was skipped is reported as skipped.

## Конституція
AGENTS.md architecture and TypeScript rules apply. User scope and explicit authorization
override skill ceremony.

## Поріг пайплайна
`spec-flow` is enabled by a new subsystem, a public API change, or a cross-cutting change
that touches more than one layer of both backend and frontend. Anything smaller uses an
inline plan and targeted verification. Say out loud which side of the threshold a piece of
work falls on before starting it.

## Розміщення
Backend modules: `backend/internal/<module>/{domain,application,ports,adapters}`.
Backend shared infrastructure: `backend/platform/<concern>/`.
Frontend: `frontend/src/{domain,application,ports,adapters,presentation,app,routes}`.

Reuse index — check these before writing anything new:

| Concern | Lives in |
|---|---|
| HTTP call to our own API (deadline, cancellation, failure classification) | `frontend/src/adapters/api/transport.ts` |
| Per-request headers and SSR cookie forwarding | `frontend/src/adapters/api/request-headers.ts` |
| UI primitives over React Aria | `frontend/src/presentation/ui/` |
| Query boundary and route-loader cancellation | `frontend/src/app/composition/catalog-queries.ts` |
| Request-scoped dependency wiring | `frontend/src/app/composition/` |
| HTTP server, error envelope, request id | `backend/platform/httpserver/` |
| DSN assembly from discrete credentials | `backend/platform/config/` |
| Dependency lifecycle (Postgres, Redis, tracing, metrics) | `backend/platform/runtime/` |
| Argon2id verification | `backend/internal/identity/adapters/argon2id/` |

Dependency direction is machine-checked: `backend/platform/architecture/boundaries_test.go`
and `frontend/scripts/check-architecture.mjs`.

## Гілки й PR
`main` is the trunk and carries only verified work. Delivery happens on a feature branch
(`feat/<phase>-<topic>`) pushed to `origin` (`git@github.com:b2vv/courses.git`, private).
Merge into `main` fast-forward when history is linear. Never force-push over a ref
carrying commits this session did not create without explicit instruction.

## Артефакти
Specification, plan, gates, evidence and report of a phase: `docs/<phase>/`
(`docs/foundation/`, `docs/identity/`, and one per later phase).
Repository briefing: `work/CTO-RESEARCH.md`. Agenda and next move: `work/AGENDA.md`.
Security audit disposition: `docs/security/`.
Tasks: `work/tasks.md`, one row per task with an explicit status
(`planned` → `in_progress` → `done` | `blocked`).
Archive: a completed row **stays in place** and its "Result / archive pointer" column
becomes the pointer — commit, report section, or evidence file. Rows are never deleted,
because a deleted task takes the trace of the decision with it.
Coordination ledger: `work/archive/ledger.md`, append-only by meaning: a correction is a new entry
referencing the earlier one.
Living notes: `docs/<phase>/assumptions.md`, or `docs/<phase>/_wip-<topic>.md` while work is
in flight.
