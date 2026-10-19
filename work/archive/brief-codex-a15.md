# Codex brief — A15: 2FA фронт (enroll + login 2-й крок + disable/regenerate)

Репо `courses`, гілка `feat/accounts-selfservice`. A14 дав HTTP 2FA + Orval-клієнт (enroll/confirm/complete/
disable/recovery-codes; login→`two_factor_required`+pendingId). A07/A08 дали SessionRepository, application,
settings/account, LoginView. Роль: реалізуєш A15; коміт/гейти — Courses 1.

РОБОЧИЙ ФАЙЛ: /Users/strelia/projects/courses/work/_wip-a15-codex.md
Append після КОЖНОГО кроку; перший рядок `ДАЛІ:`. Фінал = стислий переказ.

## Що зробити (React Aria, i18n uk/en, a11y — як A08)
1. **SessionRepository/application** (розширити A07): методи `enrollTwoFactor`, `confirmTwoFactor(code)`,
   `completeTwoFactor(pendingId, code)`, `disableTwoFactor(password, code)`, `regenerateRecoveryCodes(code?)`.
   **`logIn` тепер може повернути `two_factor_required`+pendingId** (login oneOf) — відобрази це в типобезпечному
   результаті (не помилка, а стан «потрібен 2-й крок»); zod-валідація відповідей.
2. **Логін — другий крок**: після пароля, якщо `two_factor_required` → показати форму коду (6 цифр TOTP) з
   опцією «ввести резервний код» → `completeTwoFactor` → сесія → редірект. Невірний код/вичерпані спроби — стани.
3. **Налаштування /settings/account — секція 2FA**:
   - Якщо вимкнено: кнопка «Увімкнути 2FA» → enroll → показати **QR** (з otpauthUri) + секрет текстом →
     поле коду → confirm → показати **резервні коди один раз** (копіювати/завантажити).
   - Якщо увімкнено: «Вимкнути» (пароль+код) і «Перегенерувати резервні коди» (повертає нові один раз).
4. **QR**: додай МАЛУ QR-бібліотеку (напр. `qrcode`/`qrcode.react` або рендер SVG) — нова залежність, що
   **використовується** (AGENTS дозволяє). Рендер з otpauthUri; доступний (alt/опис + секрет текстом як фолбек).
5. **i18n uk/en** усіх нових рядків; a11y (клавіатура, фокус, ≥44px, zoom, світла/темна).

## Правила
- Strict TS, zod на межі, без fetch у presentation (через application/Query), URL-стан — search params.
- Напрям залежностей (`check-architecture.mjs`) не порушувати; нова залежність — лише QR і лише там, де треба.
- НЕ комість; generated/openapi не чіпай; orval не запускай. Час — спершу тести вже-написаного.

## Критерій готовності
- `cd frontend && npm run typecheck`, `npm run lint`, `npm run test` зелені; нові view/application-тести
  (login 2-й крок happy + невірний код + резервний; enroll→confirm→recovery codes; disable; regenerate).
- `npm install` нової QR-залежності відображено в package.json/lock; збірка (`npm run build`) зелена.
- 2FA-секція реагує на стан enabled/disabled; логін коректно веде через 2-й крок.
