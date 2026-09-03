# Звіт з аудиту безпеки — courses (Phase 1–2)

| Поле | Значення |
|---|---|
| Дата | 2026-09-03 |
| Scope | Foundation Phase 1 + публічний каталог Course (Phase 2) |
| Commit / гілка | `main` (на момент аудиту) |
| Метод | Статичний огляд коду: backend HTTP/Postgres, Nginx, SSR frontend, Compose, логи/секрети |
| Поза scope | Phase 3–8 (identity, upload, playback, payments) — не оцінювались як реалізовані функції |

## Підсумок

Критичних і high-вразливостей у поточному публічному surface не виявлено.

Вхідна валідація, параметризований SQL з екрануванням LIKE, безпечні error envelope, deny `/media`, блокування metrics на edge і request-scoped SSR виглядають коректно.

Основні реальні ризики — **availability (DoS через SSR→API)** і **hardening перед публічним деплоєм**, а не витік даних чи RCE.

| Severity | Кількість |
|---|---|
| Critical | 0 |
| High | 0 |
| Medium | 2 |
| Low | 6 |
| Info (навмисні межі) | кілька — див. розділ нижче |

---

## Знахідки

### MEDIUM-1. SSR обходить Nginx rate limit → DoS каталогу / пошуку

| | |
|---|---|
| **Severity** | Medium |
| **Тип** | Availability / DoS |
| **Файли** | `infrastructure/nginx/nginx.conf` (limit_req лише на `/api/`); `frontend` SSR loader каталогу; `backend/sql/queries/courses.sql` (`ILIKE '%pattern%'`); `docker-compose.yml` (порт Nginx) |

**Опис.** Rate limit (`limit_req`) застосовується лише до `location ^~ /api/`. Маршрут `location /` (SSR) без ліміту. Frontend SSR викликає API за адресою `http://api:8080` у внутрішній Docker-мережі і **не проходить** через Nginx `limit_req`. Кожен запит з `q=` виконує leading-wildcard `ILIKE` у repeatable-read транзакції.

**Сценарій атаки.** Зловмисник масово відкриває `GET /courses?q=…` (або багато різних `q`). Nginx не обрізає SSR. API/Postgres отримують навантаження → каталог недоступний.

**Вплив.** Вичерпання CPU Postgres/API; відмова в обслуговуванні публічного каталогу.

**Поточне пом’якшення.** Compose публікує Nginx лише на `127.0.0.1` — ризик низький локально, зростає при публічному bind / TLS cutover.

**Рекомендація.** Rate-limit на `location /` і/або окремий ліміт SSR→API; обмежити вартість пошуку (індекси, мінімальна довжина `q`, budget). Перевірити E09 не лише на `/api/*`.

---

### MEDIUM-2. Спецсимволи в паролях ламають Compose credential URL

| | |
|---|---|
| **Severity** | Medium |
| **Тип** | Misconfiguration / credential handling |
| **Файли** | `docker-compose.yml` — `DATABASE_URL`, `REDIS_URL` |

**Опис.** Паролі підставляються в URL без encoding:

`postgres://${USER}:${PASSWORD}@postgres:5432/...`  
`redis://:${REDIS_PASSWORD}@redis:6379/0`

Символи `@`, `:`, `/`, `#` тощо ламають парсинг URL.

**Сценарій.** Оператор ставить «сильний» пароль зі спецсимволами → додаток підключається з усіченим/неправильним секретом або падає непередбачувано.

**Вплив.** Збій деплою або неочікувана поведінка автентифікації до залежностей при виході за `.env.example`.

**Рекомендація.** URL-encode паролів у Compose, або передавати host/user/password окремими env і збирати DSN у коді.

---

### LOW-1. Довіра до клієнтського `traceparent`

| | |
|---|---|
| **Severity** | Low |
| **Тип** | Observability abuse / cost DoS |
| **Файли** | `backend/platform/httpserver/server.go`; `backend/platform/observability/tracing.go` |

