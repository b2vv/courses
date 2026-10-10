# План foundation та першого Course slice

Статус: написаний до коду; незалежна критика виконана, blocking виправлено; незалежний plan-defense зелений, 0 дірок — [gate record](gates.md). Цей файл описує намір, а не звіт про пройдені перевірки. Поточні стани та докази — [tasks](../../work/tasks.md), [ledger](../../work/archive/ledger.md).

## Межі та порядок

Спочатку Phase 1 дає запуск інфраструктури, API/worker health і SSR shell, потім Phase 2 додає один наскрізний Course slice. Перехід у Phase 2 дозволений лише після working-result gate F06. Phase 3–8 не починаються. Оцінки 0.5–2 години задають розмір задач, а не обіцяний календарний час. Один логічний slice виконується за раз; `[P]` означає лише незалежність залежностей, не дозвіл обходити фазовий gate.

## Реалізація та dependency direction

- Monorepo має `frontend/`, `backend/`, `infrastructure/`, `docs/`, root Compose/Make/env/README/AGENTS; generated files окремо від handwritten code.
- Backend — modular monolith, `net/http`, pgx/sqlc/Goose. `internal/course/domain` залежить лише від stdlib; `application` — від domain і ports; HTTP/Postgres adapters не повертають transport/DB models у domain. Composition — `cmd/api` та platform bootstrap, без service locator.
- Frontend domain — чистий TypeScript без React/framework/generated DTO. Application залежить лише від domain і ports. Orval native-fetch client викликає API adapter, який валідовує/мапить DTO; React routes/presentation викликають use cases через composition і Query options. Прямий fetch у presentation заборонено.
- TanStack Query володіє server state. Router search params володіють q/sort/limit/offset. Кожен SSR request отримує власні QueryClient і session-aware API dependency instances; forwarded cookies ніколи не зберігаються в глобальному mutable об'єкті. Query hydration включає тільки допустимі дані поточного request. Немає shared кешування помилок чи приватних сесій.
- Публічні `/`, `/courses`, `/courses/$slug`, `/login` рендеряться на сервері. `/learn/*`, `/admin/*` — client placeholders з fail-closed поведінкою; account — зарезервований маршрут без fake identity.
- Власний UI kit над React Aria має tokens light/dark, semantic controls, visible focus, мінімальну target-size 44 CSS px і layout, що витримує 200% text zoom. Критичні Button/Link/TextField/Dialog/Select перевіряються клавіатурою; решта мають compile/smoke coverage без дублювання тестів бібліотеки.
- У першому UI kit: Button, Link, TextField, TextArea, Select, ComboBox, Checkbox, Switch, Dialog, Menu, Tabs, Tooltip, ProgressBar, Table, Pagination, FileUpload, Toast, EmptyState, ErrorState, Skeleton. Компоненти зі штатною семантикою не отримують фальшивих ARIA roles.
- Nginx: `/` → SSR, `/api/*` → Go, `/media/*` заборонений до playback authorization. Відеобайти не йдуть через Go/SSR. Health, readiness і metrics не розкривають секретів; metrics доступні лише внутрішній мережі.

## API та дані

`GET /api/v1/courses?limit=20&offset=0&q=&sort=recent` повертає `{items, pagination: {limit, offset, total}}`; `GET /api/v1/courses/{slug}` — один published курс. Slug: lowercase ASCII слова/цифри з одиночними внутрішніми дефісами, максимум 120 символів. Застосунок має явні invalid_query, invalid_slug, course_not_found та dependency_unavailable помилки, HTTP adapter мапить їх у 400/404/503 envelope `{error:{code,message,requestId}}`; unexpected errors дають безпечний 500. Request ID повертається заголовком і в error body.

Вхідні межі: limit 1–50, offset 0–10000, q ≤100 Unicode code points після trim, sort `recent`/`title`; неповні, повторені або нечислові числові query parameters відхиляються. Пошук literal substring, `%` і `_` не стають SQL wildcard. Лише `published`; `recent` сортує published_at DESC/id ASC, `title` title ASC/id ASC з явно визначеною collation. List/count використовують read-only repeatable-read transaction. Read context обмежено timeout і cancellation; без нескінченних retries. Request abort доходить до Go/pgx. SSR dependency failure не перетворюється на empty catalog.

OpenAPI 3.0.3 живе в backend і перевіряється проти handlers. Orval generation — відтворювана, exact dependencies/lockfile, generated diff-check. Майбутній `Idempotency-Key` описаний як reserved mutation contract; read-only endpoint його не потребує. Demo seed окремо і безпечний при повторі.

