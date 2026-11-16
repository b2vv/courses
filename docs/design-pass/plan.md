# design-pass — усі екрани фронтенду

| | |
|---|---|
| Дата | 2026-11-14 |
| Ініціатор | власник (через сесію-оркестратор `local_35696fcf…`), окрема задача поза §24 |
| Гілка | `claude/design-pass-all-screens-085f0e` (worktree `dg-archive-test-c6c9f4`, база `main` @ `37ec4b0` після rebase 2026-11-16) |
| Метод | скіл `design-pass` (`~/.claude/skills/design-pass/SKILL.md`) + `references/viewport-coverage.md`; фінал кожного зрізу — `interface-review` |
| Статус задачі | **S0 in_progress** (Q1, Q2 погоджено 2026-11-16, §9). Q3, Q4 відкриті. Код не змінено. |

## 1. Обмеження проєкту (design-pass §1)

Записано з файлів, не з пам'яті. Нічого з цього зріз не перевизначає.

| Вісь | Джерело | Що закрито |
|---|---|---|
| Стек | `AGENTS.md` §TypeScript | TanStack Start + Rsbuild, React Aria, Query. Без Vite/Next/Vitest/Jest/ESLint/Tailwind/Radix/shadcn. Нових залежностей не додаємо. |
| Токени | `frontend/src/app/styles/tokens.css` (1046 р.: палітра, типографіка, радіуси, тіні, light/dark, `forced-colors`, `prefers-reduced-motion`) | Глобальні токени не переписуються мовчки; зміна токена = shared-зміна → повний e2e. |
| Щільність client | `tokens.client.css` (`[data-surface='client']`) | Лише density-ручки; нових кольорів/шрифтів немає за задумом файла. |
| Щільність + палітра admin | `tokens.admin.css` (`[data-surface='admin']`, admin **і** instructor `/teach`) | **Dark-only** — рішення власника 2026-10-09 (`tokens.admin.css:10`). Light/dark для `/admin`, `/teach` = **N/A** з цим посиланням. Brand `#5273f2`. |
| Тема client | `tokens.css:77-98` — `prefers-color-scheme` + `[data-theme]` override; перемикача в UI немає | Light і dark перевіряються емуляцією `prefers-color-scheme`. Додавати перемикач — поза скоупом. |
| Примітиви | `frontend/src/presentation/ui/` (`controls.tsx`, `data.tsx`, `layout.tsx`, `overlays.tsx`, `states.tsx`, `StatTile`, `TrendChart`), стилі в `main.css` (3585 р.) | Перевикористовуємо; admin-специфіка не тягнеться в глобальний примітив (gotcha 2026-10-14, S2b/S3/S4). |
| A11y-мінімум | `AGENTS.md` | Клавіатура, видимий фокус, контроли ≥44px, light/dark, 200% тексту. |
| Попередні дизайн-рішення | `docs/design/portal-redesign.md`, `docs/design/northwind-admin-spec.md`, `docs/admin-restyle/` | Стиль уже обраний; цей прохід **вирівнює й виправляє**, не вигадує новий стиль (design-pass §2: «do not invent a new style for a targeted change»). |

## 2. Бриф (design-pass §2)

- Тип артефакту: `system` (≈25 поверхонь трьох поверхонь-щільностей: storefront/client, learn, admin/teach).
- Мета: кожен екран має зрозумілу ієрархію й основну дію, повний основний сценарій і **докази** на 4 viewport + 200% тексту; blocker/major виправлено й перевірено повторно.
- **Non-goals:** редизайн або новий стиль; нові фічі/стани, яких немає; зміна бізнес-поведінки, API, бекенду; перемикач теми; production-деплой.
- Acceptance-база: чинні acceptance фаз, які володіють екранами (catalog, accounts, playback/learning, admin, teach-editor, multi-audio…). Design-pass їх **не змінює**; UI-перевірки нижче — запропоновані доповнення, не прийняті до погодження.

## 3. Інвентар екранів (звірено з `frontend/src/routes/` @ `ef48c15`)

Список із брифу актуальний; додано 4 поверхні-оболонки, яких у ньому не було (позначено ➕).

