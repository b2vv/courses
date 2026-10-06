# Фаза `accounts` — звіт

> Дельта до `plan.md`. Що реально зроблено, як перевірено, що лишилось обмеженням.
> Статус: реалізовано A01–A15; закрито повним упорядкованим `make check`.

## 1. Що збудовано (за тасками)

| Таск | Коміт(и) | Суть |
|---|---|---|
| A01 | `1e4cd49` | `platform/mail.Sender` (SMTP, context-aware, MIME multipart, QP UTF-8, CRLF-guard) + `config.MailConfig` + Mailpit `v1.31.2` (base+e2e) |
| A02 | `0f6a44b` | Міграція 00014 (`email_verified`, `totp_secret`, `totp_enabled`, `account_tokens`, `recovery_codes`) + seed verified + `account_policy.go` (реєстрація 8–128/ім'я1–100, токени, зміна пароля/email) |
| A03 | `5e52722`,`7656d68` | Шов `EmailVerified` (`httpserver.Actor`); порти UserWriter/TokenStore/RateLimiter/Mailer; postgres/redis адаптери; **23505→`email_taken`**; атомарний consume; `DeleteUserSessionsExceptCurrent` |
| A04 | `3a53e7b`,`e981ad8` | Use-cases: register(автологін)/verify/resend/forgot(нейтрально)/reset(усі сесії)/change-password(крім поточної)/request+confirm-email-change(TOCTOU); Login remember-me (12h/3дні) + login rate-limit |
| A05 | `482e4a2` | `access.CanStream` (окремо від CanBeViewedBy — B1); гейт контенту (playback/transcript) і мутацій (`email_not_verified`) у StartPurchase/Checkout/Publish/ReportComment/PutReview/ReportReview |
| A06 | `39acf06` | HTTP `/api/v1/auth/*` + error→HTTP + composition + openapi + Orval; `/auth/me` += `emailVerified` |
| A07 | `65128de` | Фронт SessionRepository + application + zod |
| A08 | `ccb45cc`,`d4325bc` | Екрани register/verify/forgot/reset/settings; банер верифікації; play-гейт; i18n uk/en; a11y |
| A09 | `d7c771e`,`862ff94` | e2e `accounts.spec.ts` (Acc01–Acc31) + axe + читання Mailpit |
| A11 | `eb931d9`,… | CSP: per-request nonce у `server.mjs`; nginx базовий лише на api/media/error; Go JSON; **enforcing у проді, report-only в e2e** (0 app-порушень на повному e2e; e2e report-only бо axe/addStyleTag інжектять unnonce'd скрипт/стиль) |
| A12 | `a61fd66` | Домен 2FA: TOTP RFC6238 (±1 вікно, constant-time), otpauth, резервні коди, enroll/disable |
| A13 | `cef8412`,`ae5888c` | 2FA application (enroll/confirm/двокроковий Login+Complete/disable/regenerate/admin-disable), pending у Redis, postgres TwoFactor/Recovery stores |
| A14 | `2ab0968` | 2FA HTTP `/auth/2fa/*` + admin-disable + login `two_factor_required` + openapi + Orval |
| A15 | `49a2c5d` | 2FA фронт (QR enroll, 2-й крок входу, recovery, disable, regenerate) + `/me.twoFactorEnabled` |

## 2. Приймальна таблиця (Acc01–Acc31)

Усі рядки `plan.md` §7 закриті й перевірені в повному `make check` (e2e 101 passed). Ручних — 0.
Acc01–Acc24 (реєстрація, верифікація-гейт, forgot/reset, зміна пароля/email, enumeration, rate-limit,
remember-me, CSP, атомарний consume, гонка/TOCTOU) та Acc25–Acc31 (CSP, 2FA enroll/2-й крок/recovery/
disable/regenerate) покриті `accounts.spec.ts` + unit/contract/integration відповідних шарів.

## 3. Дефекти, знайдені дорогою (цінність гейту)

- **Mailer вів на 404:** лінки reset/email-change будувались на `/reset-password` і `/change-email`,
  яких немає серед маршрутів; виправлено на `/password/reset` і `/email/change/confirm` (`d813aa9`).
- **Чекбокс не клікався:** декоративна `.ui-checkbox-mark` перехоплювала pointer events (`d813aa9`).
- **Крос-модульні integration-фікстури:** гейт A05 правильно відхиляв неверифікованих; 4 модулі оновлено
  на `EmailVerified:true` (повний гейт зловив те, що точкові unit не бачили — урок AGENTS).
- **GATE 2 (план):** 3 проходи рецензії виловили B1 (архітектурні тести), B2 (шов Actor), B3 (сесія крім
  поточної), NB1–NB3 (CSP: поверхня, подвійний заголовок, `media/worker blob:` щоб не зламати Phase 6).
- **4 реальні дефекти знайдено під час GATE 1-рев'ю коду** (CRLF-injection, boundary-імпорти, валідатор
  сесії 12h ламав remember-me, copyfrom-каст у sqlc) — Codex ловив сам.

## 4. Верифікація

Повний упорядкований `make check` — **exit 0**: format-check, lint (+architecture, TruthTable/AST цілі),
typecheck, test, integration (identity/postgres+redis, 4 модулі), check-generated (sqlc+Orval у синхроні),
build, up, seed, smoke, **e2e 101 passed** (вкл. accounts.spec Acc01–Acc31, playback/ASR/медіа під enforcing
CSP без порушень). frontend rstest 616. Чужі Docker-стеки зупинялись/піднімались назад для гейту (15/15).

## 5. Обмеження (несуть магнітуду/механізм)

- **TOTP-секрет у БД зберігається як є** (не зашифрований KMS/app-key). Магнітуда: рівнозначно чутливості
  `password_hash` (теж у БД). Механізм прод-обгортки — окреме рішення при деплої.
- **CSP `style-src 'unsafe-inline'`**: inline-`style` атрибути React Aria/ProgressBar (4 місця) не nonce-яться;
  `script-src` — строгий (nonce). Механізм зняти: винести inline-стилі у класи (не ціль цієї фази).
- **CSP у проді enforcing, в e2e report-only** (`CSP_REPORT_ONLY=true` лише у `docker-compose.e2e.yml`):
  `@axe-core/playwright` і `addStyleTag` інжектять unnonce'd скрипт/стиль, які enforcing блокує — це тест-тулінг,
  не ресурси застосунку (у проді їх немає). Повний e2e під report-only дав 0 app-порушень → директива повна.
- **nginx login-ліміт послаблено лише в e2e** (`login-req.e2e.conf`, burst 50); прод лишається строгим
  (10r/m burst5). Механізм: окремий include, змонтований лише в `docker-compose.e2e.yml`.
- **Соцлогін/OAuth (Google), 2FA обов'язкове для ролей, WebAuthn, «довірений пристрій»** — поза фазою
  (майбутні шви; структура ідентичності лишена розширюваною).
- **Mailpit лише тест-режим**: реальна доставка назовні — через той самий порт `Mailer`, інший адаптер,
  при деплої (Р1 відкладено власником).

## 6. Наступне
Enrollment (K1) — наступна продуктова фаза (`work/AGENDA.md`); деплой відкладено. Після мерджу —
синк `AGENTS.md`/`CTO-RESEARCH`/`AGENDA` під закриття accounts.
