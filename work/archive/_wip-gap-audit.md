# Жива нота — K1 enrollment + аудит прогалин (2026-10-04)

> Код і `work/tasks.md` виграють за цю ноту. Obsidian-дзеркало — session note.

## Контекст
Власник обрав **K1 (enrollment)**, але назвав прогалини, які треба не прогавити:
**верифікація користувача**, **ресет пароля**, «можливо ще щось прогавили».
Тож перед плануванням — повний аудит прогалин базового функціоналу, не лише enrollment.

## Швидкий зріз коду (Courses 1, до агентів)
- **Email-інфраструктури НЕМА** — grep по smtp/mailer/sendgrid/ses у backend порожній.
  → і верифікація email, і self-service ресет пароля впираються в цей фундамент.
- `identity/application/identity.go` — лише `Login`, `Authenticate`, `Logout`. **Немає `Register`.**
- Фронт `routes/`: лише `login.tsx`. Немає register / forgot-password / verify-email.
- Адмінський ресет пароля — Є (T04/T05, admin_handler). Self-service ресету — нема.

## Дослідження (2 агенти, різні робочі файли)
- Агент A → `work/_wip-gap-audit-identity.md` — identity/auth surface: що є, чого бракує.
- Агент B → `work/_wip-gap-audit-roadmap.md` — роадмап/бриф/backlog/limitations: planned-but-not-done.
Запущено: 2. Повернуто: (чекаю).

## Знахідки
Повернуто: A (identity) ✓. Чекаю B (роадмап).

### A — identity/auth (підтверджено файл:рядок, деталі в `_wip-gap-audit-identity.md`)
Реалізовано: Login/Authenticate/Logout (cookie HttpOnly SameSite=Lax, argon2id, CSRF RequireSameOrigin);
адмінка (список, блок/розблок, soft-delete, **адмінський reset пароля**, список інструкторів).

Вердикти:
- **НЕМА**: self-service реєстрація (користувачі лише з сіду `seed-identities.sql`, 6 рядків; нема CreateUser),
  верифікація email, self-service forgot/reset пароля, зміна власного пароля, зміна власного email,
  rate-limit/анти-брутфорс на логіні, **email/нотифікаційна інфраструктура взагалі** (grep порожній).
- **ЧАСТКОВО**: життєвий цикл сесії (lifetime жорстко 12h, нема remember-me / переліку-відкликання
  сесій користувачем, хоч SQL `DeleteIdentitySessionsForUser` існує й не викликається);
  призначення ролей (ролі є й перевіряються, але нема SetRole/Grant — інструктором стають лише правкою БД).
- **Є** (повний): адмінський reset пароля (НЕ шле нотифікацію — email-інфри нема).
- Міграції: нема колонок email_verified / verification_token / reset_token / remember-me / last_used.

Пріоритети A: 1) email-інфра (передумова 2–3); 2) self-service реєстрація; 3) верифікація email;
4) forgot/reset; 5) зміна пароля/email; 6) rate-limit на /auth/login; 7) призначення ролей.

### B — роадмап/бриф/backlog/limitations (деталі в `_wip-gap-audit-roadmap.md`)
Повернуто: B ✓. Запущено 2 / повернуто 2.
- **Роадмап §24 пройдено наскрізь.** Єдина незроблена продуктова фаза — **enrollment** (= K1): модель
  членства + перевірка доступу студента + порт playback-авторизації; оплата зроблена через «grant на курс»,
  членство відкладено бо не мало споживача. Наслідки, відкладені до enrollment: per-course блокування
  авторів, огляд навчання студентів у просторі інструктора.
- **Відсутнє-базове, ЩО НІКОЛИ НЕ БУЛО В ПЛАНІ:** email-верифікація, self-service відновлення пароля,
  email/нотифікації взагалі, самореєстрація/онбординг інструктора, сертифікати, глобальний пошук по
  каталогу, refund UI/API (refund після Approved ігнорується, B-26), друга мова контенту.
- **Прогалини базового = названі обмеження закритих фаз (не дефекти):** нема пагінації (transcript,
  discussion-відповіді), нема per-user ліміту публікацій/репортів (лише nginx zone=api за IP), throttle
  CHECK_STATUS у памʼяті (N× при N інстансах), grandfather лише за playback-прогресом.
- **Відкритий борг B-*:** B-27 (ryuk, тримати), B-1 (observability), B-22 (дубль селекторів), B-19 (360p
  VMAF — питання власнику), **B-20 (ASR draft не в openapi → рукописні zod, дрейф контракту, порушує
  AGENTS)**, B-21 (транскрипт бере алфавітний код мови, не мову уроку — латентна підміна), B-24 (payment
  polish), B-26 (refund-after-Approved gap).

