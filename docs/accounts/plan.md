# Фаза `accounts` — план

> ЯК. Спирається на `spec.md` (GATE 1 2026-10-05, G1–G12, AC1–AC13). Статус: **чернетка, GATE 2 не пройдено**.

## 1. Reuse-first (що переюзуємо / що нове / чому)

| Потреба | Переюзуємо | Де |
|---|---|---|
| Хешування пароля | **`argon2id.Verifier.Hash`** (уже є, його ж юзає адмінський reset) | `identity/adapters/argon2id/verifier.go:56` |
| Звірка пароля при вході / зміні | `Verifier.Verify`, `VerifyDummy` (constant-time проти enumeration) | там само |
| Видача/резолв/відкликання сесії | `SessionStore.Rotate/Resolve/Revoke` | `identity/ports/authentication.go` |
| Відкликання **всіх** сесій користувача (для reset) | **SQL `DeleteIdentitySessionsForUser`** (уже є, викликається адмінкою 3×: administration.go:122/156/183) | `identity/adapters/postgres` |
| Відкликання всіх, **крім поточної** (для change-password) | **НОВИЙ** запит `DeleteUserSessionsExceptCurrent(user_id, current_digest)` — наявний видаляє й поточну, не підходить (B3) | `identity/adapters/postgres` (додати) |
| Годинник (expiry токенів) | `application.Clock` | `identity/application/identity.go:21` |
| Унікальність email | партковий unique (T01, `WHERE deleted_at IS NULL`) | `00004_identity_admin.sql` |
| Контракт доступу | `contracts/access` — розширюємо, не дублюємо | `internal/contracts/access/access.go` |
| Redis-клієнт | `platform/redis` (патерн адаптера — `playback/adapters/redis`) | `platform/redis`, `playback/adapters/redis/ownership.go` |
| HTTP envelope, коди помилок, request id | `platform/httpserver` | `platform/httpserver/` |
| Фронт HTTP-транспорт (дедлайн/скасування/класифікація) | `adapters/api/transport.ts` | reuse-index |
| SSR cookie-forwarding, заголовки | `adapters/api/request-headers.ts` | reuse-index |
| UI-примітиви (React Aria) | `presentation/ui/` | reuse-index |
| i18n uk/en | `app/providers/i18n.tsx`, `app/providers/locales` | — |

**Нове (бо немає аналога):**
- **Email-інфраструктура** — її немає взагалі. Shared concern → `backend/platform/mail/` (SMTP-клієнт + рендер шаблонів uk/en). Порт `identity/ports.Mailer`; адаптер `identity/adapters/mail` поверх `platform/mail`. Зовнішня залежність реальна (SMTP Mailpit) — порт виправданий.
- **Токени акаунта** (verify / reset / email-change) — нова таблиця + порт `TokenStore`.
- **Rate-limiter** — порт `identity/ports.RateLimiter`; адаптер `identity/adapters/redis` поверх `platform/redis`. Зовнішня залежність реальна (Redis).
- **UserWriter** — створення/оновлення користувача (сьогодні є лише читання + адмінські мутації).

## 2. Розміщення

