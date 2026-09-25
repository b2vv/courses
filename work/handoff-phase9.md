# Хендофф — Phase 9 (готовність до оплати), 2026-09-25

Записано перед перезавантаженням ПК. Стан репо на мить коміту цього файла: HEAD — коміт, що його
додає (батько `5c9b018`). Перш ніж діяти, звір із `git status` і `git log` (§24 п.17 бріфу).

## Що вже зроблено

- **Phase 8 (ASR) закрита 22/22**, `fc5adf5`: `docs/asr/report.md`.
- **Рішення:** оплата — остання фаза, перед нею фаза готовності (`work/PRODUCT-BRIEF.md` §24 п.21);
  правило доступу — варіант B, одна чиста функція в `internal/contracts/access` (ADR 013).
  `AGENTS.md` оновлено.
- **Брифінг** (`work/CTO-RESEARCH.md`, `work/AGENDA.md`) перезібрано на `fc5adf5`.
- **Phase 9:** `docs/readiness/spec.md` (R01–R12), `docs/readiness/plan.md`, задачі в
  `work/tasks.md` (P01–P06).
  - P01 `2780207` — контракт `access`, межі, unit-таблиця. Done.
  - P02 `e51b66b` — три модулі делегують на контракт, AST-тест, захист від порожнього суб'єкта;
    повний `make check` зелений. Done.
  - P03 `5c9b018` — одна таблиця `accesstest` через SQL чотирьох адаптерів (R06–R08). Done.

## Що в дереві НЕЗАКОМІЧЕНЕ і НЕ ПЕРЕВІРЕНЕ

Codex закінчив обидві задачі; ні я, ні рецензент їх ще не перевірили.

- **P04** (R09, чотири точки — один сценарій відмови): зміни лише в тестах
  `backend/internal/{course,playback,transcript}/adapters/http/*handler_test.go`. Звіт:
  `work/report-p04.md`, Codex — `go vet` і `go test ./internal/...` PASS, Discussion: немає.
- **P05** (R11, «немає доступу» на маршруті навчання): **підозра підтвердилась** — відмова
  `course_forbidden` показувалась як збій перевірки сесії. Змінено `frontend/src/routes/learn.$.tsx`,
  `presentation/identity/ProtectedRouteStates.tsx`, локалі, `routes/route-guards.test.tsx`; новий e2e
  `frontend/e2e/learn-access.spec.ts`. Звіт `work/report-p05.md` з ланцюжком file:line.
  `npm test` у пісочниці Codex не стартував — **не запущено**.
- Два PNG у `docs/foundation/screenshots/` — побічний продукт гейта (B-18), **не комітити**.

## Що робити далі, по порядку

1. P04: `cd backend && go test ./internal/...`; диф → Courses 2 на рев'ю; коміт лише трьох
   тестів і `work/brief-p04.md`/`report-p04.md`.
2. P05: `cd frontend && npm test`, format/lint/typecheck; **перезібрати фронтенд** (у стеку
   продакшн-білд) — `docker compose build frontend && docker compose up -d frontend`; прогнати
   `npx playwright test e2e/learn-access.spec.ts`; диф → Courses 2; коміт.
3. P06: R12 (grep, що в коді немає нічого з §4 спеки); рядок в architecture-тесті — пакет
   `accesstest` не імпортується продакшн-кодом (нотатка Courses 2); повний `make check`; звіт
   `docs/readiness/report.md` з R01–R12 по одному; `work/tasks.md`.
   У §7 звіту: це вже третій раз, коли правило «хто додає глобально лічене — звужує вибірку»
   ловить реальну колізію (7a `role=status`, C19 `role=status`, метрика `playback` у P03).

## Як працювати (ролі й пастки)

- **Ролі** (рішення користувача 2026-09-24): Courses 1 — план, брифи, код через Codex, гейти,
  коміти, звіти. Courses 2 — лише рев'ю і звіт. Рев'ю дифу — перед комітом.
- **Codex:** `printf '…бриф…\n' | codex exec -s workspace-write --skip-git-repo-check -` (stdin
  закрито — без цього висне на «Reading additional input»). Правка після рев'ю — `codex exec …
  resume <ID> - < file`; ID шукати так: `ls -t ~/.codex/sessions/2026/09/25/*.jsonl` + grep назви
  брифу. Без бектиків у промптах-аргументах. У пісочниці Codex немає Docker, не стартують `npm test`
  і `httpserver`-тест (bind) — їх ганяє Courses 1.
- **Гейт:** `make check` на повному дереві; після змін спільного стану — завжди повний e2e.
- Контейнери Cassiopeia/cargotrans зупинені на прохання користувача — **не піднімати**.
