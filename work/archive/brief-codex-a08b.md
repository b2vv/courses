# Codex brief — A08b: settings/account + банер верифікації + play-гейт UI

Репо `courses`, гілка `feat/accounts-selfservice`. A07/A08a у дереві: application session-команди
(changePassword/requestEmailChange/confirmEmailChange/resendVerification), `identity.refresh`, auth-екрани,
`currentUser.emailVerified`. Роль: реалізуєш A08b; коміт/гейти — Courses 1.

РОБОЧИЙ ФАЙЛ: /Users/strelia/projects/courses/work/_wip-a08b-codex.md
Append після КОЖНОГО кроку; перший рядок `ДАЛІ:`. Фінал = стислий переказ.

## Що зробити (React Aria, i18n uk/en, a11y — як A08a)

1. **`/settings/account`** (авторизований route + view):
   - **Зміна пароля**: поля старий + новий; submit → application changePassword; стани success/помилки
     (невірний старий, weak_password). Після успіху — повідомлення (інші сесії завершено).
   - **Зміна email**: поле новий email; submit → requestEmailChange; notice «підтвердіть на новій адресі».
   - (Підтвердження email-change — через окремий лінк, який веде на той самий verify-потік або окремий
     confirm-екран; якщо потрібен екран `/email/change/confirm` з token-search-param — додай його аналогічно
     verify-email з A08a.)
   - Guard: лише для залогінених (редірект на /login як наявні захищені маршрути).

2. **Банер верифікації**: компонент, що показується, коли `currentUser.emailVerified === false`, на
   авторизованих сторінках (інтегруй у наявний app-shell/header — знайди, де рендериться IdentityHeader/layout).
   Кнопка «надіслати лист знову» → resendVerification. Доступний, не нав'язливий, dismiss у межах сесії
   (sessionStorage) дозволений, але після перезавантаження показується знову якщо досі не verified.

3. **Play-гейт UI**: там, де рендериться контрол відтворення уроку (знайди LessonPlayer / play-кнопку в
   learn), якщо `emailVerified === false` — показати замість/поверх плеєра зрозумілий стан «підтвердіть пошту,
   щоб дивитися» + дія resend/лінк, а не зламаний плеєр. Бекенд усе одно відмовляє (403 email_not_verified);
   UI лише відображає це завчасно. Опис курсу/план лишаються видимими.

## Правила
- React Aria, семантичні токени, i18n uk/en, a11y (клавіатура, фокус, ≥44px, 200% zoom, світла/темна).
- Server-state через Query/application; presentation не fetch-ить. URL-стан — search params.
- Strict TS; без нових залежностей. Напрям залежностей (`check-architecture.mjs`) не порушувати.
- НЕ комість; generated не чіпай; orval не запускай (якщо додаси роут — `npm run build` сам оновить
  routeTree.gen.ts, це ок). Час вичерпується — спершу тести вже-написаного.

## Критерій готовності
- `cd frontend && npm run typecheck`, `npm run lint`, `npm run test` зелені; нові view-тести (settings:
  change-password happy+wrong-old, change-email; банер show/hide + resend; play-гейт для unverified vs verified).
- Банер і play-гейт реагують на `emailVerified`; описова сторінка курсу лишається доступною неверифікованому.
