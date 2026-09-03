# Repository standards manifest

## Мірник
| Док | Що вирішує | Чекліст | Діє на |
|---|---|---|---|
| AGENTS.md | Architecture, scope, verification | all sections | entire repository |
| docs/architecture/frontend.md | Frontend dependency direction | dependency rules | frontend/** |
| docs/architecture/backend.md | Backend dependency direction | dependency rules | backend/** |

## Не мірник
`docs/foundation/`, `work/` are decisions, plans, evidence and tracked work.

## Машинні гейти
Make targets: format-check → lint → typecheck → test → integration → build → e2e.
Exact commands and measured results are recorded in docs/foundation/report.md; unexecuted gates are not proof.

## Конституція
AGENTS.md architecture and TypeScript rules apply. User scope and explicit authorization override skill ceremony.

## Поріг пайплайна
A new subsystem or public API change enables spec-flow; this initial monorepo meets both conditions.
For smaller changes use an inline plan and targeted verification.

## Розміщення
Backend modules: backend/internal/<module>/{domain,application,ports,adapters}.
Frontend: frontend/src/{domain,application,ports,adapters,presentation,app,routes}.
No existing code to reuse at bootstrap.

## Гілки й PR
New local repository; main is the initial branch. No remote or publishing configured.

## Артефакти
Specification, plan, assumptions, gates and reports: docs/foundation/.
Tasks: work/tasks.md with explicit pending/in-progress/complete status.
Archive: work/archive/ containing pointers to completed results, not copies.
Coordination ledger: work/ledger.md. Living notes: docs/foundation/assumptions.md.
