# Foundation task tracker

Джерело scope: [spec](../docs/foundation/spec.md). Детальні файли, gates і acceptance links: [plan](../docs/foundation/plan.md). Стани: `planned` → `in_progress` → `done` або `blocked`. При завершенні рядок зберігається як archive pointer з посиланням на результат/commit і evidence в ledger; задачі не видаляються.

| ID | Назва | Estimate | Depends on | Status | Result / archive pointer |
|---|---|---|---|---|---|
| F01 | Repository conventions, architecture, spec/plan gates | 1h | — | done | docs/foundation/gates.md — 4/4 critiques, defense green |
| F02 | Go API/worker runtime and health | 1.5h | F01 | done | commit `5608a03`; backend-runtime-evidence.md |
| F03 | TanStack Start/Rsbuild SSR scaffold | 2h | F01 | done | commit `5608a03`; Phase 1 report |
| F04 | React Aria UI kit and design tokens | 2h | F03 | done | commit `5608a03`; ui-evidence.md |
| F05 | Compose, Nginx, CI and infrastructure smoke | 1.5h | F02,F03 | done | commit `5608a03`; smoke PASS |
| F06 | Phase 1 working-result gate | 0.5h | F04,F05 | done | commit `5608a03`; ordered gates PASS |
| C01 | Course domain, use cases, migrations, PostgreSQL | 2h | F06 | done | commit `31e8c39`; course-backend-evidence.md |
| C02 | HTTP API and OpenAPI contract | 1.5h | C01 | done | commit `31e8c39`; 26-case HTTP matrix |
| C03 | Orval client, frontend domain/application/adapters | 1h | C02 | done | commit `4d73072`; frontend-test-evidence.md |
| C04 | SSR catalog/course and states | 2h | C03,F04 | done | commit `4d73072`; 53 Rstest + screenshots |
| C05 | Complete integration, browser and build verification | 1.5h | C04 | done | 14 Playwright + final CI gates; report.md |

Phases 1–2 завершено. Phase 3 розпочато зі специфікації Identity; Phases 4–8 не починалися.

## Phase 3 — Identity

| ID | Назва | Estimate | Depends on | Status | Result / archive pointer |
|---|---|---:|---|---|---|
| I01 | Identity users/sessions migration and domain | 1.5h | green Identity plan gates | done | `docs/identity/report.md`; migration/domain tests pass |
| I02 | Login/session use cases and adapters | 2h | I01 | done | `docs/identity/report.md`; application/Argon2id/PostgreSQL tests pass |
| I03 | HTTP middleware, cookies and typed runtime config | 2h | I02 | planned | — |
| I04 | OpenAPI/Orval, Compose and Nginx auth edge | 1h | I03 | planned | — |
| I05 | Frontend session adapter and localized login | 2h | I04 | planned | — |
| I06 | Client route guards and completion gate | 1.5h | I05 | planned | — |
