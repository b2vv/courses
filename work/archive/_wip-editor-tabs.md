# admin-restructure — зріз «редактор: radio-segmented → React Aria Tabs» (V-track #5)

**Гілка:** `feat/editor-tabs-semantics` (стек на `feat/view-course-designpass` / PR #38, ще не змерджено).
**Автор:** Courses 1. **Старт:** 2026-10-17.

## Проблема

`CourseEditorView` мав таби Огляд/Уроки/Доступ, реалізовані через `SegmentedControl`
(= RA `ToggleButtonGroup selectionMode="single"` → рендерить `role="radio"`,
підтверджено `ui.test.tsx:354-363`). Семантично це radiogroup, а не вкладки: SR читає
«радіокнопка», не «вкладка N з 3»; немає зв'язку tab↔tabpanel. Proto `#course` показував
саме табовий каркас (`.ui-segmented.course-tabs`). Зріз піднімає семантику до справжніх
`role=tab`/`role=tabpanel`, зберігши вигляд пігулок.

## РІШЕННЯ (2026-10-17, Courses 1, за вибором власника «A»)

**Варіант A — force-mount панелей, один `<form>`, одна кнопка Save.** Закриває альтернативу
«розбити форму, Save на вкладку» (Варіант C).

- **Чому A, не C:** збереження курсу — одна атомарна операція на все дерево
  (`onSave(current, edit): CourseTree`, `CourseEditorView.tsx:75`; `CourseFormValues` містить
  і деталі, і `modules[].lessons[]`). C розколов би це на два submit → або дублювання правила
  (кожен шле все дерево, lost-update між вкладками), або неіснуючі partial-save ендпоінти
  (бек-робота поза скоупом design-pass). Оцінка `pattern-architecture-review`: **A 95% / C 31%**
  (KISS/DRY/SRP усі за A; C протікає UI-рішенням у save-контракт — рух залежності в хибний бік).
- **Спосіб перевірки рішення:** `onSave` лишається одним портом; один Save зберігає обидві
  вкладки; e2e save-флоу не переписувались (лише роль контрола). Якщо з'явиться потреба в
  per-tab save — це нове рішення з новим портом, не мовчазний дрейф.

## Механізм A

`Aria.Tabs` (controlled `selectedKey={tab}`) з `TabList` (клас `.ui-segmented`, кожен `Tab` —
`.ui-segment`: вигляд пігулок як був, RA на `Tab` теж емітить `data-selected`/`data-hovered`).
`overview`/`lessons` — `TabPanel shouldForceMount` **усередині** єдиної `<form>` → форма завжди
змонтована, один submit, а `focusField`(`flushSync(setTab)`→`setFocus`) крос-табовий фокус
працює як раніше. `access` (`adminTools`) — окремий `TabPanel`. `CoursePriceField` має власний
`<form>` → лишається сиблінгом поза головною формою з `hidden={tab!=='overview'}` (як було).

## Змінені файли (11)

- `frontend/src/presentation/admin/CourseEditorView.tsx` — `SegmentedControl`→`Aria.Tabs`;
  імпорт `* as Aria`.
- `frontend/src/app/styles/main.css` — `.course-tabs` flex→grid gap space-8 (ритм TabList↔панелі).
- unit: `CourseEditorView.test.tsx` (8), `LessonVideoField.test.tsx` (1): `radio`→`tab`.
- e2e (18 клік-сайтів, 7 файлів): admin, asr, enrollment, media, playback, purchase, transcript —
  `getByRole('radio', {name:/Lessons|Уроки|…/})`→`getByRole('tab', …)`.
- **НЕ чіпав:** `SegmentedControl` (живе ще в AnalyticsView/UserListView/CourseListView/
  LessonDocumentEditor); radio-фільтр «Видалений» (admin-users); рейтинг «N з 5» (ratings).

## Верифікація

- format-check ✓, lint ✓ (architecture 210 files), typecheck ✓.
- frontend unit: **772 passed** (вкл. editor axe B09, error-summary focus S2b, price/preview).
- Повний ордерний `make check`: **exit=0** (2026-10-17). e2e **121 passed (15.3m)** у реальному
  браузері проти Compose — усі 18 мігрованих tab-кліків зелені. Лог:
  `frontend/test-logs/api-nginx-20261017T102216Z.log`.

## Обмеження

- `CoursePriceField` показується на вкладці Огляд, але лежить поза `role=tabpanel` цієї вкладки
  (окрема форма). Поведінка ідентична попередній; ідеальна a11y-вкладеність потребувала б
  винести price-форму в панель — відкладено як окрему можливість, не регрес.

## Наступне

- Дочекатись `make check`; записати e2e-рядок. Відкрити PR (стек на #38).