## Зведення: прогалини згруповано
**Два незалежні потоки:**
1. **Enrollment (K1)** — фаза роадмапу, модель членства; НЕ залежить від email.
2. **Identity self-service кластер** (НЕ в роадмапі, нова scope) — усі впираються в **email-інфру як
   фундамент**: self-service реєстрація → верифікація email → forgot/reset пароля; + зміна власного
   пароля/email (зміна email теж потребує ре-верифікації). Rate-limit на логіні — суміжне, без email.

**Поза цими двома (названі, не прогавлені):** пошук, сертифікати, refund UI, нотифікації, друга мова —
усе поза планом; B-20/B-21 — борг, вартий окремого рішення.

## Рішення власника (2026-10-05)
1. **Порядок:** identity-кластер ПЕРШИЙ, enrollment (K1) — після.
2. **Email у тест-режимі:** Mailpit у Compose (dev-catcher, нічого назовні; e2e читає інбокс).
3. **Реєстрація:** публічний self-service signup (будь-хто → студент; роль інструктора ставить адмін).

## Наступний хід — фаза `accounts` через spec-flow
**Поріг: ВИЩЕ** (нова підсистема email + self-service auth, публічний API, обидва шари) → повний spec-flow.
- Гілка `feat/accounts-selfservice`, дім артефактів `docs/accounts/`.
- Scope фази: email-інфра (Mailpit) → публічна реєстрація + верифікація email → forgot/reset пароля →
  зміна власного пароля/email → rate-limit на /auth/login.
- Поза фазою (окремо, не зараз): enrollment (K1), B-20/B-21 борг, пошук/сертифікати/refund.
- Кроки: spec.md (ЩО/ЧОМУ) → GATE 1 grilling → plan.md → GATE 2 → tasks → implement (Codex) → review → report.

## Прогрес spec-flow (2026-10-05)
- Гілка `feat/accounts-selfservice` створена від main.
- **spec.md** готовий; **GATE 1 пройдено** — рішення G1–G12 (див. spec §Рішення GATE 1). Власник: «Go».
  Виправлення по ходу: Q2 переграно (мис-клік) → неверифікований read-only на КОНТЕНТІ (не грає уроки,
  лише вітрина); remember-me 3 дні (не 30).
- **plan.md** готовий: reuse-мапа (argon2id.Hash, SessionStore, DeleteIdentitySessionsForUser, Clock,
  contracts/access, platform/redis), нове (platform/mail+Mailpit, account_tokens, RateLimiter, UserWriter),
  гейт верифікації через `access.CanStream` + мапінг у 4 модулях, міграція 00014 + seed verified=true,
  конституція ✔, приймальна таблиця Acc01–Acc20.
- **GATE 2 запущено**: 2 gate-reviewer (4.8) — `_wip-accounts-gate2-critique.md`, `_wip-accounts-gate2-defense.md`.
  Запущено 2 / повернуто 2. **Обидва ЧЕРВОНІ.** Критика: B1 (не міняти CanBeViewedBy на CanStream —
  ламає access_mapping_ast_test + shared_rules TruthTable; verified-чек окремо в Play/transcript/мутації +
  розширити обидва тести), B2 (шов EmailVerified: domain.User→SessionStore.Resolve→httpserver.Actor→Actor
  модулів, не «кожен мапить свій факт»), B3 (change-password «крім поточної» — новий запит token_digest<>current).
  Non-blk: E1 seed у ОБИДВІ частини + аудит e2e; E2 назвати StartPurchase+Checkout, Publish+ReportComment+
  PutReview+ReportReview; E3 rate-limit ключ з поданого email (не оракул); E4 унікальність email на confirm.
  Фактичні: DeleteIdentitySessionsForUser викликається 3× (reset-all ок); validDisplayName ≤120 vs spec 100.
  → Закриваю B1–B3 + defense Q2–Q4 правкою плану, тоді re-review.
  - **defense = ЧЕРВОНИЙ.** Дірки: (1) SMTP-даун при signup — не прописано failure-path; (2) Redis down —
    fail-open vs fail-closed; (3) гонка signup → unique_violation(23505) змапити в `email_taken`, не 500;
    (4) зміна email на зайнятий чужий + TOCTOU; (5) верхні межі (пароль≤128, ім'я≤100) не в Acc;
    (6 non-blk) атомарний consume токена UPDATE…WHERE consumed_at IS NULL RETURNING. Non-blk: plan:12
    хибно каже, що `DeleteIdentitySessionsForUser` не викликається — викликається (administration.go:122/156/183).
    Закрито добре: повний e2e на крос-модульну зміну, міграція на даних (Acc20), мапінг AC→Acc (крім верх.меж).
