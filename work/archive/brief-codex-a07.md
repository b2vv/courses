# Codex brief — A07: фронт repository + application + zod (auth self-service)

Репо `courses`, гілка `feat/accounts-selfservice`. Backend A01–A06 готовий: `/api/v1/auth/*` живі, Orval-клієнт
згенеровано (`frontend/src/adapters/api/generated` має register/verify/resend/password.*/email.*/login rememberMe,
`AuthenticatedUser.emailVerified`). Роль: реалізуєш A07; коміт/гейти — Courses 1.

РОБОЧИЙ ФАЙЛ: /Users/strelia/projects/courses/work/_wip-a07-codex.md
Append після КОЖНОГО кроку; перший рядок `ДАЛІ:`. Фінал = стислий переказ.

## Контекст архітектури фронту (AGENTS §TypeScript, reuse-index)
- Порт `frontend/src/ports/SessionRepository.ts` (logIn/currentUser/logOut). Доменні типи — `src/domain/identity`.
- HTTP лише через `src/adapters/api/transport.ts` (дедлайн/скасування/класифікація) + `request-headers.ts`.
- Валідація зовнішніх даних — zod (strict; без `any`/небезпечних as). Server-state — Query (presentation не fetch-ить).
- Патерн наявного репозиторію — `src/adapters/api/*repository*.ts` (напр. course-repository); наслідуй його.

## Що зробити (НЕ UI — це A08)
1. **Розший `SessionRepository`** (порт) методами: `register`, `resendVerification`, `forgotPassword`,
   `resetPassword`, `changePassword`, `requestEmailChange`, `confirmEmailChange`. `logIn` — додай підтримку
   `rememberMe`. `currentUser` результат несе `emailVerified`.
2. **Реалізація репозиторію** (adapters/api) на Orval-клієнті + `transport.ts`: мапінг бізнес-кодів помилок
   (email_taken, email_not_verified, weak_password, token used/expired, invalid_credentials, rate-limited) у
   типобезпечні доменні результати (як наявний репозиторій класифікує). **zod-схеми відповідей** — валідуй усе,
   що приходить ззовні; forgot — успіх-нейтраль.
3. **Application шар** (`src/application/...`, патерн наявного): use-cases/команди, які presentation (A08) викличе
   через Query — register, verify, resend, forgot, reset, changePassword, requestEmailChange, confirmEmailChange;
   інвалідація сесійного кеша де треба (напр. після verify — оновити me).
4. **Доменні типи** (`src/domain/identity`) для вводу/результатів нових операцій, без транспорту.

## Правила
- Strict TS, zod на межі, без fetch у presentation (A08 не чіпай тут), Query для server-state, React Aria — лише A08.
- Жодних нових залежностей без потреби (QR-бібліотека — A15, не тут).
- Дотримуйся напряму залежностей (перевіряє `frontend/scripts/check-architecture.mjs`).
- НЕ комість; generated вже згенеровано (НЕ запускай orval, НЕ редагуй generated). Час вичерпується — тести вже-написаного.

## Критерій готовності
- `cd frontend && npm run typecheck` і `npm run lint` (oxlint + architecture) чисті.
- `npm run test` (rstest) зелені: нові unit-тести репозиторію (мапінг кодів, zod-парсинг, forgot-нейтраль,
  rememberMe) і application use-cases.
- Жодного fetch у presentation; напрям залежностей не порушено.
