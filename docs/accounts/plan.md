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

## 5. Сайд-ефекти · міграція · rollback

**Сайд-ефекти:**
- `contracts/access` змінюється → зачіпає course, transcript, playback, discussion (кожен мапить EmailVerified). Повний e2e обов'язковий.
- Нова колонка `users.email_verified` — **усі наявні (сідовані) користувачі мають стати `true`**, інакше e2e playback/mutation впадуть (неверифікований = read-only). Це «хто додає глобально-лічильне — прибирає за собою» (AGENTS): у `seed-identities.sql` додаємо `email_verified` у **обидві** частини — список колонок INSERT (значення `true`) **і** `ON CONFLICT … SET` (E1), не лише в UPSERT.
- **Аудит інлайн-створення юзерів (E1):** `frontend/e2e/ratings.spec.ts:139` і ~15 integration-тестів вставляють users без `email_verified` (DEFAULT false). Таск-умова: жоден e2e, що **логіниться** інлайн-створеним юзером і далі **грає/мутує**, не лишає його `false`; інлайн-SQL-автори, що пишуть повз HTTP-гейт, лишаються як є (гейт оминають). Acc20 розширюється цим аудитом.
- Новий Compose-сервіс `mailpit` у base та e2e.
- Нові маршрути в `openapi.yaml` + Orval → `check-generated`.

**Міграція `00014_accounts.sql`:**
- UP: `ALTER TABLE users ADD COLUMN email_verified boolean NOT NULL DEFAULT false`; `CREATE TABLE account_tokens (id, user_id FK, purpose text CHECK IN ('email_verify','password_reset','email_change'), token_digest bytea/char, new_email text NULL, expires_at, consumed_at NULL, created_at)`; індекс на (user_id, purpose) і на token_digest.
- Одноразовість: `consumed_at IS NULL` + `expires_at > now()`; «повторне надсилання робить попередній недійсним» → при видачі нового інвалідуємо попередні того ж purpose (consumed_at=now() або delete).
- Дані: `seed-identities.sql` ставить `email_verified=true` 6 сідованим акаунтам (у тому ж UPSERT).
- DOWN: `DROP TABLE account_tokens; ALTER TABLE users DROP COLUMN email_verified;`

**Rollback:** міграція down чиста (нова таблиця + колонка). Контрактну зміну відкочуємо git-revert гілки до мерджу (FF-гілка, не в main до зеленого гейта). Compose-сервіс mailpit — видалення сервісу, решта стека не залежить.

## 6. Конституція (маніфест `.claude/standards.md` §Конституція → AGENTS)

- **Домен чистий**: політики реєстрації/токенів/паролів — standard lib + contracts; без I/O, без framework. ✔
- **Application — домен+порти**: Mailer/TokenStore/RateLimiter/UserWriter як порти, адаптери всередину. ✔
- **Порт вимагає реальної зовнішньої залежності**: Mailer→SMTP, RateLimiter→Redis, UserWriter/TokenStore→Postgres. ✔ (жодного спекулятивного шва; Google-логін — задокументований майбутній шов, без інтерфейсу зараз).
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
- (T10) впорядкований `make check` (повний e2e — крос-модульна зміна) + звіт Acc01–Acc24 порядково + PR.
Кожен таск посилається на рядки Acc.
