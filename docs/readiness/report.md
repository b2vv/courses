# Phase 9 — готовність до оплати: звіт

| | |
|---|---|
| Фаза | 9 — готовність до оплати |
| Дата закриття | 2026-09-28 |
| Автор | сесія Courses (кодер); рецензент — Courses 2 |
| Підстава | `docs/readiness/spec.md`, `docs/readiness/plan.md`; ADR 013 (рішення «B», 2026-09-25) |
| Роль фази | Передостання фаза роадмапу. Робить доступ **одним правилом під гейтом** і документує решту механізмів оплати як шви, не будуючи для них коду без споживача. Оплата — наступна й остання фаза (`work/PRODUCT-BRIEF.md` §24 п.21). |

## 1. Що зроблено (P01–P06)

- **P01** `2780207` — правило доступу стало одним контрактом `internal/contracts/access`
  (чиста функція `CanView`), межі контрактів зафіксовано, unit-таблиця на 9 випадків.
- **P02** `e51b66b` — три модулі (`course`, `playback`, `transcript`) делегують на контракт;
  AST-тест вимагає рівно `return access.CanView(…)`; порожній суб'єкт нічим не володіє;
  повний `make check` зелений.
- **P03** `5c9b018` — одна таблиця `accesstest` проходить через SQL чотирьох адаптерів (R06–R08).
- **P04** `090b912` — R09: чужа чернетка відмовляється, опублікований дозволяється, у чотирьох
  точках перевірки, кожна своїм контрактом.
- **P05** `17a5d85` — R11: студент на чужій чернетці бачить названий стан «немає доступу»
  (без retry, з посиланням у каталог), а не збій перевірки сесії.
- **P06** (цей коміт) — architecture-рядок «фікстура не в проді», звірка R12, упорядкований
  `make check`, цей звіт.

## 2. R01–R12 — по одному рядку, кожен своїм доказом

Зелений `make check` — не доказ рядка; нижче кожен рядок вказує на власний тест чи звірку.

| Рядок | Доказ (файл) | Стан |
|---|---|---|
| **R01** — `access.CanView` правильний на всіх 9 випадках | `backend/internal/contracts/access/access_test.go:10` `TestCanView` над `accesstest.TruthTable` (9 рядків) | ✔ P01 |
| **R02** — `contracts/…` не імпортують модуль чи не-stdlib | `backend/platform/architecture/contracts_test.go:13` `TestContractsImportOnlyTheStandardLibrary` | ✔ |
| **R03** — домен може імпортувати лише `contracts/…` + stdlib | `backend/platform/architecture/boundaries_test.go:13` `TestDomainAndApplicationDependenciesPointInward` | ✔ |
| **R04** — одна реалізація: тіло кожного `CanBeViewedBy` = рівно `return access.CanView(…)` | `backend/platform/architecture/access_mapping_ast_test.go` `TestEveryCanBeViewedByDelegatesToAccessCanView` (рівно 3 методи) | ✔ P02 |
| **R05** — мапінг кожного модуля дає вердикт `CanView` | course/playback: `learning_test.go:104`, `authorization_test.go:16` ганяють `CanBeViewedBy` над `coursevisibility.Cases()` (окрема 9-кейсова таблиця — див. §7); transcript-мапінг доводиться структурно (R04, AST: тіло = рівно `return access.CanView`) + інтеграцією за `lessonID` (R08) + R01, а не домен-табличним тестом (`transcript_test.go:182` — лише одиничний assert) | ✔ P02 |
| **R06** — факти адаптера `course` дають вердикт таблиці | `backend/internal/course/adapters/postgres/learning_integration_test.go` (`accesstest` через SQL за slug) | ✔ P03 |
| **R07** — те саме для `playback` (`auth_request` і сесії) | `backend/internal/playback/adapters/postgres/access_integration_test.go` (за `videoID`) | ✔ P03 |
| **R08** — те саме для `transcript` | `backend/internal/transcript/adapters/postgres/repository_integration_test.go` (за `lessonID`) | ✔ P03 |
| **R09** — чотири точки відмовляють на чужій чернетці, кожна своїм контрактом | `TestR09…` у `course/adapters/http/learning_handler_test.go`, `transcript/adapters/http/handler_test.go`, `playback/adapters/http/handler_test.go` (session + authorize) | ✔ P04 |
| **R10** — архівний курс: сторонньому ні, власнику й адміну так | рядки `administrator/owner/stranger archived` у `accesstest/table.go:25,28,31` (Want true/true/false), проходять як unit (R01) і integration (R06–R08) | ✔ |
| **R11** — маршрут навчання показує названий «немає доступу» на відмову бекенду | `frontend/src/routes/route-guards.test.tsx:575` `R11` (unit) + `frontend/e2e/learn-access.spec.ts` (e2e, прод-шлях) | ✔ P05 |
| **R12** — шви §5 записані; ADR 013 прийнятий; жодного порту/таблиці/типу з §4 у коді | grep за переліком §4 (нижче §4) + `backend/platform/architecture/access_fixture_isolation_test.go` (фікстура не в проді) | ✔ P06 |

