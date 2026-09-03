# Foundation task tracker

Джерело scope: [spec](../docs/foundation/spec.md). Детальні файли, gates і acceptance links: [plan](../docs/foundation/plan.md). Стани: `planned` → `in_progress` → `done` або `blocked`. При завершенні рядок зберігається як archive pointer з посиланням на результат/commit і evidence в ledger; задачі не видаляються.

| ID | Назва | Estimate | Depends on | Status | Result / archive pointer |
|---|---|---|---|---|---|
| F01 | Repository conventions, architecture, spec/plan gates | 1h | — | done | docs/foundation/gates.md — 4/4 critiques, defense green |
| F02 | Go API/worker runtime and health | 1.5h | F01 | in_progress | — |
| F03 | TanStack Start/Rsbuild SSR scaffold | 2h | F01 | in_progress | — |
| F04 | React Aria UI kit and design tokens | 2h | F03 | planned | — |
| F05 | Compose, Nginx, CI and infrastructure smoke | 1.5h | F02,F03 | planned | — |
| F06 | Phase 1 working-result gate | 0.5h | F04,F05 | planned | — |
| C01 | Course domain, use cases, migrations, PostgreSQL | 2h | F06 | planned | — |
| C02 | HTTP API and OpenAPI contract | 1.5h | C01 | planned | — |
| C03 | Orval client, frontend domain/application/adapters | 1h | C02 | planned | — |
| C04 | SSR catalog/course and states | 2h | C03,F04 | planned | — |
| C05 | Complete integration, browser and build verification | 1.5h | C04 | planned | — |

Phase 3–8 не починаються в цьому циклі. Наступна найменша задача — незалежні review gates F01, потім F02. Рішення користувача наразі не потрібне.
