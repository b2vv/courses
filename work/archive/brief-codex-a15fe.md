# Codex brief — A15-fe: дофінішити 2FA фронт (settings секція + QR + тести)

Гілка feat/accounts-selfservice. Базовий 2FA seam УЖЕ в дереві (твій A15-перший прохід): login другий крок,
zod login oneOf/enrollment/recovery, repository/application команди enroll/confirm/complete/disable/regenerate.
Тепер розблоковано: `/auth/me` (і `AuthenticatedUser`) НЕСЕ `twoFactorEnabled` (Orval перегенеровано);
`qrcode.react@^4.2.0` встановлено у package.json.
РОБОЧИЙ ФАЙЛ: work/_wip-a15fe-codex.md (append, перший рядок `ДАЛІ:`).

Дофінішуй:
1. **Секція 2FA у `/settings/account`** (React Aria, i18n uk/en, a11y):
   - Стан бери з `currentUser.twoFactorEnabled` (через application/Query; presentation не fetch-ить).
   - Вимкнено → «Увімкнути 2FA»: enroll → показати **QR** (qrcode.react з otpauthUri) + секрет текстом (фолбек) →
     поле коду → confirm → показати **резервні коди ОДИН РАЗ** (копіювати).
   - Увімкнено → «Вимкнути» (пароль+код) і «Перегенерувати резервні коди» (нові один раз).
2. **View/application-тести**: enroll→confirm→recovery codes; login другий крок happy+невірний+резервний (якщо
   базовий seam ще без тесту — додай); disable; regenerate; відображення enabled/disabled за twoFactorEnabled.
3. i18n uk/en усіх нових рядків.

Правила: strict TS, zod на межі, напрям залежностей не порушувати, generated/openapi не чіпай, orval не запускай.
НЕ комість. Час — спершу тести вже-написаного.
Критерій: `npm run typecheck`, `npm run lint`, `npm run test`, `npm run build` зелені; 2FA-секція реагує на стан.
