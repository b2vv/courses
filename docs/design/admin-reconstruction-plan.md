# Admin 1:1 reconstruction — per-page plan

| | |
|---|---|
| Дата | 2026-10-09 |
| Автор | Courses (design-pass) |
| Рішення власника | «1 в 1 як на відео» = **макет-граматика NorthWind на НАШИХ реальних даних** (не дослівні infra-екрани, не лише скін). Обсяг сесії: **спершу цей план**, код — після узгодження. |
| Джерела | Дизайн-система: `docs/design/northwind-admin-spec.md`. Рекон даних: `work/archive/_wip-admin11-recon.md`. Розкадровка відео: scratchpad `admin-ref/p_*.jpg`. |
| Метод | `reference-ui-reconstruction` (факти зафіксовано) → план → слайси з Playwright-скрінами (3 ширини) + ordered `make check`. |

## 0.5. Рішення власника (2026-10-09) — закривають альтернативи

- **R1. Агрегати: A зараз + B пізніше.** Реальні числа — фронт-count-запити (`limit=1`→`total`) там, де є пагінація з total; чесний 🟡 там, де лише cursor. Справжні бек-`stats`-ендпоінти (B, §4-1) — **окремий запланований етап доробки статистики**, не блокує реконструкцію. Власник: «адмінку будемо допрацьовувати, бо багато чого не вистачає по статистиці». *Закриває:* «вигадати метрику» і «чекати бек перед реконструкцією».
- **R2. IA: `/admin` = Overview-лендинг, список курсів → новий `/admin/courses`** (+ редірект зі старих посилань). *Закриває:* «band зверху списку на тому ж маршруті».
- **R3. Icon-rail — у S0** разом із секційним сайдбаром. *Закриває:* «rail потім».

## 0. Залізне обмеження (визначає все нижче)

**На беку НЕМАЄ жодного stats/aggregate/dashboard-ендпоінта** (`work/archive/_wip-admin11-recon.md` §підсумок). Усі 5 `/admin`-сторінок — списки або редактор. Отже кожне NorthWind-число треба класифікувати:

- 🟢 **Реальне зараз** — є в наявній відповіді: `pagination.total` (курси, юзери — по 1 на фільтр), `progressPercent` 0–100 (learner×course), video `receivedBytes/declaredBytes`+`state` (список відео повний), module/lesson counts із course tree, status-пігулки (courses, users).
- 🟡 **Клієнт-агрегація** — рахується на фронті з наявного: video-states донат (повний список), розподіл по сторінці (⚠️ лише завантажене, не весь набір).
- 🔴 **Потрібна нова робота на беку** — точні агрегати там, де cursor-пагінація без `total` (learners, participant-reports) або розподіл за статусом/роллю одним запитом. **Не вигадувати число — або N запитів, або новий ендпоінт, або чесний «—».**

Правило плану: жоден блок не показує вигадане число. Де 🔴 — або тонкий count-ендпоінт (окремий бек-слайс), або блок деградує до того, що є (🟢/🟡), з чесним підписом.

## 1. Глобальний shell (усі сторінки) — слайс S0

Патерн: спека §4.11, §5. Зараз (`routes/__root.tsx:145`, `AdminShell.tsx`): плоский сайдбар (brand → 4 лінки → "to client"), простий topbar. Без лічильників.

- **Сайдбар:** секційний (UPPERCASE-лейбли груп), активний пункт = готовий градієнт-pill (✅ зроблено цю сесію на `.admin-nav-link`). Rail (вузька іконкова колонка) — **опційний**, лишаємо на потім (спека §4.11: «для нашої адмінки rail опційний»).
  - Групи під наш невеликий набір: **КУРСИ** (`/admin`), **ЛЮДИ** (`/admin/learners`, `/admin/users`), **МОДЕРАЦІЯ** (`/admin/participant-reports`). Лічильники праворуч лінків — 🔴 (нема total дешево) → **без лічильників** поки (чесно), або 🟢 total курсів/юзерів тільки на тих двох.
- **Topbar:** title (є) + `• Live`-індикатор — ❌ викидаємо (у нас нема live-стріму, це був би fake); лишаємо user.displayName + LanguageSwitcher + SignOut (є), приводимо до NorthWind-стилю (secondary-кнопка §4.1).
- **Токени:** вже застосовані минулою сесією (D1–D3). S0 добиває shell-рівень.

## 2. Per-page

### 2.1 `/admin` — Overview (аналог головного екрана відео) — слайс S1 (найбільший)