| # | Поверхня | Маршрут(и) | Роль / фікстура | Тема | Стани для перевірки (уточнюються на старті зрізу) |
|---|---|---|---|---|---|
| G0 | ➕ Оболонка client: хедер/навігація/футер, 404, error | `__root.tsx` (`NotFound`, `PageError`) | гість / студент | light+dark | nav open/closed на phone, фокус, 404, error |
| C1 | Головна | `index` | гість | light+dark | дані, довгі назви |
| C2 | Каталог | `courses.index` | гість | light+dark | loading, дані, порожньо, error |
| C3 | Сторінка курсу | `courses.$slug` | гість / студент / власник | light+dark | free/paid, вже записаний, not-found, рейтинги |
| A1 | Логін | `login` | гість | light+dark | помилка, 2FA-крок, pending |
| A2 | Реєстрація | `register` | гість | light+dark | валідація, успіх |
| A3 | Підтвердження email | `verify-email` | гість | light+dark | валідний/прострочений токен |
| A4 | Забули пароль | `password.forgot` | гість | light+dark | надіслано |
| A5 | Скидання пароля | `password.reset` | гість | light+dark | помилка токена, успіх |
| A6 | Підтвердження зміни email | `email.change.confirm` | студент | light+dark | успіх/помилка |
| A7 | Налаштування акаунта | `settings.account` | студент | light+dark | пароль, email, TOTP enroll/disable |
| P1 | Замовлення | `orders.$orderId` | студент | light+dark | pending/approved/declined |
| L1 | ➕ Навчання — оболонка / огляд | `learn` (`LearnEmptyState`, outline) | студент | light+dark | порожньо, курси, прогрес |
| L2 | Урок: плеєр, транскрипт, документ, обговорення, аудіодоріжки, відгук | `learn.$` | студент | light+dark | locked, email-gate, video/document, довгий транскрипт, гілки |
| D0 | ➕ Оболонка admin (`AdminShell`, nav, pending/failure, `admin.$` catch-all) | `admin`, `admin.$` | адмін | dark-only | nav на phone, catch-all |
| D1 | Список курсів | `admin.index` | адмін | dark-only | loading, порожньо, дані, error |
| D2 | Редактор курсу (вкладки; «Уроки»: відео + аудіодоріжки) | `admin.courses.$courseId` | адмін | dark-only | кожна вкладка, upload progress, awaiting_audio |
| D3 | Редактор документа уроку | `admin.courses.$courseId_.lessons.$lessonId` | адмін | dark-only | порожній draft, модулі, publish |
| D4 | Користувачі | `admin.users.index` | адмін | dark-only | фільтри, сортування, порожньо |
| D5 | Аналітика | `admin.analytics` | адмін | dark-only | pending, failure, графіки |
| D6 | Учні | `admin.learners` | адмін | dark-only | таблиця, порожньо |
| D7 | Скарги на учасників | `admin.participant-reports` | адмін | dark-only | черга, дія, порожньо |
| D8 | Повернення коштів | `admin.refunds` | адмін | dark-only | черга, дія, порожньо |
| T0 | ➕ Оболонка teach | `teach` | інструктор | dark-only | nav, pending/failure |
| T1 | Курс інструктора | `teach.courses.$courseId` | інструктор | dark-only | 0 уроків, список |
| T2 | Редактор документа (teach) | `teach.courses.$courseId_.lessons.$lessonId` | інструктор | dark-only | як D3 у межі `/teach` |
| T3 | Учні інструктора | `teach.learners` | інструктор | dark-only | таблиця, block, report |

Фікстури — наявні `backend/sql/seed-identities.sql` + `seed.sql` та e2e setup; нові дані в live-БД не сіємо (design-pass §5). Стан, недосяжний наявними фікстурами, — NOT_TESTED з причиною, не вигаданий.

## 4. Матриця доказів (viewport-contract)

На кожну поверхню × стан: Phone 390×844, Tablet-P 768×1024, Tablet-L 1024×768, Desktop 1440×900 (100% zoom), окремо **200% тексту** (метод: CSS `font-size:200%` на `:root` — це симуляція, так і записується) і **reflow 320 CSS px**; client — у light і dark. Рядок доказу: маршрут/стан/роль, revision, CSS-розмір, метод, очікуване/фактичне, шлях скріншота, «переглянуто: так». Статуси PASS / FAIL / NOT_TESTED / N/A / BLOCKED.

Доказ живе в `docs/design-pass/evidence/<id>.md` + `docs/design-pass/screens/<id>/…png` (лише для фінальної ревізії; проміжні — у scratchpad, не в git).

Інструмент: Playwright із `frontend/node_modules` (уже встановлений, без нового стеку) — скрипт у scratchpad знімає 4 viewport × теми детерміновано; інтерактивні стани/клавіатура — вбудований браузер на `localhost`. Не через `host.docker.internal` (gotcha 2026-10-16: `origin_forbidden`). Axe — наявний `e2e/accessibility.spec.ts`, повторно не дублюю.