## Phase 1 shell gate та відновлення

Phase 1 has no Course migration: Goose up with an empty migration directory is a successful no-op. API/worker depend on migrate service_completed_successfully; failed migration fails make up, never silently starts application. Volumes remain retained and a corrected migration can be retried. No application-ready claim after migration failure.

P01: SSR shell title exists in initial HTML with JS disabled; shell hydrates without console errors. P02: shell skip-link/Button/TextField/Dialog/Select keyboard and focus work. P03: all runtime services healthy, media403, internal metrics, rate limit429, shutdown and restart succeed with named volumes retained. F02–F06 verify P01–P03; H02/H05/H07/E10 with actual course data only apply after C01–C04. F06 runs ordered format/lint/typecheck/unit/build plus Phase1 Playwright shell checks. This removes any dependency on future Course implementation.

C04 recovery: ErrorState has an accessible reload/retry link preserving the entire URL; no-JS SSR supports ordinary document reload. E11: API initially fails, then recovers; activating retry displays matching results at the same q/sort/limit/offset URL.

Query keys include all normalized q/sort/limit/offset values. Route renders only its active key, obsolete request signals are aborted, and C03/C04 test reversed A/B responses so a late A result cannot overwrite B screen or URL.

Pagination resets offset=0 when q/sort/limit changes. Next is disabled if it would exceed offset10000 or total. An out-of-results valid URL retains total and offers first-page navigation. Seed uses one transaction, stable UUIDs, ON CONFLICT DO NOTHING; parallel/repeated execution has no duplicates or partial commits. Integration tests exercise concurrent seed and failed migration rollback on disposable databases only.

## Приймальні сценарії

Шляхи тестів нижче — заплановані, до імплементації не вважаються існуючими. Немає постійно ручних acceptance-рядків; огляд документації залишається окремим review gate.

| ID | Група / сценарій | Передумова | Дія користувача | Очікуваний спостережуваний результат | Перевірка |
|---|---|---|---|---|---|
| H01 | Happy: чистий запуск | Docker, порожні volumes, `.env` за прикладом | `make up` | Nginx, SSR, API, worker, PostgreSQL і Redis досягають healthy/running стану | `infrastructure/tests/smoke.sh` |
| H02 | Happy: initial HTML каталогу | Явно seeded published курс | HTTP GET `/courses` без JS | Назва курсу присутня у початковому HTML | `frontend/e2e/catalog.spec.ts` request/no-JS test |
| H03 | Happy: сторінка курсу | Відомий published slug | Відкрити курс з каталогу | Сторінка показує назву й опис курсу | `frontend/e2e/catalog.spec.ts` |
| H04 | Happy: URL стан | Щонайменше 3 published курси | Відкрити q/sort/limit/offset URL повторно | Відображено той самий відфільтрований порядок | API contract + `frontend/e2e/catalog.spec.ts` |
| H05 | Happy: доступність | Каталог відкрито | Tab, Enter і skip link | Фокус видимий, курс відкривається клавіатурою | `frontend/e2e/accessibility.spec.ts` |
| H06 | Happy: health | API/worker працюють | GET відповідного health/status | Повернуто documented 200 status без приватних даних | `backend/platform/httpserver/server_test.go`, smoke |
| H07 | Happy: SSR hydration | Public API доступний | Завантажити каталог з JS | У браузері немає hydration error | `frontend/e2e/catalog.spec.ts` console assertions |
| B01 | Boundary: порожній каталог | Немає published courses | Відкрити каталог | Видно EmptyState з 0 курсів | component + API repository tests |
| B02 | Boundary: один і max limit | 51 published courses | Запросити limit=1 та limit=50 | Відповідь містить відповідно 1 та 50 items | `backend/internal/course/adapters/http/handler_test.go`, repository integration |
| B03 | Boundary: крайні offset/q | offset=10000 або q із 100 code points | Відкрити валідний URL | Запит прийнято й повернуто documented pagination | HTTP contract + application tests |
| B04 | Boundary: beyond end | total менше offset | Відкрити наступну сторінку | Items порожні, total збережено | repository integration |
| B05 | Boundary: стабільні ties | Однакові title/published_at | Повторити той самий запит | Порядок id однаковий | repository integration |
| B06 | Boundary: приховані курси | draft та archived courses існують | GET catalog | Жоден прихований курс не присутній | repository integration |
| B07 | Boundary: прихований detail | draft/archived slug існує | GET details | 404 і той самий envelope, що для невідомого slug | HTTP contract + repository integration |
| B08 | Boundary: text zoom/themes | Catalog/UI kit відкриті | Увімкнути dark/light та 200% text zoom | Вміст/controls доступні без перекриття або втрати дій | `frontend/e2e/accessibility.spec.ts` |
| B09 | Boundary: повторний seed | Demo seed виконано | Виконати seed вдруге | Кількість demo courses не збільшується | `infrastructure/tests/smoke.sh` |
| E01 | Failure: malformed input | Доступний API | invalid limit/offset/sort/q/slug | 400 standardized error з requestId | table-driven HTTP contract tests |
| E02 | Failure: відсутній курс | Валідний невідомий slug | Відкрити course URL | Видно not-found page зі статусом 404 | `frontend/e2e/catalog.spec.ts`, HTTP contract |
| E03 | Failure: PostgreSQL offline | API running, DB unavailable | GET catalog | API 503 dependency_unavailable без SQL/credentials | HTTP adapter fake-port test + smoke dependency check |
| E04 | Failure: SSR dependency | API недоступний | Відкрити каталог | Видно ErrorState; сервер не видає empty success | frontend adapter MSW + SSR failure test |
| E05 | Failure: slow request | Затриманий API | Відкрити каталог | Операція завершується timeout error в межах налаштованого budget | application/adapter timeout tests |
| E06 | Failure: abort | Slow DB/API request | Перервати запит/навігацію | Запит завершується без завислого connection | backend cancellation integration + frontend adapter abort test |
| E07 | Failure: media bypass | Відомий або довільний media path | GET `/media/x/master.m3u8` | Nginx повертає 403 без bytes зі storage | infrastructure smoke |
| E08 | Failure: unauthorized routes | Немає identity session | Відкрити `/admin` та `/learn` | Видно недоступність захищеної функції, приватні дані відсутні | `frontend/e2e/route-guards.spec.ts` |
| E09 | Failure: excessive traffic | Rate-limit burst вичерпано | Повторити запити через Nginx | HTTP 429 без secret logging | infrastructure rate-limit smoke |
| E10 | Failure: shutdown/restart | Сервіси running з demo data | Stop, потім `make up` | Система знову healthy, demo data збережено | infrastructure smoke, volumes retained |

