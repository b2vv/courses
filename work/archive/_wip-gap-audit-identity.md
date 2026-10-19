# Gap audit: identity / auth

ДАЛІ: інвентаризувати backend/internal/identity/** — структуру тек, application use-cases, handlers, domain, порти, postgres-адаптери; тоді openapi, frontend, міграції.

## ФАКТ: application use-cases
- Identity (application/identity.go): ПУБЛІЧНІ методи лише Login(45), Authenticate(71), Logout(90). НЕ існує Register/Signup, ChangePassword(self), ChangeEmail, ForgotPassword/Reset, VerifyEmail.
- sessionLifetime = 12h фіксовано (identity.go:13). Немає "remember me" (LoginInput: Email/Password/PresentedToken лише, :29). PresentedToken — ротація наявної сесії при логіні.
- Logout → sessions.Revoke (відкликання однієї сесії за токеном). Authenticate → Resolve(token, now) перевіряє активність.
- Administration (application/administration.go): ListUsers(34), SetStatus(54, active/blocked), Delete(76), ResetPassword(92 — адмінський, по targetID+password, домен ValidPassword). Немає email-нотифікацій.
- Ролі domain/user.go: Student/Instructor/Admin (13-17). CanAccessAdministration = Instructor|Admin; CanAdministerAccounts = Admin. Немає use-case призначення/зміни ролі.

ДАЛІ: handler.go, admin_handler.go, ports (authentication.go, administration.go), domain/administration.go, errors.go

## ФАКТ: HTTP routes (handler.go:68-77)
POST /api/v1/auth/login; GET /api/v1/auth/me; POST /api/v1/auth/logout; GET /api/v1/admin/instructors; GET /api/v1/admin/users; POST /api/v1/admin/users/{id}/status; DELETE /api/v1/admin/users/{id}; POST /api/v1/admin/users/{id}/password.
- НЕМА маршрутів: register/signup, change-password(self), change-email, forgot/reset, verify-email.
- Cookie: HttpOnly, SameSite=Lax, Secure(config), MaxAge = session lifetime (handler.go:186-187). Logout відкликає ВСІ cookie-сесії (198-206).
- RequireSameOrigin (142) — CSRF-захист origin-перевіркою на mutating routes. Немає rate-limit middleware у ланцюгу.

## ФАКТ: міграції
- 00002_identity.sql: users(id,email UNIQUE,display_name,role CHECK student/instructor/admin,password_hash argon2id,created_at,updated_at,disabled_at). user_sessions(id,user_id,token_digest 32B UNIQUE,created_at,expires_at). НЕМА колонок email_verified, verification_token, reset_token, remember-me, last_used, revoked.
- 00004_identity_admin.sql: + deleted_at (soft delete) + частковий unique email WHERE deleted_at IS NULL. Токенів верифікації/скидання НЕМА в жодній міграції.

## ФАКТ: rate-limit / email / ролі
- Rate-limit / brute-force: ВІДСУТНІЙ. Єдиний "attempts" у identity — repository.go:22 rotationAttempts=3 (ретрай транзакції ротації, не захист логіну).
- Email/SMTP/mailer/notification інфраструктура: grep порожній — ВІДСУТНЯ повністю.
- Призначення ролей: НЕМА use-case SetRole/AssignRole/Grant. ListInstructors (repo:161, SQL role IN instructor,admin) лише читає. Роль задається тільки в БД/сіді.

ДАЛІ: repository.go (Rotate/Resolve/Revoke, чи є Create), seed/cmd, openapi auth paths, frontend auth routes

## ФАКТ: створення користувачів і frontend
- Створення акаунтів: ТІЛЬКИ через фіксований сід backend/sql/seed-identities.sql:4 (6 рядків: student/instructor/admin + block-me/delete-me/reset-me). Немає INSERT INTO users у проді (repository.go/generated: лише GetIdentityCredential, Create/Resolve/Delete session, SetUserDisabledAt, SetUserPasswordHash, SoftDeleteUser — жодного CreateUser).
- Frontend: єдиний auth-маршрут routes/login.tsx. SessionRepository (ports/SessionRepository.ts) = logIn/currentUser/logOut — три методи. Жодних екранів register/forgot/verify/change-password(self). ResetPasswordDialog.tsx — адмінський.
- openapi.yaml auth-секція: лише /api/v1/auth/login(635), /me(666), /logout(685). Admin: /admin/users(196), /users/{id}(233 DELETE), /status(259), /password(290 admin-reset), /instructors(612).

## ПІДСУМКОВА ТАБЛИЦЯ: можливість → статус → доказ → що бракує
| Можливість | Статус | Доказ | Що бракує для повноти |
|---|---|---|---|
| Self-service реєстрація (signup) | НЕМА | немає маршруту/use-case; користувачі лише з seed-identities.sql:4; жодного CreateUser у repository.go/generated | endpoint POST register, use-case, UI, валідація унікальності email (БД-констрейнт уже є) |
| Верифікація email | НЕМА | users-таблиця без email_verified/token (00002,00004); жодного verify-маршруту; email-інфра відсутня | колонка verified+token, маршрути send/confirm, email-доставка |
| Self-service скидання пароля (forgot) | НЕМА | немає reset-token колонок/маршрутів; email-інфра відсутня (grep порожній) | reset-token store, forgot/confirm endpoints, email-доставка, UI |
| Зміна власного пароля (авторизований) | НЕМА | Identity має лише Login/Authenticate/Logout (identity.go:45,71,90); SetUserPasswordHash викликає лише адмін-шлях | change-password use-case+endpoint (з перевіркою старого пароля), UI |
| Зміна власного email | НЕМА | немає use-case/маршруту/колонок; mapUser лише читає email | change-email flow (часто з ре-верифікацією), endpoint, UI |
| Адмінський reset пароля | Є | application/administration.go:92 ResetPassword; handler admin_handler.go:111 POST /admin/users/{id}/password; ValidPassword; UI ResetPasswordDialog.tsx | — (повне; без email-нотифікації користувача) |
| Життєвий цикл сесії | ЧАСТКОВО | lifetime 12h фіксовано identity.go:13; expires_at CHECK (00002); Resolve перевіряє протермінування; Logout=Revoke усіх cookie (handler.go:198-206); DeleteExpiredIdentitySession існує | "запам'ятати мене" (НЕМА, LoginInput:29); налаштовувана тривалість; масове відкликання/перелік сесій користувача (DeleteIdentitySessionsForUser є у SQL, але не викликається з use-case) |
| Ролі й призначення | ЧАСТКОВО | ролі student/instructor/admin (user.go:13); перевірки CanAdminister (user.go:31); ListInstructors читає | НЕМА use-case/endpoint призначення ролі — інструктором стають лише правкою БД/сіду; немає self-serve чи admin grant-role |
| Rate-limit / анти-брутфорс на логіні | НЕМА | жодного rate/throttle middleware у ланцюгу (handler.go:68-77); єдині "attempts" — rotationAttempts=3 (ретрай tx, repository.go:22) | rate-limit на /auth/login (per-IP/per-email), lockout/backoff |
| Email/нотифікаційна інфраструктура | НЕМА | grep smtp/mailer/notification/send-email порожній по backend | SMTP/черга листів — передумова для верифікації та forgot-reset |

РІШЕННЯ: аудит завершено, усе підтверджено файлом:рядком.
