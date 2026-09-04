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
| I03 | HTTP middleware, cookies and typed runtime config | 2h | I02 | done | `docs/identity/report.md` I03; full gate sequence + live curl evidence; microsecond-clock defect fixed |
| I04 | OpenAPI/Orval, Compose and Nginx auth edge | 1.5h | I03 | done | `docs/identity/report.md` I04; 12-case OpenAPI contract test; audit MEDIUM-1 closed with the `web` zone and a strict `login` zone |
| I05 | Frontend session adapter and localized login | 2h | I04 | done | `docs/identity/report.md` I05; 112 frontend tests; A03/A06 covered; login redirect defect fixed |
| I06 | Client route guards and completion gate | ~12h, cut into I06a–I06h below | I05 | planned | `docs/identity/i06/plan.md` 3rd edition; two GATE 2 refusals recorded in `docs/identity/_wip-i06.md` |

## Pre-deploy hardening

| ID | Назва | Estimate | Depends on | Status | Result / archive pointer |
|---|---|---:|---|---|---|
| S01 | CSP (report-only first), Permissions-Policy, HSTS and `Secure` cookies after TLS | 1.5h | TLS termination | planned | audit LOW-5, LOW-6 |
| S02 | `__Host-` prefix on the session cookie | 0.5h | S01 (needs `Secure`) | planned | Removes the precondition for the duplicate-cookie class entirely: a `__Host-` cookie cannot be set by a sibling or parent domain. Blocked until TLS, because the prefix requires `Secure` and local development runs on http. See `docs/identity/report.md` §Security review finding |

### I06 розрізано на задачі

Оцінка 1.5h була неправдивою; реальна ≈12h, і план фази обмежує задачу двома годинами.
Домінує не сама фіча, а те, що у `frontend/src` **немає жодного маршрутного тесту**, немає
перевірки SSR-виводу й немає інтеграційного тесту сіду — а таблиця вимагає всіх трьох.

| ID | Назва | Estimate | Depends on | Status | Result / archive pointer |
|---|---|---:|---|---|---|
| I06a | Кореневий `beforeLoad`: резолв ідентичності, перевірка сесійної куки за іменем, дискримінований результат без кидків; `login.tsx` знімає власний лоадер | 2h | GATE 2 green | planned | закриває G07, G10, G15, G22a |
| I06b | Layout-роути `/learn` і `/admin` із гардами, `errorComponent` і `pendingComponent`; видалення `LockedArea` і ключів `locked.*` | 2h | I06a | planned | закриває G01–G06, G08, G09, G11, G39 |
| I06c | Звуження `safeReturnTo` під `spec.md:52` плюс збереження `search` у `returnTo` | 1h | — | planned | закриває G15a, G15b, G16, G17, G18 |
| I06d | `SessionMenu`, вихід, посилання на адмінку за роллю, uk/en, доступність | 2h | I06a | planned | закриває G19, G20, G24–G28, G29–G31 |
| I06e | Екрани `/learn` і `/admin` на `EmptyState`, uk/en | 0.5h | I06b | planned | закриває G03, G40 |
| I06f | Фікстури: три ідентичності, `ON CONFLICT DO UPDATE`, гейт через psql-змінну, `ci.yml`, README | 1.5h | — | planned | закриває G32–G38 |
| I06g | Тестові поверхні, яких немає: маршрутні тести, сире SSR-HTML через `page.request`, лічильник звернень, паралельний SSR, інтеграційний тест сіду | 2h | I06a–I06f | planned | несуче для більшості рядків вище |
| I06h | `Vary` через append, ціль `check` у порядку маніфесту, переписаний `route-guards.spec.ts`, фінальний впорядкований гейт | 1h | усі | planned | закриває G21, G21a, G41 |