## Малі задачі та залежності

| ID | Phase / оцінка | Залежності | Файли / результат | Мінімальний gate | Acceptance |
|---|---|---|---|---|---|
| F01 | Repository/docs, 1h | — | Root AGENTS/README/Make/env/gitignore, `docs/architecture/*`, `docs/adr/001–006*`, `docs/development/*`, manifest і цей план | Огляд scope/структури/посилань; незалежні spec/plan gates | H01; усі контрактні припущення |
| F02 | Go runtime, 1.5h [P] | F01 | `backend/go.mod/go.sum`, `cmd/api`, `cmd/media-worker`, platform httpserver/observability/postgres/redis, Dockerfile | `gofmt`, `go vet ./...`, `go test ./...`, binaries build; health smoke | P03 |
| F03 | SSR scaffold, 2h [P] | F01 | Frontend npm lock/configs, app composition/router/providers, routes, SSR entry/server, Dockerfile | npm install/ci, `tsc --noEmit`, Rstest smoke, production client/server build, `/health` | P01 |
| F04 | React Aria UI kit, 2h | F03 | `presentation/ui/*`, `app/styles/*`, critical component tests | Oxfmt/Oxlint/typecheck; Rstest critical keyboard/focus tests | P02 |
| F05 | Compose/Nginx/CI, 1.5h | F02, F03 | Compose/dev override, Nginx config, optional monitoring config, `.github/workflows/ci.yml`, smoke script | `docker compose config`, build/up, proxy/health/media-deny checks | P03 |
| F06 | Phase 1 gate, 0.5h | F04, F05 | Phase 1 verification record, corrected setup/testing docs | Format → lint → typecheck → unit → runtime smoke → build; no fake green | P01–P03 |
| C01 | Course data/use cases, 2h | F06 | course domain/application/ports/postgres, migrations/sql/sqlc, explicit seed | table-driven domain/use-case tests; Testcontainers repo + up/down/up migration test; sqlc diff | B01–B07, B09, E05–E06 |
| C02 | HTTP/OpenAPI, 1.5h | C01 | HTTP handler/routes, `backend/openapi/openapi.yaml`, error/request middleware | HTTP contract cases, spec lint, Go checks | H04, E01–E03 |
| C03 | Generated client/adapters, 1h | C02 | Orval config/generated client; frontend domain/ports/application/API mapper/Query options | Orval generation + clean diff; Rstest with MSW; typecheck | H04, E04–E06 |
| C04 | SSR catalog/course, 2h | C03, F04 | Course routes + presentation pages/loading/empty/error, search state, SSR Query hydration | Rstest components/use cases, no-JS SSR requests, production build | H02–H05, H07, B01, B08, E02, E04, E08 |
| C05 | End-to-end completion, 1.5h | C04 | Playwright tests, full Compose seed/smoke, CI gates, final factual docs/report | Full ordered CI locally; browser against production compose; secret/dependency boundary review | All rows; no waived failed gates |

