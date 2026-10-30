# Deploy readiness — single checklist of all deploy-gated tails

| | |
|---|---|
| Дата | 2026-10-30 |
| HEAD | `a876b87` |
| Статус деплою | **Відкладено власником 2026-10-03** (test-mode, без prod-сервера). Це людське рішення, не черга. |
| Навіщо цей файл | Усі хвости, які свідомо лишили «до deploy», були розкидані по `docs/security/audit-2026-10-24.md`, SA13 wip-нотах, phase-report §Limitations і `docs/payment/`. Тут вони зведені в один gate, щоб рішення про деплій мало конкретний, перевірюваний обсяг. **Жоден пункт не роблять зараз** — більшість потребує реального prod-середовища й секретів від власника. |

**Легенда власник/агент:** **[O]** = потребує інфри/секрету/рішення власника; **[A]** = агент може підготувати код/конфіг наперед, коли власник дасть go; **[O+A]** = власник дає секрет/інфру, агент вмонтовує.

---

## A. TLS / транспорт (розблоковує S01, S02)

- [ ] **A1 [O] TLS termination у prod.** Передумова всього блоку. Джерело: `work/tasks.md:38` (S01 dep «TLS termination»). Перевірка: сервіс відповідає лише по HTTPS, http→redirect.
- [ ] **A2 [O+A] S01 — HSTS + Permissions-Policy + enforce CSP + `Secure` cookies.** Після A1. CSP зараз enforcing у SSR (zod jitless), але HSTS/Permissions-Policy та `Secure`-флаг чекають TLS. Джерело: `work/tasks.md:38` (audit LOW-5/LOW-6). Перевірка: response headers несуть `Strict-Transport-Security`, cookies мають `Secure`.
- [ ] **A3 [A] S02 — `__Host-` prefix на session-cookie.** Після A2 (вимагає `Secure`). Прибирає передумову duplicate-cookie класу повністю. Джерело: `work/tasks.md:39`, `docs/identity/report.md` §Security review. Перевірка: session-cookie названий `__Host-…`, браузер відхиляє його без `Secure`/`Path=/`/без domain.

## B. SA13 prod-хвости (потребують prod-env/секретів)

- [ ] **B1 [O+A] `COOKIE_SECURE=true`.** Dev-default `false`; у prod має бути `true` (йде разом із A1/A2). Джерело: `docs/security/audit-2026-10-24.md:142`, `docs/identity/report.md` §2 (`config.go` парсить `COOKIE_SECURE`). Перевірка: session-cookie з `Secure`.
- [ ] **B2 [O] Grafana — змінити default admin password.** Зараз network-only профіль, але при експозиції default-pw недопустимий. Джерело: audit:142, `docs/observability/report.md` §107 / AC5-AC6. Перевірка: дефолтний логін відхилено.
- [ ] **B3 [O+A] Prometheus metrics-auth.** `/metrics` має in-app bearer-guard (`METRICS_TOKEN`, default-open у test). У prod: `prometheus.yml` → `authorization.credentials(_file)` + `METRICS_TOKEN` у env КОЖНОГО сервісу (api, media-worker, asr-worker, upload). Джерело: `docs/security/_wip-sa13-e-metrics-guard.md:26`. Перевірка: scrape без токена → 404; з токеном → 200.
- [ ] **B4 [O+A] `POSTGRES_SSLMODE=verify-full` + CA bundle.** Для ВСІХ DSN-споживачів (api, media-worker, asr-worker, upload, migrate). Порожня env тихо → `disable` (прийнятий prod-tail, не код — owner 30.10). Джерело: `docs/security/_wip-sa13-f-pg-sslmode.md:6,12`. Перевірка: `psql "$DSN"` конектиться лише за валідного серверного сертифіката; відсутність CA → старт падає. ⚠️ integration-харнес НЕ йде через `PostgresDSN` → цей seam покритий лише unit-тестами, перевіряти вручну при deploy.

## C. Identity 2FA ключ / крипто (I-F6 хвости)

- [ ] **C1 [O+A] `IDENTITY_2FA_SECRET_KEY` — required/fail-closed у prod.** base64 ≥32B випадкового майстра; відсутній/невалідний → api не стартує. Dev/test мають dev-default. Джерело: I-F6 (`docs/security/_wip-sa13-i-f6-totp-encryption.md`, `config.Identity2FAKey`). Перевірка: старт api без валідного ключа падає.
- [ ] **C2 [O] Політика ротації майстра (N2).** Формат без key-id → ротація `IDENTITY_2FA_SECRET_KEY` = re-enrol усіх 2FA (міграція-reset 00024 скидає стан). Вирішити політику/комунікацію ПЕРЕД deploy. Джерело: I-F6 N2 (accepted-bar). Перевірка: задокументована процедура ротації.
- [ ] **C3 [A] (опційно, hardening) GCM AAD=user_id (N1).** Зараз AAD=nil → ciphertext не прив'язаний до user_id (DB-write transplant поза DB-read моделлю). Окремий слайс, не блокер. Джерело: I-F6 N1 (accepted-bar).

## D. Payment live (launch)

- [ ] **D1 [O] Новий WayForPay-магазин + новий ключ.** Старий `site_ex` видаляється власником; майбутній live-run — новий магазин/ключ. Джерело: `docs/payment/assumptions.md:250-257`, `docs/payment/report.md:93-94`. Перевірка: у кабінеті `site_ex` зник; підписаний запит старим ключем падає.
- [ ] **D2 [O] ФОП + публічний HTTPS `serviceUrl`.** Live-платежі потребують ФОП і досяжного callback-URL. Джерело: `docs/payment/report.md:156`. Перевірка: реальний callback `POST → 200`, signature-verified, applied once.
- [ ] **D3 [O+A] Live Approved-прогін на новому магазині** перед go-live (повторити Pay25-live на новому shop/key). Джерело: `docs/payment/report.md:53` (Pay25). Перевірка: 1-грн Approved, callback 200, рівно один `paid` + один active grant.

## E. Observability / конфіг prod

- [ ] **E1 [O+A] Monitoring-стек у prod.** Grafana/Loki/Prometheus/promtail зараз network-only профіль, не задеплоєні; provisioning (datasources + `platform-overview` dashboard) готовий. Джерело: `docs/observability/report.md` §107. Разом із B2/B3.
- [ ] **E2 [O+A] `APP_ENV=production`, канонічний `PUBLIC_ORIGIN`, `APP_RELEASE`.** Джерело: `docs/identity/report.md` §2, observability. Перевірка: origin-порівняння канонічне; release-мітка в логах/метриках збігається build↔runtime.

---

## Хто що робить (підсумок)

- **Чисто власник [O] (інфра/секрети/рішення):** A1 TLS, B2 Grafana pw, C2 rotation policy, D1 shop/key, D2 ФОП/URL.
- **Власник дає, агент вмонтовує [O+A]:** A2, B1, B3, B4, C1, D3, E1, E2.
- **Агент наперед [A]:** A3 (`__Host-` після TLS), C3 (AAD hardening, опційно).

**Нічого з цього не стартує без owner-go на деплій** (стояча відкладка 03.10). Коли власник відкриває деплій — цей файл стає gate; кожен пункт закривається з перевіркою вище.