**Backend** (модуль `identity`, shared — `platform/`):
- `platform/mail/` — SMTP sender + шаблони (uk/en), конфіг з `platform/config`.
- `identity/domain/` — **`User.EmailVerified bool`** (несе шов B2, нижче §4); політики: `NewRegistration` (валідація email / пароль 8–128 / ім'я 1–100 — **ріже на 100, хоч наявний `validDisplayName` допускає 120**), `AccountToken` (вид, строк, одноразовість), `ChangePassword` (стара≠нова, перевірка старої), `ChangeEmail`.
- **Шов `EmailVerified` (B2):** `SessionStore.Resolve` уже повертає `domain.User` → додаємо туди поле; `identity.Authenticate` кладе його в `httpserver.Actor` (нове поле `Actor.EmailVerified`, actor.go); кожен модуль читає `httpserver.ActorFromContext` і мапить у `access.Subject.EmailVerified`. Інші модулі НЕ читають таблицю `users` — факт постачає identity через Actor (виправляє хибне «кожен мапить свій факт»).
- Розмежування сесій (B3/G8): `ResetPassword` → `DeleteIdentitySessionsForUser` (усі); `ChangePassword` → новий `DeleteUserSessionsExceptCurrent` (усі, крім поточної).
- `identity/application/` — use-cases: `Register`, `VerifyEmail`, `ResendVerification`, `RequestPasswordReset`, `ResetPassword`, `ChangePassword`, `RequestEmailChange`, `ConfirmEmailChange`; розширення `Login` (remember-me).
- `identity/ports/` — `UserWriter`, `TokenStore`, `Mailer`, `RateLimiter`.
- `identity/adapters/postgres/` — UserWriter + TokenStore (sqlc); `identity/adapters/redis/` — RateLimiter; `identity/adapters/mail/` — Mailer → `platform/mail`.
- `identity/adapters/http/` — нові маршрути (нижче); композиція у відповідному root.
- `internal/contracts/access/` — `Subject.EmailVerified` + предикат `CanStream` (деталі §4).

**Frontend** (`frontend/src/…`):
- `routes/` — `register.tsx`, `verify-email.tsx` (лендинг токена), `password/forgot.tsx`, `password/reset.tsx` (лендинг), `settings/account.tsx` (зміна пароля + email); remember-me на `login.tsx`.
- `ports/SessionRepository.ts` — +`register`, `resendVerification`, `forgotPassword`, `resetPassword`, `changePassword`, `requestEmailChange`, `confirmEmailChange`.
- `adapters/api/` — реалізація репозиторію (через `transport.ts`), zod-схеми відповідей, Orval для нових DTO.
- `presentation/` — екрани форм, банер «підтвердіть пошту», гейт кнопки «грати» для неверифікованого.

**HTTP-маршрути** (під `/api/v1/auth`, mutating → `RequireSameOrigin`):
`POST /register`, `POST /verify-email`, `POST /resend-verification`, `POST /password/forgot`, `POST /password/reset`, `POST /password/change`, `POST /email/change`, `POST /email/change/confirm`. Усі — у `openapi.yaml`, Orval-генерація (інакше дрейф контракту, як B-20).

## 3. Email-інфра (Mailpit)

- Сервіс `mailpit` (`axllent/mailpit`) у `docker-compose.yml`: SMTP `1025`, web/API `8025`. `api` отримує `SMTP_HOST=mailpit`, `SMTP_PORT=1025` через `platform/config`.
- Також у `docker-compose.e2e.yml` (щоб e2e мав інбокс). e2e читає листи через Mailpit REST (`GET /api/v1/messages`, `GET /api/v1/message/{id}`), дістає посилання з токеном.
- Назовні нічого не йде (тест-режим). Реальний провайдер — при деплої (Р1), той самий порт `Mailer`, інший адаптер.

## 4. Контракт доступу — гейт верифікації (G1/AC2)

Проблема 1: `CanView` гейтить **метадані** курсу (опис/план), які неверифікованому **дозволені** → `CanView` НЕ чіпаємо.
Проблема 2 (B1): два архітектурні тести прибиті до форми `CanBeViewedBy`:
- `platform/architecture/access_mapping_ast_test.go` вимагає **рівно N** методів `CanBeViewedBy`, кожен — рівно один `return access.CanView(...)`.
- `shared_rules_test.go::TestTheAccessPredicateAgreesAcrossModules` звіряє всі копії з TruthTable, де **немає** виміру `email_verified`.
Тому **жоден `CanBeViewedBy` не перейменовуємо й не змінюємо** — усі 4 лишаються `= access.CanView(...)`; truth-table лишається без `email_verified`. Обидва тести лишаються зеленими без змін.

Механізм гейта:
- `access.Subject.EmailVerified bool` (саме поле CanView/TruthTable не ламає — підтверджено рецензентом).
- **Новий окремий чистий предикат** `access.CanStream(subject, course) = CanView(subject,course) && subject.EmailVerified`. Це **НЕ** метод `CanBeViewedBy`, тож AST-тест його не рахує; покривається власним юніт-тестом (окремо від TruthTable).
- Застосування `CanStream` — **у use-case шарі, не в domain `CanBeViewedBy`**:
  - **playback** — у use-case відтворення (Play / mint playback), поряд із наявним `CanBeViewedBy`.
  - **transcript** — у use-case читання транскрипту уроку.
- **Мутації** — перевірка `subject.EmailVerified` у точно названих use-cases (код помилки `email_not_verified`):
  - orders: **`StartPurchase`** (`orders/application/service.go:43`) і **`Checkout`** (:108) — перша мутація неверифікованого саме StartPurchase, не лише Checkout (E2).
  - discussion: **`Publish`** (коментар), **`ReportComment`**, **`PutReview`**, **`ReportReview`** (E2).
- **Не гейтимо:** каталог, сторінку курсу (опис/outline), читання коментарів/відгуків.

Шов факту — §2 (через `httpserver.Actor.EmailVerified`), не «кожен модуль читає users». Це зміна спільного контракту + Actor → **повний e2e**, не точковий (AGENTS §Lessons).

## 4b. Відмови й конкурентність (GATE 2 defense Q2–Q4, critique E3/E4)

- **SMTP/Mailpit недоступний:**
  - *signup* — user + сесія створюються першими (G3 автологін), лист — best-effort після; збій надсилання
    **не відкочує** акаунт: відповідь несе `verificationEmailSent:false`, користувач вживає resend (Acc03).
  - *forgot* — завжди нейтральна відповідь (anti-enum); для наявного акаунта токен створено, надсилання
    best-effort, збій логується без секретів, користувач повторює запит.
  - *email-change* — користувач авторизований, enumeration не стоїть: збій надсилання → транзієнтний
    `mail_unavailable`, повтор.
- **Redis/RateLimiter недоступний:** **fail-open** із гучним логом — доступність важливіша за вторинний
  захист; первинна перевірка пароля лишається constant-time (`VerifyDummy`). Стосується і login, і листів.
- **Гонка signup на той самий email (Q3):** pre-check дає `email_taken` у послідовному разі; у гонці
  обидва пройдуть pre-check, але партковий unique (`users_email_active_key`) дасть `23505` на INSERT —
  **мапимо pgconn 23505 → бізнес-код `email_taken`**, не 500.
- **Зміна email на зайнятий + TOCTOU (Q4/E4):** унікальність `new_email` перевіряємо і на request, і
  **повторно на confirm** (між ними адресу могли зайняти) → `email_taken`.
- **Одноразовість токена (атомарно):** consume = `UPDATE account_tokens SET consumed_at=now()
  WHERE id=$1 AND consumed_at IS NULL AND expires_at>now() RETURNING user_id`; 0 рядків → `token_expired`/
  `token_used`. Витримує подвійний submit verify/reset-лінка.
- **Rate-limit ключ (E3):** per-email ключ береться з **поданого** email незалежно від існування акаунта —
  інакше лічильник стає оракулом існування, що зводить нанівець constant-time login.

## 4c. Двофакторна автентифікація TOTP (G14)

- **Без нової зовнішньої залежності на бекенді:** TOTP (RFC 6238, 30с, 6 цифр, SHA1) — stdlib
  `crypto/hmac`+`crypto/sha1`+`encoding/base32`; звірка constant-time з вікном ±1 крок. QR **не** малюємо
  на бекенді: HTTP віддає `secret` + `otpauth://` URI, QR рендерить фронт (мала QR-бібліотека — нова
  залежність фронту, використовується → дозволено AGENTS).
- **Домен** (`identity/domain`): `TwoFactor` (секрет, enabled), політика enroll (підтвердити кодом перед
  увімкненням), резервні коди (генерація, одноразовість), disable (вимагає пароль + код/резервний).
- **Login стає двокроковим (змінює наявний `Login`):** якщо `totp_enabled`, `Login` **не видає сесію**,
  а повертає `two_factor_required` + короткоживучий **одноразовий** pending-токен (Redis, ~5 хв, прив'язаний
  до user+IP, видаляється при першому успішному `CompleteTwoFactor` або вичерпанні спроб). Pending **несе
  прапор remember-me**, обраний на кроці 1 (щоб крок 2 видав сесію потрібної тривалості). Новий use-case
  `CompleteTwoFactor` звіряє TOTP **або** резервний код і лише тоді видає сесію.
- **Rate-limit 2FA (N5):** RateLimiter ключем per-user+pending-токен — **5 спроб на pending** (вичерпання
  знищує pending → повторний логін, який лімітований per-email+IP, AC8); це обмежує перебір 6-значного коду.
- **Резервні коди (N2):** consume **атомарно**, як account_tokens — `UPDATE recovery_codes SET consumed_at=now()
  WHERE user_id=$1 AND code_hash=$2 AND consumed_at IS NULL RETURNING id`; 0 рядків → невірний/вжитий.
  Зберігаються лише як hash; регенерація видаляє всі попередні (consumed і ні).
- **Відновлення від локауту:** резервні коди — основний шлях; втрата пристрою+кодів → **адмін вимикає 2FA**
  користувачу (нова адмін-дія, поряд із наявним reset пароля). Без цього втрата = перманентний локаут.
  **Журнал (N1):** identity не має audit-таблиці (її нема в модулі), тож дію пишемо **структурованим
  лог-івентом** без секретів (не нова БД-таблиця — audit-таблиця identity поза скоупом фази). Це ж
  стосується наявного адмінського reset пароля, який теж не має DB-audit.
- **Сайд-ефект на існуючі e2e входи:** усі сідовані/інлайн-юзери `totp_enabled=false` за замовчуванням →
  наявний однокроковий вхід не змінюється для них. 2FA-гілку покриває лише новий e2e.

## 4d. CSP (G13)

**Поверхня (виправлено за GATE 2 B1):** скрипти-гідрацію несе **SSR-HTML з Node `frontend/server.mjs`**
через TanStack `<Scripts/>` (`routes/__root.tsx`), НЕ Go `platform/httpserver` (той віддає лише JSON
/api/v1 — браузер його як документ не рендерить). Тому:
- **Первинний CSP — на HTML-документі:** per-request **nonce** генерується у SSR-entry (`server.mjs`/роутер
  через `router.options.ssr.nonce`, який TanStack `<Scripts/>`/`ScriptOnce` вбудовано ставлять на кожен
  скрипт) і заголовок `Content-Security-Policy` ставить `server.mjs`:
  `default-src 'self'; script-src 'self' 'nonce-…'; style-src 'self' 'unsafe-inline'; img-src 'self' data:;
  media-src 'self' blob:; worker-src 'self' blob:; connect-src 'self'; object-src 'none'; base-uri 'self';
  frame-ancestors 'none'`.
- **`media-src`/`worker-src` з `blob:` обов'язкові (NB3)** — щоб не зламати вже-закриту Phase 6 playback:
  hls.js (`adapters/playback/hls-media.ts`) грає через MSE → `blob:` URL (media-src), `enableWorker:true`
  плодить blob-Web-Worker (worker-src), а VTT-субтитри (`learn/LessonPlayer.tsx`) теж `blob:`. `connect-src 'self'`
  покриває fetch HLS-сегментів (той самий origin). CSP — зміна, що зачіпає весь застосунок, тож повний e2e
  включно з playback обов'язковий (а не лише accounts).
  **`script-src` — без `unsafe-inline`** (nonce). **`style-src` лишає `'unsafe-inline'`** (NB1): є inline-`style`
  атрибути (ProgressBar/Meter width — `ui/data.tsx`, React Aria оверлеї позиціонуються inline у рантаймі),
  які React Aria не nonce-ить; зняти їх — не ціль цієї фази, тож це **назване обмеження** (мета — script-CSP).
- **nginx** додає базовий CSP (`add_header … always`) **лише на `/api`, `/media` та error-локаціях** —
  НЕ на `location /` (SSR): інакше другий заголовок AND-иться з nonce-CSP із `server.mjs` і без nonce
  заблокує inline hydration-скрипти (NB2).
- **Go `platform/httpserver`** — базовий CSP на JSON-відповідях (defense-in-depth), але це **не** скрипто-несуча
  поверхня; AC14 «строгий CSP» стосується саме HTML-документа.
- **Метод — report-only-first (щоб не вгадувати директиви):** спершу вмикаємо `Content-Security-Policy-Report-Only`
  з цільовою директивою вище, ганяємо **повний** e2e (включно з playback/ASR/медіа), збираємо реальні
  порушення з report у консолі/endpoint — це емпірично виявляє будь-яку пропущену директиву (media/worker/
  font/connect тощо) замість здогадок. Потім перемикаємо на enforcing `Content-Security-Policy`. Якщо
  enforcing-перемикання у цій фазі лишиться ризиковим — фіксуємо report-only + число/перелік порушень як
  назване обмеження у звіті (магнітуда, не обіцянка).

## 5. Сайд-ефекти · міграція · rollback

**Сайд-ефекти:**
- `contracts/access` змінюється → зачіпає course, transcript, playback, discussion (кожен мапить EmailVerified). Повний e2e обов'язковий.
- Нова колонка `users.email_verified` — **усі наявні (сідовані) користувачі мають стати `true`**, інакше e2e playback/mutation впадуть (неверифікований = read-only). Це «хто додає глобально-лічильне — прибирає за собою» (AGENTS): у `seed-identities.sql` додаємо `email_verified` у **обидві** частини — список колонок INSERT (значення `true`) **і** `ON CONFLICT … SET` (E1), не лише в UPSERT.
- **Аудит інлайн-створення юзерів (E1):** `frontend/e2e/ratings.spec.ts:139` і ~15 integration-тестів вставляють users без `email_verified` (DEFAULT false). Таск-умова: жоден e2e, що **логіниться** інлайн-створеним юзером і далі **грає/мутує**, не лишає його `false`; інлайн-SQL-автори, що пишуть повз HTTP-гейт, лишаються як є (гейт оминають). Acc20 розширюється цим аудитом.
- Новий Compose-сервіс `mailpit` у base та e2e.
- Нові маршрути в `openapi.yaml` + Orval → `check-generated`.

**Міграція `00014_accounts.sql`:**
- UP: `ALTER TABLE users ADD COLUMN email_verified boolean NOT NULL DEFAULT false`; **`ADD COLUMN totp_secret text NULL`, `ADD COLUMN totp_enabled boolean NOT NULL DEFAULT false`**; `CREATE TABLE account_tokens (id, user_id FK, purpose text CHECK IN ('email_verify','password_reset','email_change'), token_digest bytea/char, new_email text NULL, expires_at, consumed_at NULL, created_at)`; **`CREATE TABLE recovery_codes (id, user_id FK, code_hash, consumed_at NULL, created_at)`**; індекс на (user_id, purpose), на token_digest, на (user_id) recovery_codes.
- Одноразовість: `consumed_at IS NULL` + `expires_at > now()`; «повторне надсилання робить попередній недійсним» → при видачі нового інвалідуємо попередні того ж purpose (consumed_at=now() або delete).
- Дані: `seed-identities.sql` ставить `email_verified=true` 6 сідованим акаунтам (у тому ж UPSERT).
- DOWN: `DROP TABLE recovery_codes; DROP TABLE account_tokens; ALTER TABLE users DROP COLUMN totp_enabled, DROP COLUMN totp_secret, DROP COLUMN email_verified;`

**Rollback:** міграція down чиста (нова таблиця + колонка). Контрактну зміну відкочуємо git-revert гілки до мерджу (FF-гілка, не в main до зеленого гейта). Compose-сервіс mailpit — видалення сервісу, решта стека не залежить.

## 6. Конституція (маніфест `.claude/standards.md` §Конституція → AGENTS)

- **Домен чистий**: політики реєстрації/токенів/паролів — standard lib + contracts; без I/O, без framework. ✔
- **Application — домен+порти**: Mailer/TokenStore/RateLimiter/UserWriter як порти, адаптери всередину. ✔
- **Порт вимагає реальної зовнішньої залежності**: Mailer→SMTP, RateLimiter→Redis, UserWriter/TokenStore→Postgres. ✔ (жодного спекулятивного шва; Google-логін — задокументований майбутній шов, без інтерфейсу зараз).
- **2FA/TOTP — stdlib, без нової бекенд-залежності** (`crypto/hmac`,`sha1`,`base32`); QR — на фронті (мала бібліотека, використовується). Секрет TOTP у БД як є (тест-режим) — названо обмеженням, прод-шифрування окремо. ✔
- **CSP**: `script-src` без `unsafe-inline` (nonce); `style-src 'self' 'unsafe-inline'` (inline-стилі React
  Aria/ProgressBar — назване обмеження, NB1); `media-src/worker-src 'self' blob:` щоб не зламати Phase 6
  playback (hls.js MSE+worker, VTT blob, NB3); nonce лише на HTML (server.mjs), nginx-CSP не на SSR (NB2);
  метод report-only-first із заміром. Defense-in-depth, рішення/замір у звіт. ✔
- **contracts — лише чисті предикати**: `CanStream` — чиста функція, standard lib, без імпорту модулів; **окрема від `CanBeViewedBy`**, тож `access_mapping_ast_test` і `shared_rules` TruthTable лишаються зеленими без змін (B1). ✔
- **Генерований sqlc/Orval у адаптерах**, явний маппінг у домен. ✔
- **Без ORM / service locator / глобального мутабельного клієнта**: SMTP/Redis клієнти request-scoped через композицію. ✔
- **TS strict**: без `any`; зовнішні дані (токени, форми) валідуються zod; без fetch у presentation; Query для server-state; React Aria; семантичні токени, a11y/zoom/keyboard. ✔
- **Безпека**: токени зберігаються лише як digest (як сесії); у логах без секретів і без query-рядків; коди помилок явні (`email_not_verified`, `token_expired`, `invalid_credentials`, `email_taken`, `weak_password`). ✔

## 7. Приймальні сценарії (делеговано `acceptance-spec`; рядки Acc01…)

Групи: happy / межі / відмови — усі три присутні. Рівень: unit (U) / contract (C) / integration (I) / e2e (E).

| ID | Сценарій | Передумова | Дія | Спостережуваний результат | Чим |
|---|---|---|---|---|---|
| Acc01 | Реєстрація happy | гість | signup email+пароль+ім'я | створено user(student, verified=false), є сесія (автологін), у Mailpit 1 лист верифікації | E |
| Acc02 | Верифікація happy | є лист | клік по лінку | `email_verified=true`; банер зникає; грати можна | E |
| Acc03 | Повторне надсилання | лист протермінований | «надіслати знову» | новий лист; старий токен недійсний | E/I |
| Acc04 | Forgot happy | акаунт є | ввести email | 1 лист reset у Mailpit | E |
| Acc05 | Reset happy | є reset-лист | новий пароль | вхід новим паролем; усі старі сесії завершені | E |
| Acc06 | Зміна пароля | авторизований | стара+нова | пароль змінено; інші сесії завершені, поточна жива | I/E |
| Acc07 | Зміна email | авторизований | нова адреса | лист на нову; до кліку email старий; після — новий, verified | E |
| Acc08 | Гейт верифікації (межа) | verified=false, безкошт. урок | спроба грати | відмова `email_not_verified`; опис курсу видно | E |
| Acc09 | Гейт на мутацію (межа) | verified=false | спроба відгук/коментар/купівля | відмова `email_not_verified` | C/E |
| Acc10 | Signup на зайнятий email (відмова) | email існує | signup | пряма відмова `email_taken` зі шляхами вхід/ресет | C/E |
| Acc11 | Forgot на невідомий email (відмова) | email не існує | forgot | нейтральна відповідь, листа немає | C/I |
| Acc12 | Токен протермінований/вжитий (відмова) | старий/вжитий токен | verify/reset | відмова `token_expired`/`token_used` + шлях «надіслати знову» | C/I |
| Acc13 | Невірна стара при зміні пароля (відмова) | авторизований | хибна стара | відмова; пароль не змінено; сесії цілі | C/I |
| Acc14 | Rate-limit вхід (відмова) | 5 невдалих поспіль | 6-й вхід | блок/backoff ~15 хв; повідомлення не розкриває існування | I/E |
| Acc15 | Rate-limit листи (межа) | кілька forgot/signup поспіль | ще запит | подальші гальмуються (без спаму листами) | I |
| Acc16 | Remember-me (межа) | вхід без/з галочкою | логін | cookie живе 12 год / 3 дні відповідно | I |
| Acc17 | Валідація нижні+верхні межі (відмова) | — | пароль <8 **і >128** / ім'я порожнє **і >100** / кривий email | бізнес-коди `weak_password`/валідація на обох межах; без секретів у логах | U/C |
| Acc18 | Мова листа (межа) | UI uk / en | signup | тіло листа відповідною мовою | I/E |
| Acc19 | a11y нових екранів | — | обхід клавіатурою + axe | фокус видимий, ≥44px, 200% zoom, 0 порушень | E(axe) |
| Acc20 | Сідовані лишаються verified + аудит інлайн (межа) | після міграції | — | 6 сідів `email_verified=true`; жоден e2e, що логіниться інлайн-юзером і грає/мутує, не лишає його false; наявні e2e грають/мутують як раніше | I/E |
| Acc21 | Гонка signup на той самий email (відмова) | email вільний | 2 паралельні signup | один успіх, інший — `email_taken` (23505→бізнес-код), не 500 | I |
| Acc22 | Зміна email на зайнятий + TOCTOU (відмова) | нова адреса зайнята (або зайнята між request і confirm) | request/confirm email-change | `email_taken` на відповідному кроці; стара адреса лишається | C/I |
| Acc23 | SMTP недоступний при signup (відмова залежності) | Mailpit зупинено | signup | user+сесія створені, `verificationEmailSent:false`; resend потім надсилає | I |
| Acc24 | Атомарний consume токена (межа) | дійсний verify/reset токен | подвійний submit лінка | перший успіх, другий — `token_used`; стан не подвоюється | I |

| Acc25 | CSP присутній, playback цілий (межа) | — | повний e2e (вкл. playback/медіа) | HTML несе CSP (script-src nonce, style/media/worker за §4d); **повний e2e зелений, консоль без CSP-порушень** (report-only+замір, якщо enforcing ризиковий) | E |
| Acc26 | 2FA enroll (happy) | авторизований без 2FA | ввести дійсний TOTP у налаштуваннях | 2FA увімкнено; набір резервних кодів показано один раз | E |
| Acc27 | Двокроковий вхід (happy) | 2FA увімкнено | пароль → TOTP | сесія лише після коду; невірний код — відмова, лімітується | E |
| Acc28 | Резервний код (межа) | 2FA увімкнено, пристрій недоступний | вхід резервним кодом | успіх один раз, код згорає; повтор — відмова | I/E |
| Acc29 | Вимкнення 2FA + адмін-reset (відмова/межа) | 2FA увімкнено | disable без пароля/коду; окремо — адмін вимикає | без пароля+коду відмова; адмін-дія знімає 2FA (анти-локаут), журналюється | C/E |
| Acc30 | 2FA pending не дає доступу (відмова) | 2FA увімкнено, лише пароль введено | спроба API до CompleteTwoFactor | сесії немає; доступ відмовлено; pending одноразовий | I |
| Acc31 | Регенерація резервних кодів (межа) | 2FA увімкнено, коди видано | перегенерувати | новий набір; усі старі коди недійсні | I |

Ручних рядків: 0.

## 8. Нотатки до розбивки на таски
Порядок залежностей:
- (T1) `platform/mail` + Mailpit у base/e2e compose + `platform/config` SMTP.
- (T2) міграція 00014 (`email_verified`, `account_tokens`) + seed verified у обидві частини + domain
  (`User.EmailVerified`, політики реєстрації/токена/пароля/email, ім'я ≤100).
- (T3) шов B2: `SessionStore.Resolve` несе EmailVerified, `httpserver.Actor.EmailVerified`, заповнення в
  `identity.Authenticate`; ports (`UserWriter`, `TokenStore`, `Mailer`, `RateLimiter`) + postgres/redis/mail
  адаптери; новий `DeleteUserSessionsExceptCurrent`; мапінг pg 23505→`email_taken`.
- (T4) application use-cases (Register/Verify/Resend/Forgot/Reset/ChangePassword/RequestEmailChange/
  ConfirmEmailChange + Login remember-me) з failure-path §4b; атомарний consume.
- (T5) `contracts/access.CanStream` (+юніт) і застосування в use-case playback (Play/mint) і transcript;
  гейти мутацій `email_not_verified` у StartPurchase/Checkout, Publish/ReportComment/PutReview/ReportReview;
  мапінг Actor.EmailVerified→Subject у 4 модулях+orders.
- (T6) HTTP-маршрути + openapi.yaml + Orval; **`GET /auth/me` DTO += `email_verified`** (банер Acc02 на нього спирається, N1).
- (T7) фронт `SessionRepository` + application + zod.
- (T8) фронт екрани (register/verify/forgot/reset/settings) + банер + гейт кнопки «грати» + remember-me + i18n.
- (T9) e2e `accounts.spec.ts` + axe + читання Mailpit API; аудит інлайн-юзерів (E1).
- (T10) впорядкований `make check` (повний e2e — крос-модульна зміна) + звіт Acc01–Acc30 порядково + PR.
- **(T11) CSP** (G13): nonce + строгий CSP на HTML-документі у **`frontend/server.mjs`** (TanStack
  `<Scripts/>`); базовий CSP на рівні nginx-сервера; базовий CSP на JSON у Go (defense-in-depth). Fallback
  report-only + замір, рішення у звіт. → Acc25.
- **(T12) 2FA домен+міграція** (G14): `totp_secret`/`totp_enabled`/`recovery_codes` у 00014; домен TwoFactor,
  TOTP RFC6238 (stdlib), резервні коди. → Acc26–Acc29.
- **(T13) 2FA application**: enroll/CompleteTwoFactor/disable/regenerate-codes + двокроковий Login
  (pending-токен у Redis) + адмін-disable-2FA; rate-limit 2FA. → Acc26–Acc30.
- **(T14) 2FA HTTP+openapi+Orval**: `/auth/2fa/*` (enroll/confirm/complete/disable/recovery), адмін-маршрут;
  зміна контракту Login (two_factor_required). → Acc26–Acc30.
- **(T15) 2FA фронт**: налаштування (QR+секрет+резервні коди), другий крок входу, i18n, a11y; QR-бібліотека. → Acc26–Acc29.
Порядок: T11 незалежний (паралельно). 2FA (T12→T13→T14→T15) після базового кластера A03/A04 (шов сесії),
вливається в e2e (T9) і гейт (T10). Кожен таск посилається на рядки Acc.
