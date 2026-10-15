# admin-restructure — план (повний масштаб)

**Дата:** 2026-10-15. **Автор:** Courses 1 (за запитом власника).
**Статус:** чернетка плану; код не починався, крім editor-пілоту (див. §6).

## 0. Навіщо це фаза, а не точковий фікс

Власник (2026-10-15): editor — і **вся адмінка (8 views)** — не виглядає як на мокапі.
Проблема — **розміщення елементів і візуальна структура (layout), НЕ палітра** (кольори ок).

Це не регрес — це **недовиявлений gap**. Фаза `admin-restyle` (S0–S6, PR #23–#31) формально
«наклала» мокап на всі 8 views, АЛЕ її acceptance звірявся на **наявності CSS-класів + прогоні
axe**, а **ніколи — на layout-diff проти proto** (`docs/admin-restyle/proto/`). Токени й палітра
лягли; геометрія/композиція/порядок елементів — ні, і це пройшло «зелений» гейт.

> AGENTS.md: «зелений агрегатний гейт — не доказ рядка»; gotcha 2026-10-14: «admin-специфіку тягнуть
> у глобальний примітив і звуть scoped». Візуальна fidelity ніколи не мала перевірного кроку
> «порівняй із proto у тому самому viewport» — цей план його вводить.

## 1. Дві незалежні осі

| Ось | Що | Скіл | Доказ готовності |
|---|---|---|---|
| **V — Layout fidelity** | Геометрія/композиція/порядок елементів кожного з 8 views = proto | `design-pass` | side-by-side proto↔prod у цільовому viewport, перелік розбіжностей закрито |
| **A — Файлова архітектура** | Фронт отримує модульну структуру «як на беку» замість пласкої купи | `pattern-architecture-review` | кожен view має свій дім; гейт зелений після суто-механічного переносу |

Осі розведені НАВМИСНЕ: A — поведінково-нейтральний перенос файлів (diff = лише шляхи+імпорти),
V — зміни розмітки/CSS. Змішувати їх в одному коміті = нечитабельний diff. Порядок у §6.

---

## 2. Ось A — архітектурна знахідка (pattern-architecture-review)

### Факт: шари є, але принцип групування НЕсуперечливий

Верхні шари фронту вже віддзеркалюють гексагон беку і напрям залежностей коректний:
`domain/ ← application/ ← adapters/ports/presentation/`, composition-root у `app/`. Це **не**
«купа без структури» на рівні шарів — цю частину ламати НЕ треба.

Проблема — **усередині шарів**. `domain/` і `application/` нарізані **за модулем** (feature):
`domain/{course,discussion,document,enrollment,identity,media,playback,purchase,teach,transcript}`.
А `presentation/` нарізаний лише **за аудиторією** (`admin/ learn/ teach/ storefront/ identity/`)
і далі — **пласко**:

- `presentation/admin/` = **33 файли** в одній теці, що змішують ≥7 незв'язаних фіч:
  analytics (`AnalyticsView`), courses (`CourseListView`, `CourseEditorView`, `CoursePriceField`,
  `CourseGrantPanel`), lesson-editor (`LessonDocumentEditor`+`RichText`+`Loading`+`DraftField`+
  `VideoField`+`TranscriptField`+`DocumentLibrary`+`tiptap-document`), transcript (`TranscriptEditor`,
  `use-transcript-recovery`, `transcript-draft-offer`), users (`UserListView`), reports
  (`ParticipantReportQueue`), refunds (`RefundQueue`) + chrome (`AdminShell`, `AdminNavIcon`) + чужа
  `ResetPasswordDialog` (identity, лежить не там).
- `app/composition/` = **42 файли** пласко (`*-queries.ts` усіх модулів вперемішку).
- `app/router/` = **12 файлів** пласко (`*-search.ts` усіх модулів).

**Проблема (наслідок):** у фічі немає дому. Щоб торкнути «editor», треба знайти її шматки,
розкидані по `presentation/admin/*`, `app/composition/admin-document-queries.ts`,
`app/router/admin-search.ts`, `application/admin/*`, `domain/document/*`. На беку `internal/<module>/`
дає фічі один дім через усі шари; на фронті такого дому немає. Саме це й відчув власник.

### Лінзи (чесно, без накручування)

- **GoF-патерни:** N/A. Це не патерн-проблема (не Strategy/State/Factory). Притягувати GoF сюди
  немає підстав — тиск у коді суто організаційний.
- **SRP (на рівні теки) — головне:** `presentation/admin/` має ≥7 незалежних причин змінитися.
  Ціль: тека = одна фіча.
- **Clean Architecture:** напрям залежностей **коректний**, нові шари/порти/DI **не потрібні**.
  Єдиний дефект меж — `ResetPasswordDialog` в `admin/` замість `identity/`.
- **KISS:** НЕ інвертувати робочий гексагон у feature-first моноліт (`src/modules/<f>/…`) — це
  велика ризикована операція без доведеного виграшу понад те, що дають підтеки. Простіше й достатньо —
  додати feature-підтеки всередині наявних шарів.

### Цільова структура — A1 (feature-first модулі, РІШЕННЯ ВЛАСНИКА 2026-10-15)

Власник обрав **A1**: фронт дзеркалить гексагон беку — кожна фіча отримує вертикальний модуль зі
своїм внутрішнім шаруванням (як `internal/<module>/{domain,application,adapters}` на беку).

```
src/
  modules/
    <module>/          # course, document, transcript, identity, enrollment, discussion,
                       #   playback, catalog, media, purchase, teach, admin-*
      domain/          # ← з src/domain/<module>
      application/     # ← з src/application/<module> (+ локальні ports цього модуля)
      adapters/        # ← api/storage-адаптери модуля
      ui/              # ← з presentation/*/<feature> (admin + client views цього домену)
      index.ts         # PUBLIC SURFACE — єдиний вхід імпорту для інших модулів
  contracts/           # НЕЙТРАЛЬНА крос-модульна зона (аналог internal/contracts, ADR 013):
                       #   domain-safe типи/чисті предикати, спільні багатьом модулям
  ui/                  # спільні presentation-примітиви (controls/data/overlays) — НЕ модуль
  app/                 # composition root, providers, router, styles — зшиває модулі,
                       #   МОЖЕ імпортувати index.ts модулів
  routes/              # TanStack file-based (лишаються пласкі — конвенція фреймворку)
```

Editor мапиться на модуль **`document`** (+ дотичні `transcript`, `media`). Він — перший
міграційний зріз, що ДОВОДИТЬ точні межі (що переїжджає в модуль, що лишається в `app/`).

### Контракти vs index.ts — відповідь на питання власника (доказова, 2026-10-15)

**Замір поточного коду:** `domain/<A>` ніколи не імпортує `domain/<B>` (**0** випадків) — домени вже
незалежні. Єдиний справді крос-модульний тип — `AuthenticatedUser` (identity), **18** використань
через усі шари.

- **`index.ts` на модуль — НЕОБХІДНИЙ** (barrel, не директорія): єдиний дозволений вхід імпорту для
  інших модулів + lint-правило «no deep cross-module import». Це головний механізм інкапсуляції;
  його достатньо, щоб модуль «віддавав» функції/типи назовні контрольовано.
- **Спільний `contracts/` (одна нейтральна зона) — ВИПРАВДАНИЙ, але МІНІМАЛЬНИЙ:** лише для дійсно
  крос-модульних domain-safe типів/предикатів (сьогодні — фактично `AuthenticatedUser`). Починати з
  одного файла, НЕ з розлогого дерева (skill: «не додавай невикористаний seam; документуй майбутній
  шов перед створенням»). Альтернатива — лишити такий тип публічним у модулі `identity`, який інші
  імпортують як фундаментальний; `contracts/` робить це явним і прибирає напрям «усі→identity».
- **Per-module `port`/`contract` директорія — НЕ виправдана:** домени вже 0 крос-імпортів, окрема
  contracts-тека на КОЖЕН модуль = порожній seam. Увага на термінологію: фронтовий **`port`** =
  інтерфейс до ЗОВНІШНЬОЇ залежності (I/O-межа), і він уже існує в `src/ports/` — це ІНШЕ поняття,
  ніж крос-модульний contract, не плутати. Ports лишаються (у модулі або спільні), contracts — окремо.

**Стисло:** `index.ts` достатньо для public-surface кожного модуля; окрема port/contract-тека на
модуль — надлишок; **одна** спільна `contracts/`-зона потрібна лише для жмені крос-cutting типів
(старт — `contracts/identity.ts` з `AuthenticatedUser`).

### Крок A — виконання (поведінково-нейтрально)

1. Перенос `git mv` у підтеки, правка імпортів, `index.ts` на тек у разі потреби.
2. Нуль змін логіки/розмітки. Diff = лише шляхи та імпорти.
3. Гейт `make check` зелений ДО і ПІСЛЯ — доказ нейтральності.
4. Один PR на механічний перенос (великий, але тривіальний для рев'ю).

---

## 3. Ось V — метод design-pass (однаковий для кожного view)

1. **Render proto:** `docs/admin-restyle/proto/` → `npm ci && npm run build && http.server`
   (або esbuild-бандл на вільний порт). Відкрити `#<view>`.
2. **Render prod:** стек на `localhost:8080` (nginx; **НЕ** :4000/:8899 — то Cassiopeia), залогінений
   admin, цільовий маршрут.
3. **Side-by-side той самий viewport** (desktop 1440×900 + mobile 390×844): спершу великі пропорції
   й ієрархія, потім типографіка/відступи/деталі. Зафіксувати **перелік layout-розбіжностей**.
4. **Виправити** розмітку/CSS у межах view. **Не чіпати** палітру/токени (вони ок).
5. **Переперевірити** свіжим скріном (скрін до-фіксу не доказ фіксу) + axe@200% + клавіатура + стани.
6. Acceptance view закрито, коли перелік розбіжностей порожній на обох viewport.

Кожен view = один PR (branch+PR, owner merge). Гейт локальний (`make check`; CI Actions червоний
через білінг, не код).

---

## 4. 8 views — черга й acceptance (Ось V)

| # | proto | prod route | prod файли | нотатки |
|---|---|---|---|---|
| 1 | `#lesson` | `/admin/courses/:c/lessons/:l` | lesson/* | **ПІЛОТ** — див. §6 |
| 2 | `#courses` | `/admin` | CourseListView | таблиця→row-cards ≤52rem |
| 3 | `#course` | `/admin/courses/:c` | CourseEditorView | форма, модулі |
| 4 | `#analytics` | `/admin/analytics` | AnalyticsView | tiles, numeric align |
| 5 | `#learners` | `/admin/learners` | (teach/admin learners) | progress track |
| 6 | `#users` | `/admin/users` | UserListView | роль/сорт фільтри |
| 7 | `#reports` | `/admin/participant-reports` | ParticipantReportQueue | status pills |
| 8 | `#refunds` | `/admin/refunds` | RefundQueue | confirm-flow |

Порядок після пілоту узгодити; за замовчуванням — за частотою використання/ризиком.

---

## 5. PR #33 (editor-polish, поведінка) — ЗМЕРДЖЕНО 2026-10-15

Рішення власника 2026-10-15: **змерджити зараз**. Виконано: squash-merge у `main` 19:53 UTC
(CI UNSTABLE = білінг, не код). Поведінка (per-image upload, no-text-loss, focus-management,
чесний save-status, invalid-URL-keeps-link, announce) тепер у `main`.

Його UI-рішення **конфліктують із proto** і будуть перероблені в editor-пілоті (§6):
- лінк: inline-поле → **ModalBox** (E11);
- тулбар: текстові кнопки → **icon-toolbar з Tooltip + caption** (E10).

Поведінку ЗБЕРЕГТИ на новій розмітці; поведінкові тести PR #33 лишаються регресом.

---

## 6. Порядок робіт — editor як вертикальний пілот

Власник: «саму роботу починай з editor». Editor — **пілот обох осей** і еталон методу:

1. **V-діагностика (цей крок, зараз):** render proto `#lesson` + prod editor, side-by-side,
   **точний перелік layout-розбіжностей** (не «виглядає не так», а список: PageHeader eyebrow,
   segmented Editor|Preview, save-status dot, section-card index «01» + drag-handle, icon-toolbar,
   Empty-states, link-ModalBox). → надіслати власнику.
2. **V-пілот:** закрити перелік розбіжностей §6.1 у розмітці/CSS + інтегрувати UI-переробку (E10/E11)
   зі збереженням поведінки PR #33 (вже в `main`). Гейт зелений.
3. **A-пілот (A1):** винести editor у модуль **`modules/document/`** (`ui/` + `domain/` +
   відповідні queries), `index.ts` як public surface, завести `contracts/identity.ts`. Поведінково-
   нейтрально, гейт зелений ДО і ПІСЛЯ. Цей зріз ДОВОДИТЬ точні межі A1 — власник дивиться й
   коригує гранулярність ПЕРШ НІЖ масова міграція решти модулів.
4. **Гейт + design-pass переперевірка** (desktop+mobile, axe@200%, клавіатура, стани), PR, owner merge.
5. **Пілот = еталон:** метод (V) і структуру модуля (A1) тиражувати на views 2–8 / решту модулів.

**Порядок V перед A у пілоті:** V не має відкритих рішень (точний E-список є) і дає видимий виграш
одразу; A1 великий — хай переробка UI осяде, тоді чистий mv у модуль. Глобальна A1-міграція
(domain/application/adapters усіх модулів) — окремий трек після схвалення пілот-меж.

## 7. Ризики

- Перенос файлів ламає нетреновані імпорти → гейт `make check` + typecheck ловлять; PR окремо.
- proto — не білд (dist немає) → рендер через esbuild/http.server, НЕ плутати порти (8080 prod,
  4000/8899 = Cassiopeia).
- Переробка editor-UI не має відкотити поведінку PR #33 → поведінкові тести лишаються регресом.
