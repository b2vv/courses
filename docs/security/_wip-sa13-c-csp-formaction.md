# SA13-C — T-F2 CSP form-action

## Pre-code checkpoint (2026-10-29)
- **Задача / готовність:** SA13-батч, наступний за I-F5 (AGENDA-порядок). Джерело: `audit-2026-10-24.md:136` «T-F2 CSP lacks `form-action`». READY (standing-auth власника «Го»; окремий PR + незалежне рев'ю).
- **Дефект (механізм, звірено):** SSR HTML CSP у `frontend/server.mjs:8` не містить `form-action`. Ця директива **не успадковується від `default-src`**, тож нативний HTML form-submit нічим не обмежений — ін'єктована/перехоплена `<form>` може POST-ити куди завгодно (exfiltration). Enforcing у prod (report-only лише в e2e-оверлеї).
- **Impact (критичний):** `OrderStatusView.submitCheckoutForm` (`frontend/src/modules/purchase/ui/OrderStatusView.tsx:19`) будує **нативну cross-origin `<form method=post>`** на WayForPay і сабмітить. Дозволені origin'и вже кодифіковані клієнт-валідатором `purchase-repository.ts:35-48`:
  - prod: hostname `secure.wayforpay.com` + https — завжди;
  - e2e: hostname === `PUBLIC_WAYFORPAY_PAY_HOST` + http — лише коли env задано (порожнє в prod НЕ розширює allowlist).
  `make up-stub` (`docker-compose.stub.yml`) ганяє **enforcing** CSP на `http://localhost:8090/pay` → наївний `form-action 'self'` зламав би і прод-оплату, і локальний stub. `CSP_REPORT_ONLY=true` стоїть лише в `docker-compose.e2e.yml`.
- **Фікс (план):** у `server.mjs` зробити `form-action` динамічним, дзеркалячи валідатор:
  - завжди `'self' https://secure.wayforpay.com`;
  - коли `PUBLIC_WAYFORPAY_PAY_HOST` непорожній → додати `http://<host>:*` (http, той самий host, будь-який порт — бо порт живе у `WAYFORPAY_PAY_URL` на боці api, а валідатор і так ігнорує порт).
  `'self'` лишаємо для будь-яких same-origin нативних форм. Вставити директиву в CSP-рядок (напр. після `base-uri 'self'`).
- **Decision (closes alternative, 2026-10-29, author Courses13):** CSP form-action — **динамічний, не статичний `secure.wayforpay.com`**. Причина: `make up-stub` enforced'ить CSP на localhost-origin; статичний варіант розійшовся б із клієнт-валідатором (той приймає localhost-action) → форма проходить валідацію, але CSP блокує submit — мовчазний злам, який ніхто не тримає. Перевірка: enforcing-браузер проти up-stub (submit проходить) + порожній PUBLIC_WAYFORPAY_PAY_HOST → лише secure.wayforpay.com.
- **Поза scope:** API-CSP у `backend/platform/httpserver/server.go` (`default-src 'none'`) — JSON, форм немає, form-action нерелевантний. Інші SA13-зрізи.
- **Доказ (production path):**
  - Прецедент репо: CSP доводиться **реальним enforcing-браузером**, не unit-тестом (пор. zod-jitless close-out 2026-10-19; server.mjs CSP ніколи не був unit-тестований, rstest globs = `src/**`). e2e report-only enforcing не доведе.
  - up-stub (enforcing, PUBLIC_WAYFORPAY_PAY_HOST=localhost): відкрити checkout → submit → навігація на stub без CSP-violation у консолі (форма POST-иться). Негатив: `<form action=https://evil.test>` під тією ж CSP блокується (console violation).
  - Ordered `make check` зелений (e2e report-only — регресії немає; payment-спеки зелені).
- **Impact / гейти:** лише `frontend/server.mjs` (runtime SSR shim) — не src, не typecheck/rstest. Форматування oxfmt (per-slice, memory). Ordered `make check` перед PR.
- **Review:** `review_mode=existing_session`; reviewer `courses-96 [557618]`; writer Courses13 [c49267]; contact = SendMessage; обов'язкове незалежне рев'ю перед merge.
- **Transfer (SA13-C review):** snapshot = commit `9c7ffad` на гілці `sa13-csp-form-action` (НЕ запушено; спільне дерево → read-only `git show`/`git diff main...9c7ffad`). Стан: `sent` (2026-10-29). Чекаю ACK + verdict. Required: PASS на 9c7ffad → push + PR + self-merge. Авторське PASS/зелений гейт ≠ verdict.

## Доказ (enforcing, реальний браузер — виконано 2026-10-29)
Ізольований Node-харнес віддавав сторінку з ТОЧНО тим CSP-заголовком, який будує `server.mjs` (curl підтвердив байт-у-байт для обох режимів). Реальний Chrome (claude-in-chrome), кожна форма у ВЛАСНОМУ iframe (спільна ціль давала хибний ALLOWED — браузер скасовує всі submit, крім останнього, до CSP-перевірки; зловлено й виправлено):

| режим CSP | secure.wayforpay.com | http://localhost:8090 | evil.test |
|---|---|---|---|
| `form-action 'none'` (sanity) | BLOCKED | BLOCKED | BLOCKED |
| prod (`'self' …wayforpay`) | **ALLOWED** | BLOCKED | BLOCKED |
| e2e (`… + http://localhost:*`) | **ALLOWED** | **ALLOWED** | BLOCKED |

Висновки: (1) фікс блокує чужий origin (evil завжди BLOCKED) — ядро T-F2; (2) prod-оплата на secure.wayforpay.com проходить; (3) up-stub/e2e localhost проходить ЛИШЕ коли `PUBLIC_WAYFORPAY_PAY_HOST` заданий — порожнє в prod не розширює (інваріант тримається); (4) `'none'`-sanity доводить, що CSP таки блокує localhost (нема loopback-винятку). Прилад помінято після двох хибних ALLOWED (shared-iframe-артефакт), не гіпотезу.

## Журнал
- 2026-10-29: investigation T-F2 завершено (cross-origin WayForPay form; up-stub enforced; валідатор як джерело allowlist; e2e report-only). Гілка `sa13-csp-form-action` off main @7f5eb39.
- 2026-10-29: `server.mjs` form-action динамічний; CSP-білдер вивід звірено (node) для обох режимів; enforcing-браузерний доказ (таблиця вище) — працює. Ordered `make check` у фоні.
