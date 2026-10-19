# Codex brief — A11: CSP (report-only-first, nonce у server.mjs)

Репо `courses`, гілка `feat/accounts-selfservice`. План `docs/accounts/plan.md` §4d (CSP, з урахуванням B1/NB1/NB2/NB3).
Роль: реалізуєш A11; гейт+коміт — Courses 1.

РОБОЧИЙ ФАЙЛ: /Users/strelia/projects/courses/work/_wip-a11-codex.md
Append після КОЖНОГО кроку; перший рядок `ДАЛІ:`. Фінал = стислий переказ.

## Поверхня (уже зʼясовано в §4d — не плутай)
- HTML-документ зі скриптами віддає **`frontend/server.mjs`** (TanStack SSR, `<Scripts/>` у `routes/__root.tsx`),
  НЕ Go. Go `platform/httpserver` віддає лише JSON. nginx проксує SSR через `location /`, API через `location /api`.

## Що зробити
1. **`frontend/server.mjs`**: генеруй **per-request nonce** (crypto random) і прокинь у TanStack так, щоб
   `<Scripts/>`/`ScriptOnce` ставили `nonce` на кожен інлайн-скрипт (механізм: `router.options.ssr.nonce`
   або що підтримує встановлена версія TanStack Start — звір у node_modules). Став заголовок
   **`Content-Security-Policy-Report-Only`** (report-only-first!) із цільовою директивою:
   `default-src 'self'; script-src 'self' 'nonce-<N>'; style-src 'self' 'unsafe-inline'; img-src 'self' data:;
    media-src 'self' blob:; worker-src 'self' blob:; connect-src 'self'; object-src 'none'; base-uri 'self';
    frame-ancestors 'none'`.
   Чому report-only спершу: щоб повний e2e (A10) емпірично зібрав реальні порушення (напр. шрифти/ще щось) до
   enforcing. Передбач змінну/прапор, щоб легко перемкнути Report-Only→enforcing (`Content-Security-Policy`) —
   рішення про перемикання ухвалить Courses 1 за результатом e2e.
2. **`infrastructure/nginx/nginx.conf`**: базовий CSP `add_header ... always` **лише** на `location ^~ /api/`,
   `location .../media/...` та error-локаціях — **НЕ на `location /`** (SSR), інакше буде другий заголовок і
   зламається nonce'd гідрація (NB2). Де додаєш add_header у локацію без власних — повтори наявні security-хедери
   (X-Content-Type-Options/Referrer-Policy/X-Frame-Options), бо add_header не успадковується при override.
3. **Go `platform/httpserver`**: базовий статичний CSP на JSON-відповідях (`default-src 'none'; frame-ancestors 'none'`
   або як доречно для суто-JSON) як defense-in-depth. Не чіпає SSR.

## Правила
- Без нових залежностей (nonce — stdlib crypto на Node). Не ламай наявну гідрацію.
- `style-src` лишає `'unsafe-inline'` (React Aria/ProgressBar inline-стилі — NB1); `script-src` без unsafe-inline.
- НЕ комість. Браузерну перевірку CSP лишаєш Courses 1 (A10 e2e).

## Критерій готовності
- `cd frontend && npm run typecheck`, `npm run lint`, `npm run build` зелені; `npm run test` зелений.
- `docker compose -f docker-compose.yml -f docker-compose.e2e.yml config` валідний (nginx-зміни синтаксично ок —
  Courses 1 перевірить).
- У робочому файлі чітко напиши, ЯК перемкнути Report-Only→enforcing (один прапор/рядок), щоб Courses 1 зробив це
  після зеленого e2e.