**Опис.** API витягує W3C `traceparent` з клієнтських заголовків. Sampler — `ParentBased(TraceIDRatioBased(0.1))`: якщо клієнт ставить sampled-flag, форсується повний sampling (не 10%). Також підставляються attacker-chosen trace ID у телеметрію.

**Вплив.** Роздуття обсягу OTLP і плутанина в трейсах, коли exporter увімкнено.

**Рекомендація.** На edge/API ігнорувати inbound `traceparent` від публічних клієнтів, або sampler без довіри до untrusted parent.

---

### LOW-2. Redis пароль у process argv

| | |
|---|---|
| **Severity** | Low |
| **Тип** | Credential exposure (local) |
| **Файли** | `docker-compose.yml` — `redis-server --requirepass` |

**Опис.** Пароль Redis потрапляє в командний рядок процесу і може бути видимий через `ps` / `docker top`.

**Вплив.** Розкриття credentials у shared/local середовищі. Пом’якшено localhost-only publish у base/dev Compose.

**Рекомендація.** Пароль через config-файл / Redis ACL, не через argv.

---

### LOW-3. Production frontend image містить повний `node_modules` (включно з devDependencies)

| | |
|---|---|
| **Severity** | Low |
| **Тип** | Attack surface / supply chain blast radius |
| **Файли** | `frontend/Dockerfile` |

**Опис.** `npm ci` ставить усі залежності; production stage копіює весь `node_modules` (Playwright, MSW, Oxlint тощо).

**Вплив.** Більший образ і ширша поверхня при компрометації контейнера; само по собі не remote RCE.

**Рекомендація.** `npm ci --omit=dev` або копіювати лише runtime-залежності.

---

### LOW-4. Course slug без URL-encoding у generated client

| | |
|---|---|
| **Severity** | Low |
| **Тип** | Defense-in-depth |
| **Файли** | `frontend/src/adapters/api/generated/courses.ts`; мітигація: `ValidateSlug` / `validSlug` |

**Опис.** Шлях збирається як `` `/api/v1/courses/${slug}` `` без `encodeURIComponent`. Зараз charset slug блокує `/`, `?`, `#`.

**Вплив.** Практичний ризик низький завдяки валідації; прогалина defense-in-depth якщо валідацію колись обійдуть.

**Рекомендація.** `encodeURIComponent(slug)` у Orval mutator / generated client.

---

### LOW-5. Обмежений набір browser security headers

| | |
|---|---|
| **Severity** | Low |
| **Тип** | Defense-in-depth |
| **Файли** | `infrastructure/nginx/nginx.conf` |

**Опис.** Є `X-Content-Type-Options`, `Referrer-Policy`, `X-Frame-Options`. Немає CSP, Permissions-Policy, COOP/COEP, HSTS.

**Вплив.** Слабша позиція проти XSS/clickjacking для майбутнього багатшого UI. HSTS коректно відсутній до TLS.

**Рекомендація.** CSP (спочатку report-only), Permissions-Policy; HSTS після TLS.

---

### LOW-6. Locale cookie без прапорця `Secure`

| | |
|---|---|
| **Severity** | Low |
| **Тип** | Cookie hygiene |
| **Файли** | `frontend/src/presentation/LanguageSwitcher.tsx` |

**Опис.** Cookie `courseforge_locale` — `Path=/; Max-Age=…; SameSite=Lax` без `Secure`.

**Вплив.** Нечутлива preference-cookie; прийнятно для локального HTTP. Додати `Secure` після TLS.

---

## Info — навмисні межі Phase 1–2 (не вразливості)

