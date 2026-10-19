# Codex brief — A09b: 2FA e2e сценарії у accounts.spec.ts (Acc26–Acc31)

Гілка feat/accounts-selfservice. `frontend/e2e/accounts.spec.ts` уже покриває Acc01/02/04–08/10/16/19.
2FA повністю реалізовано (backend+frontend): /settings/account секція (enroll QR+секрет→confirm→recovery codes),
login 2-й крок (two_factor_required→код/резервний), disable, regenerate; admin disable 2FA.
РОБОЧИЙ ФАЙЛ: work/_wip-a09b-codex.md (append, перший рядок `ДАЛІ:`).

Додай у accounts.spec.ts (патерн наявних сценаріїв, unique emails на прогін, Mailpit де треба, детермінованість):
- **Acc26**: enroll у /settings/account → показано QR+секрет → ввести валідний TOTP (згенеруй код із секрета в
  тесті — RFC6238, як бекенд; візьми секрет з UI/otpauth) → 2FA увімкнено → recovery codes показані один раз.
- **Acc27**: вийти → увійти (пароль) → з'являється 2-й крок → ввести валідний TOTP → сесія; невірний код — відмова.
- **Acc28**: ще раз login → 2-й крок → ввести **резервний код** (з Acc26) → успіх; той самий код удруге — відмова.
- **Acc29**: disable 2FA (пароль + код) → 2FA вимкнено; окремо перевір admin-disable (адмін вимикає 2FA юзеру),
  якщо практично в e2e — інакше познач як покрите integration/HTTP, не дублюй.
- **Acc31**: regenerate recovery codes → старі недійсні (спроба старим на login 2-му кроці — відмова).
- axe на 2FA-екранах (enroll + login 2-й крок).

Правила: TOTP-код рахуй у тесті (hmac-sha1/base32, 30s, 6 цифр) — без зовнішніх залежностей, або мінімальний
helper. Детермінованість (fake-timers не треба; бери поточний крок). Не ламай наявні сценарії; прибирання в afterAll.
НЕ комість; generated/openapi не чіпай. Браузерний прогін лишаєш Courses 1 (A10 у Compose).

Критерій: accounts.spec.ts компілюється (frontend typecheck+lint чисті); нові сценарії покривають Acc26–Acc29, Acc31.