## Сайд-ефекти

Створюються файли нового репозиторію, lockfile, Docker images/networks/named volumes. Зовнішні повідомлення, платежі та deployment у публічний хостинг відсутні. `make up` застосовує лише versioned schema migrations через one-shot migrate service після healthy DB. Seed публікує тільки явні demo records після окремої команди. Контейнери мають health conditions, restart policy, log rotation і graceful stop; worker має CPU/memory limits. Nginx бачить media volume read-only; API/SSR не отримують цей volume.

## План міграції

Goose версіонує schema migration `courses`: UUID PK, unique canonical slug, title/summary/description/status, created/updated timestamps, published_at і database constraints/indexes. Міграція не містить sample data. sqlc queries відокремлені від domain. Testcontainers запускає PostgreSQL 17 із одноразовими даними, перевіряє up/down/up, constraints і adapter поведінку. `down` запускається тільки у спеціально створеній ephemeral test DB. Локальні/майбутні production volumes автоматично не відкочуються і не очищаються. Query list/count використовують спільний transaction snapshot.

## Rollback

При невдалому релізі зупинити залежні API/SSR/worker і повернути попередні image/config revisions, зберігши named volumes. Поточна initial migration не потребує автоматичного schema rollback; виправлення — forward migration після оцінки даних. `docker compose down` без `-v` дозволяє зберегти стан. Будь-який reset volume або destructive down поза ephemeral test DB є окремою явною дією користувача. Timeout/abort операцій читання не змінює бізнес-даних.

## План-defense — незалежний review завершено

| # | Питання | Закрито рядком | Вердикт |
|---|---|---|---|
| 1 | Кривий інпут | [API та дані](#api-та-дані), рядки 22–24: numeric bounds, sort enum, q code points, canonical slug, duplicate rejection та 400 envelope; E01, рядок 62 | Закрито незалежним review |
| 2 | Відмова залежності | [API та дані](#api-та-дані), рядки 22–24: 503, timeout/abort; [shell/recovery](#phase-1-shell-gate-та-відновлення), рядки 30, 34: failed migration блокує запуск, доступний retry зі збереженим URL; E03–E06 | Закрито незалежним review |
| 3 | Гонка й повторний виклик | [API та дані](#api-та-дані), рядки 24–26: snapshot/read-only; [shell/recovery](#phase-1-shell-gate-та-відновлення), рядки 36, 38: повний Query key, active-key rendering, obsolete abort, reversed A/B check, atomic concurrent seed | Закрито незалежним review |
| 4 | Порожні та граничні дані | [B01–B09](#приймальні-сценарії), рядки 53–61: 0/1/max, ties, hidden data, zoom і repeated seed; [shell/recovery](#phase-1-shell-gate-та-відновлення), рядок 38: pagination ceiling/first-page recovery | Закрито незалежним review |
| 5 | Rollback | [Rollback](#rollback), рядок 99: image/config revisions, volumes retained, forward fixes; [міграції](#план-міграції), рядок 95: destructive down лише в ephemeral DB; [side effects](#сайд-ефекти), рядок 91: немає зовнішнього deployment/платежів/повідомлень | Закрито незалежним review |

Reviewer `plan_defense`, 2026-09-03: **0 дірок, gate зелений**. Закрито останню дірку щодо запізнілої відповіді попереднього URL. Повна перевірка рішень — [gates.md](gates.md); runtime і acceptance tests ще мають бути виконані під час реалізації.

## Конституція та review

Під час review перевірити: domain/application dependency direction; відсутність transport DTO і SQL models у domain; відсутність mutable global SSR state; відсутність fake auth/media/security claims; runtime versions and exact lockfile; no secrets; no unused speculative ports/dependencies. Результат незалежних gates, автор і кожне blocking рішення записуються у ledger до імплементації. Якщо перевірка заблокована середовищем, залишити її pending із точною причиною й командою відтворення; не називати foundation Done.