Зараз: **список курсів** (CourseListView: heading+count → tabs(status) → table → pagination). Дані: `admin.list` → `{items[id,slug,title,status,instructorId,updatedAt], pagination.total}` + `admin.instructors`.

NorthWind-граматика: hero-метрика + таби + бар-чарт + health-таблиця + lifecycle-панель. Мапінг:

- **Hero-band (ui-stat × 2–3):** 🟢 **Курсів усього** (`admin.list` total без фільтра) + 🟢 **Юзерів усього** (`admin.users.list` total). Delta `vs` — 🔴 (нема історії) → **без delta** або прибрати порівняльний підпис. Третя плитка «Відкриті скарги» — 🔴 (cursor, нема total) → **не ставити**, поки нема count-ендпоінта.
- **Бар-чарт (time-series у відео):** ⚠️ у нас **нема часових рядів**. Чесні варіанти: **(а)** замінити на **розподіл-бар «Курси за статусом»** = 4 легкі запити (draft/published/archived/all) — 🟢×4; **(б)** гістограма `updatedAt` по завантаженій сторінці — 🟡 оманлива (лише сторінка) → відхилити. **Рекомендація: (а)** — це НЕ time-series, але чесний бар на реальних числах.
- **Health-таблиця (= Service health):** 🟢 **таблиця курсів** у NorthWind-стилі (§4.7): status-крапка+назва (mono), owner (instructor displayName), updated, status-пігулка. Це по суті поточний список, переодягнений.
- **Права колонка (= lifecycle/summary):** 🟢 донат «курси за статусом» (ті самі 4 числа) + легенда. Lifecycle-прогрес-бари — 🔴 (нема deploy-процесу) → замість нього «останні оновлені курси» (updatedAt sort, 🟢).
- **Табки Requests/Latency/…** → ❌ не мапляться (нема метрик); заміна — status-фільтр як сегмент-контрол (§4.6), що вже є логікою табів.

**Parity-gap (чесно):** немає часових рядів і delta → головний чарт відео (time-series) замінено розподілом; «Live» прибрано. Це свідомий відхід — записати в план-лімітаціях.
**IA (R2):** `/admin` стає чистим Overview-лендингом; список курсів (CourseListView) переїжджає на **новий `/admin/courses`** + редірект зі старих внутрішніх посилань. S1 включає створення маршруту `/admin/courses`.

### 2.2 `/admin/users` — Fleet-таблиця — слайс S2

Зараз: heading+count → search → tabs(all/active/blocked/deleted) → table[email·name·role·status·actions] → pagination. Дані: `admin.users.list` → `{items[email,displayName,role,status,updatedAt], pagination.total}`.

- **Фільтр-бар §4.8:** search (є) + Role-дропдроп + Status-сегмент (є як tabs) + лічильник «N results» 🟢 (`pagination.total` на фільтр) + Sort.
- **Таблиця §4.7:** status-крапка+email (mono) + displayName-підзаголовок; колонки Role (env-тег-стиль), Status-пігулка (🟢 є), updated (mono), дії (block/restore, delete, reset-pw — є).
- **Групові заголовки §4.7** (як регіони Fleet): групування за **роллю** (student/instructor/admin) — ⚠️ клієнт групує лише завантажену сторінку → або прибрати групування, або N-запит на роль. Рекомендація: **без груп** поки (пагінація посторінкова), лишити плаский NorthWind-table.
- **Summary-панель праворуч (= Fleet summary):** розподіл за status/role — 🔴 (N запитів: 3 status + 3 role). Рекомендація: **або** 6 легких count-запитів (`limit=1`, беремо `total`) → реальний донат; **або** панель прибрати. Схиляюсь до 6 запитів — числа реальні, запити дешеві.
- **Inline-метри (CPU/mem Fleet):** ❌ у юзерів нема числових метрів → не ставити.

### 2.3 `/admin/learners` — таблиця з прогрес-метрами — слайс S3

Зараз: h1 → filters(q, state, course) → ul learners (article: name, lastActivity, курси з ProgressBar) → infinite. Дані: `teach.learners('admin')` → `{entries[userId,displayName,lastActivityAt,courses[courseId,title,state,progressPercent,enrolledAt]], nextCursor}` (⚠️ **нема total**).

- **Найкращий фіт для метрів §4.5:** `progressPercent` 🟢 готовий 0–100 → inline-метр per course з mono-% і порогом.
- **Таблиця §4.7:** рядок learner = статус-крапка (active vs never за lastActivityAt 🟢) + name (mono) + lastActivity; під ним курси зі state-пігулкою (🟢) + progress-метр (🟢).
- **Summary-панель:** донат «середній прогрес» / «active vs never» — 🔴 (cursor, нема total; точно лише прокрутивши все). Рекомендація: **або** панель-«по завантажених N» з чесним підписом «серед завантажених», **або** count-ендпоінт. Схиляюсь до чесного 🟡 підпису (не брехати про весь набір).
- **Фільтр-бар §4.8:** q + state-Select (є) + course-filter (є) + «N завантажено» (не «N усього»).