## 3. Architecture-рядок P06 — фікстура не на прод-шляху

`accesstest` — тестова фікстура (спільна таблиця істинності). Прод-шлях не має тягнути її:
інакше тестовий дубль підмінив би справжнє правило. `TestAccessFixtureIsNotImportedByProductionCode`
обходить кожен не-тестовий `.go` модуля й відмовляє імпорт `courses/internal/contracts/access/accesstest`.

Тест **звужує вибірку** (предикат «чи цей прод-файл імпортує фікстуру?») замість асертити ім'я на
позиції — за правилом AGENTS.md «хто читає глобально лічене — звужує вибірку». Це вже **четвертий**
раз, коли це правило ловить реальну колізію (7a `role=status`; C19 `role=status`; метрика `playback`
у P03; тут — щоб рядок не був вакуумним).

Не вакуумність доведена негативним контролем: детектор, націлений на `…/access` (який імпортують
3 прод-файли з `CanBeViewedBy`), спіймав рівно 3 файли; після повернення на `accesstest` — PASS.

## 4. R12 — grep за переліком §4 (чого немає в коді)

Прод-код (`backend/internal`, `backend/cmd`, `frontend/src`; без тестів, e2e, generated):

| Термін §4 | Збігів | Нотатка |
|---|---|---|
| payment provider / port, order (таблиці), enrollment, price, currency, money, webhook, refund, payout, access audit, invoice, checkout | **0** | немає жодного споживача — саме як вимагає §4 і ADR 013 |

Хибнопозитивні, перевірені руками: `order` — сортування документів транскрипту; `currency` —
збіг у «Con**curren**cy»; єдина згадка `payment` — коментар у `access.go:19` «правило стане
несучим для оплати», тобто задокументований шов §5, а не код оплати. Шви §5 лишаються документом,
не кодом; ADR 013 прийнятий.

## 5. Перевірка

Упорядкований `make check` на повному дереві (`format-check lint typecheck test integration
check-generated build up drain-queue seed-identities seed smoke e2e`):

- **PASS, exit 0** (2026-09-28). Go: 0 FAIL по всьому прогону (unit + integration на ізольованому Postgres); фронт-unit: 483 passed, 0 failed; браузерний e2e: **75/75 passed** (11.0 хв, включно з новим `learn-access`). `check-generated`, `build`, `smoke` — усі зелені.

Проміжні прогони кодером до гейта: gofmt/vet/`go test` пакетів доступу; `-run R09 -v` — 4 тести
RUN+PASS; фронт format/typecheck/lint; `npm test route-guards` — R11+G09 PASS; прод-ребілд фронту
+ e2e `learn-access` + ширший набір (playback happy-path, route-guards, admin-users) — 71 passed;
`TestAccessFixtureIsNotImportedByProductionCode` PASS + негативний контроль.

## 6. Рев'ю

P04 і P05 незалежно перевірені рецензентом (Courses 2) на прод-шляху, не з довіри до звітів,
до коміту. Два не блокуючі мінори занотовані як follow-up (§7). P06 (architecture-рядок + звіт)
надіслано на рев'ю перед комітом.

## 7. Обмеження й follow-up

- **Дві паралельні 9-кейсові таблиці (pre-existing, не P06 — чесність звіту):** домен-юніти
  `course`/`playback` ганяють `internal/testsupport/coursevisibility/cases.go` `Cases()`, а не
  `accesstest.TruthTable`. Дрейф між ними ловиться, бо обидві асертять проти того самого
  `access.CanView` (R04 гарантує рівно одну реалізацію правила), — але наратив «одна таблиця
  `accesstest`» цього не покриває. `transcript`-домен табличного тесту `CanBeViewedBy` не має
  взагалі (лише одиничний assert `transcript_test.go:182`); його мапінг тримають R04+R08+R01.
  Кандидат на прибирання дубля — окрема задача, не блокує закриття Phase 9.
- **Не зроблено (§4, свідомо):** порт платіжного провайдера, таблиці замовлень, enrollment,
  права на курс, ціни, валюти, гроші як тип, вебхук-inbox, повернення, аудит доступу, виплати,
  новий механізм сесій. Причина — жоден не має споживача; це шви §5, будуються у фазі оплати.
- **Мінор P04 (не блокує):** transcript-тест R09 пінить лише статус 403, тоді як learning/playback
  пінять і `"code"`. Під ADR 010 (403 невідрізнюване від «немає транскрипту») це коректно, але
  слабше; за бажання додати `"code":"course_forbidden"` для симетрії.
- **Мінор P05 (не блокує):** якщо в e2e POST створення пройшов, а PUT з уроком упав до `course =`,
  run-scoped чернетка лишиться (afterAll видаляє лише коли `course` заданий). Слаг run-scoped +
  ефемерна CI-БД → вплив мізерний; за бажання ставити `course` і з POST-відповіді.

## 8. Наступне

Фаза оплати — остання (`work/PRODUCT-BRIEF.md` §24 п.21): адаптер платіжного провайдера й
вебхук-inbox, що спираються на це єдине правило доступу як несучу умову «чи має користувач право
на цей курс». Шви — `docs/readiness/spec.md` §5.
