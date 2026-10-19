# Codex brief — A08a: auth-екрани (register / verify / forgot / reset) + login remember-me + i18n

Репо `courses`, гілка `feat/accounts-selfservice`. A07 у дереві: `SessionRepository` += self-service методи,
application session-команди, `identity.refresh`. Наявний `routes/login.tsx` + `presentation/identity/LoginView`.
Роль: реалізуєш A08a; коміт/гейти — Courses 1.

РОБОЧИЙ ФАЙЛ: /Users/strelia/projects/courses/work/_wip-a08a-codex.md
Append після КОЖНОГО кроку; перший рядок `ДАЛІ:`. Фінал = стислий переказ.

## Екрани (TanStack Start routes + presentation, патерн наявного login.tsx/LoginView)
1. **`/register`**: форма email + пароль + ім'я (React Aria). Submit → application register → автологін →
   редірект на головну/навчання; показати ненав'язливий notice «ми надіслали лист, підтвердіть пошту».
   Помилки: email_taken (прямо «вже зареєстрований» + лінки вхід/скидання), weak_password, валідація.
2. **`/verify-email`**: читає `token` із search-param (Router search params, НЕ location.search вручну);
   викликає verify; стан success / expired-or-used (+ кнопка «надіслати знову» через resend); після успіху
   `identity.refresh`.
3. **`/password/forgot`**: email → forgot → **нейтральне** підтвердження «якщо акаунт існує, лист надіслано»
   (не розкриває існування).
4. **`/password/reset`**: `token` із search-param + новий пароль → reset → редірект на `/login` з нотісом.
5. **`/login`**: додай чекбокс «Запам'ятати мене» (React Aria Checkbox) → `rememberMe` у logIn.

## Правила (AGENTS §TypeScript, design)
- React Aria компоненти (reuse `presentation/ui/`), **НЕ** нові UI-залежності. Семантичні дизайн-токени.
- URL-стан — Router search params; server-state — Query (presentation НЕ fetch-ить напряму — через application).
- **i18n uk/en** усіх рядків (ключі у `app/providers/locales`, патерн наявних). 
- a11y: клавіатура, видимий фокус, контролі ≥44px, світла/темна, 200% zoom. Форми — label/помилки з aria.
- Strict TS; валідація вводу перед сабмітом; без `any`.
- НЕ комість; generated не чіпай; orval не запускай. Час вичерпується — спершу тести вже-написаного.

## Критерій готовності A08a
- `cd frontend && npm run typecheck`, `npm run lint`, `npm run test` (rstest) зелені — нові view-тести
  (happy + ключові помилки: register email_taken, verify expired+resend, forgot neutral, reset, remember-me).
- Нові маршрути зареєстровані у роутері; i18n uk/en повні для нових екранів; жодного fetch у presentation.
- (Settings/account, банер верифікації й play-гейт — це A08b, НЕ тут.)