## 5. Порядок проходу

1. **S0 — базовий огляд (без коду).** Знімаю всі 27 поверхонь у поточному стані → `docs/design-pass/findings.md`: знахідка, екран, viewport, severity (blocker/major/minor), доказ. Мета — побачити спільні дефекти до того, як правити по екрану, щоб не лагодити один примітив 10 разів. Власник бачить реєстр і може переставити пріоритети.
2. **S1 — спільні примітиви/токени** (лише якщо S0 знайде спільний blocker/major): окремий зріз, повний e2e, перевірка client **і** admin щільності.
3. **Поекранні зрізи** у порядку цінності для студента → автора:
   G0 → C1 → C2 → C3 → A1…A7 → P1 → L1 → L2 → D0 → D1 → D2 → D3 → D4…D8 → T0 → T1 → T2 → T3.
   Кожен зріз: стани з §3 → план композиції/ієрархії → фікс blocker/major → повторна перевірка тих самих viewport/станів → `interface-review` → рядок у `evidence/<id>.md`.
   Екран без blocker/major закривається доказами без коду (теж зріз, коміт лише evidence).
4. Minor-знахідки — реєстр у `findings.md`; виправляються лише якщо дешеві й у межах того самого файла, інакше лишаються видимими, не «прийнятими».

## 6. Верифікація й рев'ю

- Кожен зріз з кодом: `make format-check → lint → typecheck → test → integration → build → e2e` (CSS у `main.css`/примітивах = shared → **повний** e2e, `AGENTS.md`). `npm run format` перед комітом (Codex/oxfmt gotcha).
- Рев'ю: `review_mode=independent_subagent` (окремий runtime — `codex exec … < /dev/null`, не fork/inline) перед кожним мерджем; або Courses 2, якщо власник призначить.
- Звіт: `docs/design-pass/report.md` (ship-report, PASS/FAIL/NOT_TESTED/N/A).

## 7. Pre-code checkpoint (`docs/development/ai-factory.md` §Перед кодом)

| Поле | Стан |
|---|---|
| Scope / acceptance | §2–§3; full (≈27 поверхонь, багато зрізів). Owner — власник. Acceptance фаз-власників не змінюється. |
| Grilling / рішення | Відкриті питання — §8 (Q1–Q4). Висновки автора тут ≠ рішення власника. |
| Залежності / writers | Ресурс-конфлікт: compose-проєкт `name: courses` і порт 8080 фіксовані (`docker-compose.yml:1`, `.env.example:2`); стек зараз піднято з основного checkout (`multi-audio-s6`). `make check`/`e2e-reset` з цього worktree **знесе й обнулить** спільний стек → потрібне ексклюзивне heavy-вікно (Q3). |
| Scoped impact | Файли змін: `frontend/src/app/styles/*.css`, `frontend/src/presentation/**`, `frontend/src/modules/*/ui/**`, маршрути лише для розмітки. Callers/e2e-селектори звіряються per-slice (`frontend/e2e/*.spec.ts`, gotcha «друга file-контрола ламає strict mode»). Root `ef48c15`, dirty: лише цей план. |
| Планова критика / defense | Для S0 (read-only) — N/A: нема рішення, що закриває альтернативу. Для S1 та першого кодового зрізу — незалежна критика плану зрізу до коду. |
| Допуск | S0 — **BLOCKED до Q1/Q3**. Решта — BLOCKED до S0 + погодження. |

## 8. Питання власнику

- **Q1 — порядок і S0.** Погоджуєте базовий огляд S0 (без коду) перед поекранними фіксами і порядок §5?
- **Q2 — каденс PR.** Варіант A (рекомендую): коміт на екран, PR на групу поверхонь (client/auth, learn, admin, teach — ≈5 PR), повний гейт + незалежне рев'ю на PR. Варіант B: PR на кожен екран (≈27 PR, ≈27 повних гейтів).
- **Q3 — стек і heavy-вікно.** Браузерні знімки й `make check` потребують compose-стеку на 8080, яким зараз користується основний checkout. Дозволяєте брати ексклюзивне вікно (зупинити/перезібрати стек з цієї гілки) і коли — чи окремий compose-проєкт/порт (зміна інфри, окреме рішення)?
- **Q4 — рев'юер.** independent_subagent (Codex напряму) чи Courses 2?

## 9. Рішення

