# SA13-F — P-F6 PostgresDSN sslmode default

## Decision (2026-10-30, Courses 1 + owner, закриває альтернативу)
- **Finding:** `audit-2026-10-24.md:137` «P-F6 PostgresDSN default sslmode=disable». `backend/platform/config/dsn.go::PostgresDSN` читає `POSTGRES_SSLMODE`, порожнє → default `"disable"`. Prod, що забув env, тихо конектиться без TLS. Клас LOW (аудит), двійник COOKIE_SECURE.
- **Рішення: закрити як задокументований prod-tail, БЕЗ коду в `dsn.go`.** Автор: Courses 1; схвалив власник 2026-10-30 (після розбору двох критиків-субагентів).
- **Спосіб перевірки (щоб finding був чесно закритий при deploy):** prod-конфіг ВСІХ DSN-споживачів (api, media-worker, asr-worker, upload, migrate) мусить мати `POSTGRES_SSLMODE=verify-full` + `sslrootcert`/CA-bundle, доступний цим процесам. Перевірка: `psql "$DSN"` конектиться лише за валідного серверного сертифіката; відсутність CA → старт падає. Разом із COOKIE_SECURE=true / Grafana pw — у prod-tail (`report`/AGENDA).

## Чому не код (розбір критиків, обидва read-only)
- **Критик безпеки** (`scratchpad/pf6-critic-security.md`): варіант «validate + keep default» — security-theater щодо ЦЬОГО finding. Небезпечний вектор — порожня env у prod (`empty→disable`), і він проходить валідацію мовчки; typo (`requre`) pgx/libpq і так відхиляє при конекті. Валідація не дає нової безпекової властивості. Рекомендація: prod-tail (verify-full), без коду — дзеркало COOKIE_SECURE. Ключове: `require` шифрує, але НЕ автентифікує сервер → MITM; безпечний prod-мінімум — **verify-full**.
- **Критик scope/гейта** (`scratchpad/pf6-critic-scope.md`): fail-closed (B) чіпає 5 compose-сервісів поштучно (спільного env-anchor нема: `<<: *runtime` — лише logging/restart/init) + `.env`; найгірше — тихий промах у list-form overlay override лишає гейт ЗЕЛЕНИМ, ламається пізніше в руках людини. P-F3 /metrics уже обрав default-open «бо default-closed червонив би» — прецедент проти fail-closed. COOKIE_SECURE і Grafana pw лишені prod-tail, не fail-closed; AGENDA зве P-F6 «дешевим, подрібнюваним». B ламає патерн; prod-tail його тримає.
- **Синтез:** варіант A схлопнувся (валідація — театр; лишалось тільки «document»). Реальний вибір був prod-tail-only vs B; власник обрав prod-tail-only. integration-харнес (`testsupport/containers`) НЕ використовує PostgresDSN → цей seam покривається лише unit-тестами на PostgresDSN, не e2e/integration (зафіксовано для deploy).
- **Залишковий ризик (прийнято):** prod усе ще мовчки `disable` при порожній env — той самий клас, що COOKIE_SECURE=false dev-default; закрито документуванням як prod-tail, не кодом. Магнітуда: discloses нічого локально (pre-launch, prod не існує); механізм закриття при deploy — verify-full+CA вище.

## Зроблено
- `.env.example`: додано `POSTGRES_SSLMODE=disable` з коментарем prod-вимоги (verify-full, не require; CA bundle; empty→disable тож ставити явно).
- Код `dsn.go` НЕ змінено (свідомо). Гейт не потрібен (docs/config-only, runtime-поведінка незмінна).
- `work/tasks.md` + prod-tail список оновлено.

## Наступне за AGENDA
- **P-F5** (panic value/stack лог) → M7 (HLS umask 0027); I-F6 (totp_secret-шифрування) окремо. Prod-хвіст (Grafana pw, COOKIE_SECURE=true, prometheus.yml metrics-auth, **POSTGRES_SSLMODE=verify-full+CA**) — до deploy.
