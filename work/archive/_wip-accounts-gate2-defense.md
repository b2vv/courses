claude-opus-4-8

# GATE 2 plan-defense — accounts

ДАЛІ: застосувати скіл plan-defense; прочитати docs/accounts/spec.md і docs/accounts/plan.md; збудувати таблицю загроз.

## ФАКТИ (звірка з кодом)

- Міграції: найвища наявна `backend/migrations/00013_ratings.sql`; `00014_accounts.sql` — коректний наступний номер. plan §5 вірно.
- Партковий unique email: `backend/migrations/00004_identity_admin.sql:15` — `CREATE UNIQUE INDEX users_email_active_key ON users (email) WHERE deleted_at IS NULL`. plan:14 вірно. Наслідок: паралельний signup на той самий email → один INSERT отримає 23505 (unique_violation).
- `contracts/access/access.go`: є `Subject{UserID, Administrator}`, `CanView`, `PurchaseRequired`. Немає `EmailVerified`, немає `CanStream`. plan §4 додає обидва — ок, факт підтверджено.
- Seed: `backend/sql/seed-identities.sql` — 6 рядків (UPSERT `ON CONFLICT (id) DO UPDATE`), наразі без `email_verified`. plan §5 додає `email_verified=true` у той самий UPSERT.
- `DeleteIdentitySessionsForUser`: plan:12 стверджує «уже є, не викликається» — НЕВІРНО. Викликається у `backend/internal/identity/adapters/postgres/administration.go:122,156,183` (admin reset/disable/delete). Запит існує і переюзабельний, але твердження «не викликається» хибне. Non-blocking для GATE 2, але неточність.

РІШЕННЯ: будую таблицю загроз + 5 фіксованих питань plan-defense.

## Таблиця загроз (що може піти не так → чим план закриває)

| Загроза | Закрито рядком плану/spec | Статус |
|---|---|---|
| SMTP (Mailpit) недоступний під час signup — що бачить користувач, чи акаунт консистентний | НЕМА рядка. G3/AC1 (автологін незалежно від листа) + Acc03 (resend) дають механізм відновлення, але план НЕ каже, що робить `Register` при відмові `Mailer` (fail-request? rollback user? success+no mail?), ні що бачить користувач. §3 описує лише happy-path. | **ДІРКА** |
| Redis (RateLimiter) недоступний — fail-open чи fail-closed | НЕМА рядка. plan §1/§5 згадують адаптер, але поведінку при відмові Redis не визначено (блокує всі входи / пропускає всі). | **ДІРКА** |
| Два паралельні signup на той самий email | Частково. G4/AC9/Acc10 закривають ПОСЛІДОВНИЙ випадок (pre-check exists→`email_taken`). Гонка (обидва проходять pre-check, один ловить 23505 на INSERT) НЕ описана — план не каже, що unique_violation мапиться в `email_taken`, а не 500. §6 лише перелічує код `email_taken`. | **ДІРКА** |
| Reset-токен вжито двічі / verify-лінк відкрито двічі (повторний submit, гонка) | Функційно закрито: AC3/AC5, §5 «Одноразовість: consumed_at IS NULL + expires_at > now()», Acc12. АЛЕ атомарність одночасного подвійного submit (UPDATE…WHERE consumed_at IS NULL RETURNING) не прописана — два конкурентні читання можуть обидва побачити NULL. | частково (функція так, конкурентність не явна) |
| Зміна email на вже зайнятий чужий email | НЕМА рядка. AC7/G7/Acc07 покривають лише вільну нову адресу. Зайнята нова адреса (та TOCTOU: вільна на запит, зайнята на confirm) не адресована; `email_taken` до email-change не прив'язано. | **ДІРКА** |
| Міграція 00014 на БД з даними (down/up без втрати; default=false + seed) | Закрито: §5 (ADD COLUMN NOT NULL DEFAULT false; seed ставить 6 сідів `true`; DOWN DROP TABLE+COLUMN), Acc20 «6 сідів verified; наявні e2e грають/мутують як раніше». Нюанс: залежить від повторного прогону seed після міграції (поза міграцією). Прод нема — втрата колонки на DOWN прийнятна. | закрито |
| Чи дійсно ганяється повний e2e (крос-модульна зміна contracts/access) | Закрито явно: §4 «зміна спільного контракту → повний e2e, не точковий (AGENTS §Lessons)», §5, §8 T9/T10 (впорядкований гейт). | закрито |
| Acc01–Acc20 покриває AC1–AC13 | Закрито: кожен AC має ≥1 рядок (мапінг нижче). Виняток у межах AC12: верхні межі (пароль ≤128, display_name ≤100) не тестуються жодним Acc. | частково |