- **2026-11-16, власник** («Так, погоджую» на питання «огляд S0 без коду, далі коміт на екран і PR на групу поверхонь?»): **Q1** — S0 перед фіксами, порядок §5; **Q2** — варіант A (коміт на екран, PR на групу ≈5, повний гейт + незалежне рев'ю на PR). Перевірка: цей рядок + `work/tasks.md` DP0.
- **2026-11-16, автор (не власник):** S0 знімає **живий стек** на `localhost:8080` (Docker context `kiri`, `AGENTS.local.md`) без перебудови й без `make check` — тому Q3 для S0 не потрібен. Межа: образ `courses-frontend` створено 2026-11-15T00:51Z, точний коміт невідомий (за часом між `6ad8d8d` MS1 і `37ec4b0` F3) — S0-знахідки є кандидатами; кожен поекранний зріз перевіряє на власній ревізії. Записи в БД від S0 — лише сесії логіну трьох fixture-акаунтів (`frontend/e2e/auth.ts`), без мутацій контенту.
- **2026-11-16, автор:** D2 (`admin.courses.$courseId`) уже має UX-аудит `docs/ux-review/admin-course-editor-2026-11-15/` (основний checkout, не закомічений), фікс якого власник 2026-11-15 відклав до кінця MS2–MS4. S0 не дублює аудит: D2 знімається лише для доказів viewport, знахідки посилаються на той звіт; фікс D2 — у тій відкладеній задачі, не тут, поки власник не вирішить інакше.

## 10. Зріз AUTH — A1/A2/A4/A5 (власник 2026-11-16: «Спершу auth-зріз»)

**Scope / acceptance.** Закрити S0 A1-1 (blocker), A2-1, A5-1 і тупик A4 (сторінка «Забули пароль» без шляху назад; той самий сценарій відновлення). Спостережувана поведінка:
- **AUTH-1.** На `/login` (крок пароля) є посилання «Забули пароль?» → `/password/forgot` і «Зареєструватися» → `/register`. На кроці 2FA й у стані «вже увійшли» їх немає.
- **AUTH-2.** На `/register` завжди є «Вже маєте обліковий запис? Увійти» → `/login`. За `email_taken` до помилки додається «Скинути пароль» → `/password/forgot`.
- **AUTH-3.** На `/password/forgot` є «Згадали пароль? Увійти», а в стані «лист надіслано» — «Повернутися до входу».
- **AUTH-4.** На `/password/reset` без токена, а також після відповіді `invalid_account_token` / `token_used`, форми немає. Замість неї — заголовок стану, пояснення й основна дія «Запросити нове посилання» → `/password/forgot`. З валідним токеном усе як було; `weak_password` і помилки мережі лишають форму.

Non-goals: хедер гостя (S-2), спільні S-1/S-3/S-4, бекенд, тексти помилок.

**Small/full.** Small: лише presentation, без контрактів, стану чи API. Планова критика N/A за локальною нормою small; незалежне рев'ю коду перед мерджем лишається обов'язковим.

**Scoped impact.** Файли: `modules/identity/ui/{LoginView,RegisterView,PasswordForgotView,PasswordResetView}.tsx`, `app/providers/locales/{uk,en}.ts`, `main.css` (одне правило відступу `.auth-alt`), unit-тести `LoginView.test.tsx` і `AccountSelfServiceViews.test.tsx`. Перевірені селектори (grep `e2e/`, `src/`):
- `accounts.spec.ts:189,237,258,295,358` — кнопки «Увійти», «Зареєструватися», «Зберегти пароль», з валідним токеном; не зачеплені.
- `accounts.spec.ts:361` — axe на `/password/reset` без токена: тепер це стан AUTH-4, axe мусить лишитися чистим.
- `recovery.spec.ts:29` і `accessibility.spec.ts:236` — посилання «Увійти» на `/` і після logout, не на `/register`; не зачеплені.
- `AccountSelfServiceViews.test.tsx:38-39` — обидва посилання за `email_taken` лишаються.

Ризик strict mode: на `/register` буде два посилання «Увійти» (хедер і форма) з однією ціллю. Жоден наявний селектор не шукає його на цій сторінці.

**Докази.** Unit-тести на AUTH-1…4. Браузер: `/login`, `/register` (+ `email_taken` через UI), `/password/forgot` (форма і «надіслано»), `/password/reset` (без токена) — 4 viewport × light/dark, 200%, 320, Tab. Гейт: ordered `make check` перед PR групи client/auth (Q2); heavy-вікно — Q3.

**Допуск: READY** (AUTH-1…4 погоджені рішенням власника про зріз, залежностей немає).