| Пункт | Доказ / примітка |
|---|---|
| `/media/*` → 403 | `nginx.conf`; volume RO на Nginx. За AGENTS.md / plan до playback authorization |
| `/metrics`, `/readyz` → 404 на edge | Metrics лишаються у Docker-мережі для Prometheus — by design |
| `/admin`, `/learn` | Client placeholders + `LockedArea`; fail-closed UI без приватних даних/API. Не Phase 3 auth |
| Немає CSRF | Лише read-only GET — N/A до mutations |
| Немає CORS headers | Fail-closed для browser cross-origin; same-origin через Nginx |
| Request-ID spoofing | API генерує власний ID; клієнтський не довіряється |
| Cookie forwarding на SSR | Per-request; публічний каталог; Query keys без cookie — OK до private data |
| `sslmode=disable` | Очікувано у приватній Compose-мережі |
| `.env.example` placeholders | Local-only; `.env` у `.gitignore` |
| Worker `/status` | Internal-only; порожні handlers — чесна поведінка |

---

## Що перевірено і виглядає добре

1. **HTTP / query / SQL injection** — дублікати ключів, невідомі параметри, non-decimal числа, NUL, oversize `q`, поганий sort/slug → `invalid_*`. LIKE: escape `!`, `%`, `_` + `ESCAPE '!'`. Помилки — лише коди/безпечні повідомлення.
2. **Postgres** — sqlc parameters; RR read-only tx для list+count; `statement_timeout`; slug lookup після validate.
3. **Nginx** — media 403; metrics/readyz blocked; safe JSON access log без URI/query; `error_log /dev/null`; `server_tokens off`; API rate limit присутній.
4. **Frontend SSR** — немає `dangerouslySetInnerHTML`; course fields як React text; немає `fetch` у presentation; `apiOrigin()` відхиляє userinfo/path/query; `Cache-Control: private, no-store`; per-request QueryClient.
5. **Секрети / health** — HTTP логи: method/route/status/ids; startup DB/Redis errors sanitized; edge `/healthz` — inert JSON.
6. **Redis queue** — foundation: порожній Asynq mux; logger без payload.
7. **Path traversal / SSRF** — slug charset блокує `../`; frontend відхиляє `/api/` і `/media/` як static; API origin не user-controlled.

---

## Рекомендації перед публічним деплоєм (пріоритет)

1. Rate-limit Nginx `location /` (і/або SSR→API token bucket); оцінити вартість пошуку.
2. Ігнорувати inbound `traceparent` від публічних клієнтів або змінити sampler.
3. URL-encode паролів DB/Redis (або non-URL env vars); `encodeURIComponent(slug)` у клієнті.
4. Prod image: `npm ci --omit=dev` (або лише runtime deps).
5. Redis: пароль не через `--requirepass` argv.
6. CSP (report-only → enforce), Permissions-Policy; `Secure` на cookies з TLS.
7. Phase 3: session identity у Query keys; ніколи не dehydrate cookies; CSRF для mutations; server-side auth на будь-які admin/learn API.
8. Опційно: приймати `X-Request-ID` лише від Nginx для кореляції логів.
9. Не монтувати `media-data` на Nginx, доки немає `auth_request` (менший ризик misconfig «зняти 403»).

---

## Відповідність архітектурним правилам (security-relevant)

| Правило | Статус |
|---|---|
| Media denied until auth | Pass |
| Metrics not on public edge | Pass |
| No secrets in logs / no query strings in access logs | Pass |
| Domain/application inward; no fetch in presentation | Pass |
| Request-scoped SSR deps / cookies not global | Pass |
| No fake auth/playback claims | Pass |
| Rate limiting completeness for all public entrypoints | Gap (SSR) |

---

## Вердикт

Система **придатна** для заявленого локального foundation + публічного catalog slice.

Перед публічним bind / TLS cutover варто закрити **MEDIUM-1** (SSR rate limit / вартість пошуку) і **MEDIUM-2** (encoding паролів), плюс OTEL parent trust і prod image contents як deploy hardening.

Phase 3–8 (identity, uploads, playback, payments) потребують окремого аудиту після реалізації.