### 2.4 `/admin/participant-reports` — Incidents-черга + drawer — слайс S4

Зараз: heading → список article(kind, reporter, target, course, reason, state, createdAt, дії Resolve/Revoke) → infinite. Дані: `enrollment.participantReports` → `{items[id,kind(on_lecturer|on_student),reason,state(open|resolved),createdAt,reporterName,targetName,courseTitle], nextCursor}` (⚠️ нема total).

- **Черга §5:** рядки-інциденти з kind-тегом + state-пігулкою (🟢: open=warning/degraded, resolved=healthy), createdAt-timeline (🟢 відносний час).
- **Drawer-деталь (= Incident panel відео):** клік по рядку → бічна панель: reporter→target, course, reason (повний), createdAt, дії Resolve / Revoke(діалог, лише on_student). Це головний «1:1»-елемент (панель інциденту з відео).
- **Лічильники open/resolved:** 🔴 (cursor) → заголовок без числа або 🟡 «N завантажено».

### 2.5 `/admin/courses/$courseId` — Deployments-detail (вільний аналог) — слайс S5

Зараз: **редактор** (course tree форма, price, status-transition, modules+lessons, video upload/resume, транскрипти, ASR drafts, CourseGrantPanel). Дані: `admin.detail` → CourseTree(modules[lessons], price, status, timestamps); `admin.videos.list` → video[lessonId,state,generation,receivedBytes,declaredBytes,failureReason] (**повний список**).

- Це редактор, не метрики — **лишаємо функціонал**, одягаємо в NorthWind-панелі (§5 картки-секції).
- **Header-summary (реальні агрегати 🟢/🟡):** модулів/уроків count (з tree 🟢); **донат станів відео** (processing/ready/failed) — 🟡 але з **повного** списку → фактично 🟢 реальний; статус курсу-пігулка.
- **Inline-метр аплоаду/транскоду §4.5:** `receivedBytes/declaredBytes` 🟢 готовий — per-lesson video-прогрес.
- Пайплайн-кроки / canary / Promote — ❌ нема такого процесу → не мапимо (це найдальший від нас екран відео).

## 3. Послідовність слайсів

Кожен слайс: реалізація → Playwright-скрін 3 ширини (моб/планшет/десктоп) → ordered `make check` (format→lint→typecheck→test→integration→build→e2e) → запис у звіт. Порядок від спільного до окремого:

1. **S0 shell** (сайдбар-секції, topbar, прибрати Live) — основа для решти.
2. **S2 `/admin/users`** — Fleet-таблиця найчистіший фіт, мінімум нових даних.
3. **S3 `/admin/learners`** — метри вже є, додає компонент прогрес-таблиці.
4. **S4 `/admin/participant-reports`** — черга + drawer (новий паттерн drawer).
5. **S1 `/admin` Overview** — найбільший, залежить від IA-рішення і count-запитів.
6. **S5 `/admin/courses/$courseId`** — редактор-у-панелях, останній (найбільший файл, найменше «метрик»).

(Overview не перший попри те, що він перший у відео — він найбільше впирається у відсутні агрегати й IA-рішення; дешевші чисті слайси йдуть раніше.)

## 4. Рішення власника — ВИРІШЕНО (див. §0.5)

1. **Агрегати:** A (фронт-count) зараз; B (бек-`stats`) — окремий пізній етап доробки. → R1.
2. **IA `/admin`:** Overview-лендинг + список курсів на новому `/admin/courses` (+редірект). → R2.
3. **Rail:** у S0. → R3.

## 5. Parity-лімітації (магнітуда/механізм, записати у звіт при реалізації)

- **Немає часових рядів** → головний time-series-чарт відео замінено розподілом-баром на реальних числах. *Механізм закриття:* додати бек-ендпоінт із бакетами по `created_at/updated_at`.
- **Delta/«vs yesterday»** → прибрано (нема історії). *Механізм:* той самий time-series-ендпоінт.
- **`• Live`** → прибрано (нема стріму). Повернення = SSE/polling, окрема робота.
- **Точні лічильники learners/reports** → недосяжні без прокрутки (cursor без total). *Механізм:* count-ендпоінт.
- Мікролейбл-контраст, шрифти — успадковані parity-винятки зі спеки §8.