### Мапінг AC→Acc
AC1→Acc01; AC2→Acc08+Acc09; AC3→Acc02+Acc03+Acc12; AC4→Acc04+Acc11; AC5→Acc05; AC6→Acc06+Acc13; AC7→Acc07; AC8→Acc14+Acc15; AC9→Acc10+Acc11; AC10→Acc16; AC11→Acc01/04+Acc18; AC12→Acc17 (лише нижні/порожні межі; **верхні 128/100 не покриті**); AC13→Acc19. Жоден AC не лишився без рядка.

## П'ять фіксованих питань plan-defense

| # | Питання | Закрито рядком | Вердикт |
|---|---|---|---|
| 1 | кривий інпут | `plan.md §2` (NewRegistration email/пароль≥8/ім'я1–100), `§6` (zod, коди `weak_password`), `§7 Acc17` | закрито (крім верхніх меж — див. Q4) |
| 2 | відмова залежності (SMTP/Redis) | — (нема рядка про поведінку при відмові Mailer чи Redis) | **ДІРКА** |
| 3 | гонка й повторний виклик (signup-гонка, атомарний consume токена) | частково: `§5` consumed_at для одноразовості, `AC9` pre-check; але insert-гонка→`email_taken` і атомарний UPDATE токена не прописані | **ДІРКА** |
| 4 | порожні та граничні дані (верхня межа пароль≤128, ім'я≤100; зайнятий email у change) | частково: `Acc17` лише нижні/порожні межі; верхні межі й email-change-to-taken без рядка | **ДІРКА** |
| 5 | rollback | `plan.md §5 Rollback` (down чиста, git-revert FF-гілки до мерджу в main, видалення mailpit-сервісу) | закрито |

## РІШЕННЯ / ВЕРДИКТ

Незакритих рядків: **4** (Q2 відмова залежності; Q3 гонка/повтор; Q4 граничні+зміна email на зайнятий; плюс окремий рядок таблиці загроз «зміна email на зайнятий», що й тягне Q4). Гейт тримають рядки Q2, Q3, Q4.

ВЕРДИКТ: **ЧЕРВОНИЙ**.

Незакриті (блокуючі) діри:
1. Mailer-відмова під час signup — не визначено поведінку й стан акаунта (Q2, таблиця загроз р.1).
2. Redis/RateLimiter-відмова — fail-open vs fail-closed не визначено (Q2, р.2).
3. Паралельний signup на той самий email — unique_violation(23505) на INSERT не змаплено в `email_taken` (ризик 500) (Q3, р.3).
4. Атомарність подвійного consume verify/reset-токена не прописана (Q3, р.4) — non-blocking за бажанням, але названо.
5. Зміна email на вже зайнятий чужий email (+TOCTOU на confirm) — без AC/коду (Q4, р.5).
6. Верхні межі валідації (пароль ≤128, display_name ≤100) не покриті жодним Acc (Q4, в межах AC12).

Non-blocking неточність: plan:12 «DeleteIdentitySessionsForUser … не викликається» — насправді викликається (administration.go:122/156/183).
