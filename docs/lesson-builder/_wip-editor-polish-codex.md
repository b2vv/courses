# WIP — Codex реалізація: editor-polish Run A (QW1 + QW2 + QW4)

Робочий файл. **Дописуй ПІСЛЯ КОЖНОГО КРОКУ (append).** Перший запис — `ДАЛІ:`. Далі ФАКТ/РІШЕННЯ/ДАЛІ.

## Контекст
Гілка `feat/editor-polish`. Власник (2026-10-13): редактором уроку **неможливо працювати** — привести в
порядок. Повний аналіз і план — `docs/lesson-builder/_wip-editor-design-codex.md` (твій же, читай його).
Це **Run A** — ядро «щоб працювало»: **QW1 + QW2 + QW4**. QW3/QW5/QW6 і D-слайси — окремі наступні прогони,
НЕ чіпай їх зараз.

## Задача Run A
- **QW1** — надати секції редактора видиму структуру: header (тип ліворуч, ↑/↓/видалення праворуч) →
  toolbar → окреслена зона вводу з видимою висотою, padding, focus-ring. Scoped `.lesson-editor-*` у CSS +
  токени (`--space-*`, `--color-background/surface/hairline/border`, `--radius-panel`, `--control-min-size`).
- **QW2** — замінити сирі upload/alt поля на наявні примітиви `FileUpload`/`TextField`
  (`frontend/src/presentation/ui/controls.tsx`), вирівняти image-layout, обмежити preview шириною секції.
- **QW4** — увімкнути grid колонок Row у редакторі (аналог reader-grid, `minmax(0,1fr)`, стек за вузької
  ширини); дати видимі підписи палітр «Додати секцію»/«Додати колонку» (native `legend`), внутрішні межі Row.

## ⛔ ОБОВ'ЯЗКОВО ЗБЕРЕГТИ (інакше e2e `frontend/e2e/lesson-builder.spec.ts` червоний)
Не перейменовувати й не прибирати (ролі / accessible names / класи, які тести селектять):
- heading «Редактор уроку»; `role="group"` з назвою **«Додати секцію»**; кнопки палітри за назвами
  **«Текст» / «Зображення» / «Відео» / «Завдання»**; кнопки **«Опублікувати» / «Зберегти чернетку» /
  «Пересунути вгору»** (і «Пересунути вниз»/«Видалити» аналогічно); `role="textbox"` з назвами **«Текст»**,
  **«Завдання»**.
- file input лишається досяжним через `getByLabel('Завантажити зображення')`; alt-поле —
  `getByLabel('Опис зображення')`; клас прев'ю **`.lesson-editor-image-preview`** лишається на елементі
  зображення; клас секції **`.lesson-editor-section`** лишається.
- reader: `role="region"` з назвою **«Матеріали уроку»**; зображення лишається `role="img"` з alt.
Тобто `FileUpload`/`TextField` переюзуй так, щоб ЦІ accessible names і класи збереглися (за потреби —
проп label/className). Якщо примітив не дає зберегти — лиши наявний елемент, додай лише оформлення.

## Константи
- Адмінка **dark-only** (`tokens.admin.css`, owner 2026-10-09) — стилізуй під dark-скоуп; reader — light/dark.
- Строгий TS, React Aria, **семантичні токени + scoped CSS** (Tailwind у репо є, але house-style — токени, не
  utility-класи). Без нових залежностей. Контроли ≥44px, видимий фокус, 200% зум, без обрізання.
- НЕ чіпати: модель документа (`domain/document/document.ts`), конвертер (`tiptap-document.ts`), контракти,
  правила save/publish/доступу. Лише presentation + CSS. Без D-слайсів (статуси/undo/preview) цього прогону.

## Файли
- `frontend/src/presentation/admin/LessonDocumentEditor.tsx` (редактор)
- `frontend/src/presentation/admin/LessonDocumentRichText.tsx` (toolbar/контейнер — лише якщо QW1 вимагає зони вводу)
- `frontend/src/presentation/learn/LessonDocumentView.tsx` (reader — лише grid Row для узгодження QW4, без QW3)
- CSS: куди покласти scoped правила — звір із наявним (`frontend/src/app/styles/main.css` секція lesson-document
  ~1352–1379; плюс `tokens.css`/`tokens.admin.css` для значень). Де вже є клас без правил — додати правило.
- Наявні примітиви: `frontend/src/presentation/ui/{controls,states,index}.tsx`.

## Після реалізації (НЕ ти — Claude зробить)
Claude зганяє FE-гейт (format-check/lint/typecheck/test/build) + повний e2e (бо shared CSS) + перегляне в
браузері. Ти код не комітиш.

## Очікуване від тебе
Внеси зміни QW1+QW2+QW4 у перелічені файли. Після — допиши сюди стислий звіт: які файли/класи додано, які
e2e-селектори свідомо збережено, що лишив на наступні прогони.

---
## Журнал Codex (append)

ДАЛІ: 2026-10-13, Codex — звірити QW1/QW2/QW4 з поточними компонентами, примітивами й токенами; реалізувати лише presentation та scoped CSS зі збереженням обов’язкових селекторів.

ФАКТ: звірено компоненти та QW1/QW2/QW4; наявні editor-класи не мали оформлення. Додано групу дій секції, підписи колонок, native legend обох палітр, TextField для alt і клас контейнера rich-text.
РІШЕННЯ: 2026-10-13, Codex — upload лишається нативним label/input (дозволений fallback): FileUpload обгортає Aria.FileTrigger без явного зв’язку label із прихованим input; зберігаємо чинний getByLabel та callback/reset. Перевірка — input у label «Завантажити зображення», незмінний accept і e2e setInputFiles.
ДАЛІ: додати scoped CSS для структури секцій, поля, upload та адаптивних колонок.

ФАКТ: додано scoped CSS на чинних токенах: секція/header/actions, toolbar із переносом, поле rich-text (min-height 8rem, border, padding, focus-ring), image/upload, палітри й внутрішні межі Row. Адмінка успадковує чинні dark-only токени, нових тем/залежностей немає.
РІШЕННЯ: 2026-10-13, Codex — редактор і reader використовують однакові container queries: 2 колонки від 40rem доступної ширини, 3 від 60rem, нижче — стек. Перевірка: 1/2/3 колонки, вузька область поруч із sidebar і текстовий зум 200%; ці пороги — початковий layout, не заява про пройдений browser-check. Reader JSX не потребував змін.
ДАЛІ: перевірити формат і diff, збереження селекторів та межі Run A; повний FE-гейт/e2e/browser лишаються Claude відповідно до завдання.

ФАКТ: Run A реалізовано у трьох кодових файлах: `LessonDocumentEditor.tsx`, `LessonDocumentRichText.tsx`, `main.css`. Додані класи: `.lesson-editor-section-actions`, `.lesson-editor-column-bar`, `.lesson-editor-palette`, `.lesson-editor-upload`, `.lesson-editor-input`; решта правил оформлюють наявні `.lesson-editor-*`. Reader змінився лише CSS grid/колонок, його JSX не змінено.
ФАКТ: diff звірено з `frontend/e2e/lesson-builder.spec.ts`: свідомо збережено heading «Редактор уроку», group «Додати секцію», назви кнопок «Текст»/«Зображення»/«Відео»/«Завдання», «Опублікувати»/«Зберегти чернетку», «Пересунути вгору»/«Пересунути вниз»/«Видалити»; textbox «Текст»/«Завдання», label «Завантажити зображення»/«Опис зображення», `.lesson-editor-section` і `.lesson-editor-image-preview` на img. Reader region «Матеріали уроку» й img alt лишилися незмінними. Це статична перевірка, не результат e2e.
ФАКТ: локальні `oxfmt --check` для трьох змінених кодових файлів та `git diff --check` — exit 0. Комітів немає; модель, конвертер, контракти, save/publish/access, залежності та сторонні наявні зміни не зачеплено.
ДАЛІ: Claude — FE-гейт (format-check/lint/typecheck/test/build), повний e2e та браузерна перевірка dark-admin, reader light/dark, порожнього/довгого rich-text, upload/alt, 1/2/3 колонок і 200% зуму. Ці перевірки тут НЕ запускались і НЕ оголошуються пройденими. QW3 (prose/task), QW5 (активні стани toolbar), QW6 (loading/error/empty) та D-слайси залишено наступним прогонам.

---
## Run B — QW3 + QW5 (окремий прогін, поверх Run A)

Run A (QW1/QW2/QW4) уже в дереві (гілка feat/editor-polish, не закомічено). Тепер **Run B**:
- **QW3** — прозова типографіка reader І текстової зони editor: ритм p/list (відступи), виразні h2/h3,
  видимі маркери списків, підкреслені посилання; **task** оформити стриманою окремою областю з локалізованим
  підписом «Завдання» (reader). Орієнтир: текст 1–1.125rem, line-height ~1.6–1.75, прозова колонка ~65ch
  (лише для текстової колонки, НЕ для player/row). Spared CSS-правила типографіки, не shared Tiptap runtime.
- **QW5** — toolbar rich-text: згрупувати відступами + перенос; стилізувати активний стан
  `[aria-pressed='true']` (видимий не лише кольором); лишити текстові accessible names і цілі ≥44px.

ОБОВ'ЯЗКОВО ЗБЕРЕГТИ — той самий список селекторів, що в Run A (heading/group/кнопки/textbox/labels/класи/
reader region). Task у reader лишається з текстом «Виконайте вправу» досяжним (e2e `region.getByText('Виконайте вправу')`).
Константи ті самі (dark-admin, токени+scoped CSS, React Aria, без залежностей, без D-слайсів і без QW6).
Файли: `LessonDocumentView.tsx` (reader QW3), `LessonDocumentRichText.tsx` (QW5 + editor prose), `main.css`.
НЕ коміть. Після — допиши звіт Run B нижче.

## Журнал Codex — Run B (append)

ДАЛІ: 2026-10-13, Codex — реалізувати QW3/QW5 поверх Run A у трьох визначених файлах, зберегти доступні назви й текст завдання; звірити формат та diff.

ФАКТ: прочитано Run A/B і дизайн QW3/QW5; палітра й шрифти визначені tokens.css/tokens.admin.css, кнопки — наявний React Aria Button.
РІШЕННЯ: 2026-10-13, Codex — системний інтерфейс для автора й учня: спільний клас прози, 1.0625rem/1.65 і максимум 65ch лише для тексту; task — підпис без додаткового heading, рамка й padding; toolbar — чотири візуальні групи, активність через підкреслення й border. Перевірка: scoped CSS та збереження авторських h2/h3, ролей і назв; browser-перевірка лишається окремим гейтом Claude за брифом.
ДАЛІ: внести presentation/CSS без зміни моделі й поведінкових D-слайсів.

ФАКТ: QW3 реалізовано у `LessonDocumentView.tsx`, `LessonDocumentRichText.tsx` і scoped `main.css`: спільний `.lesson-prose` задає ритм p/list, виразні h2/h3, disc/decimal-маркери, підкреслені посилання й перенесення довгого тексту. Tiptap li > p має нульові відступи для відповідності reader. Обмеження 65ch стосується лише прози; player, row та image не звужено. `.lesson-document-task-label` використовує чинний локалізований ключ `administration.document.task` («Завдання»); підпис — p, авторські заголовки не змінено. Task відрізняється рамкою, потовщеним початковим краєм і відступами.
ФАКТ: QW5 — чотири `.lesson-editor-toolbar-group` (marks/headings/lists/link), flex-wrap усередині й між групами; `[aria-pressed='true']` отримав контрастну рамку та підкреслення. Наявні React Aria Button, текстові назви, мінімальні цілі через `--control-min-size` і focus-стилі збережено. Палітра, шрифти, dark-only admin, світла/темна тема reader — через наявні токени. Нових залежностей, анімацій, декоративних іконок або utility-класів немає.
ДАЛІ: завершити статичну звірку й передати Claude повний гейт за брифом.

ФАКТ: статично звірено обов’язкові e2e-селектори. Збережено heading «Редактор уроку», group «Додати секцію», усі назви палітри/save/publish/move/delete, textbox «Текст»/«Завдання», upload/alt labels, `.lesson-editor-section`, `.lesson-editor-image-preview` на img; region «Матеріали уроку», img alt і незмінний рендер авторського task-тексту «Виконайте вправу». Toolbar role/name і назви семи команд також збережено. Run A та сторонні зміни не відкинуто; комітів немає.
ФАКТ: `oxfmt --check` трьох кодових файлів, `oxlint --deny-warnings` двох TSX, `tsc --noEmit` і `git diff --check` завершились exit 0. Спроба наявного `LessonDocumentView.test.tsx` через rstest заблокована середовищем ДО тестів: `listen EPERM: operation not permitted ::1:3000` (exit 1), не passed. Повний FE-гейт, build, e2e, браузерні скриншоти, перевірки 200% zoom/контрасту/клавіатури не виконувались; за брифом їх виконує Claude. Візуальну відповідність не оголошено перевіреною.
ФАКТ: окремий наявний ризик QW5 — встановлений `@tiptap/react/src/useEditor.ts:360–369` не підписує компонент на transaction rerender, якщо `shouldRerenderOnTransaction` не заданий; тут він не заданий. Тож перенесення каретки без інших React-оновлень може залишати `aria-pressed` застарілим. CSS коректно оформлює атрибут, але синхронізацію selection не доведено браузером і не виправлено: дизайн QW5 прямо відносить це до окремого поведінкового фіксу. Link-кнопка й надалі позначає відкриття наявного поля URL (`addingLink`), не факт link-mark.
ДАЛІ: Claude — повний FE-гейт + повний e2e через shared CSS; браузером звірити prose p→h2→p→h3→list, task, light/dark reader, dark-admin, narrow/Row і 200% zoom, focus та active toolbar при перенесенні каретки. Синхронізація selection — окремий поведінковий follow-up за QW5; QW6 та D-слайси не виконувались. Модель, конвертер, контракти, save/publish/access не змінено.
## Журнал Codex — Run C (перенесено з frontend/editor-polish-run-c-report.md)
## Журнал Codex — Run C

ДАЛІ: 2026-10-13, Codex — реалізувати лише QW6 поверх незакомічених Run A/B: loading/error/empty редактора, зберегти обов’язкові селектори та чинні правила операцій.

ФАКТ: прочитано журнали Run A/B, дизайн-план QW6, брифінг репозиторію, план lesson-builder і work/tasks.md. Розділу `## Run C` у `_wip-editor-polish-codex.md` на час читання не було; обсяг узято з прямого запиту власника та рядка QW6 дизайн-плану.

ФАКТ: `LessonDocumentEditorLoading.tsx` переюзує Skeleton із `ui/states.tsx` (4 рядки, локалізований доступний текст). Його використовують і Suspense fallback маршруту `admin.courses.$courseId_.lessons.$lessonId.tsx`, і завантаження документа у `LessonDocumentEditor.tsx`. Компонент не імпортує editor/Tiptap; lazy-межу збережено за графом імпортів, production bundle тут не перевірявся.

ФАКТ: помилка завантаження використовує ErrorState з чинними текстом і retry→refetch; помилки save/publish/upload — ErrorState з чинним повідомленням, без зміни обробників. Документ із нулем секцій показує EmptyState з підказкою; палітра «Додати секцію» та save/publish лишаються доступними. Додано empty/emptyHint у локалі uk/en. Reader не змінено цим прогоном.

ФАКТ: `main.css` отримав лише scoped правила `.lesson-editor > .ui-state`, `.lesson-editor > .ui-skeleton`, `.lesson-editor > .ui-state--error`, `.lesson-editor > .ui-state > :last-child`, `.lesson-editor-sections:empty`: відступи, рамка, радіус, перенесення тексту й приховування порожнього ol. Кольори/розміри — чинні семантичні токени dark-admin; глобальні примітиви не змінено.

ФАКТ: статично збережено h1 «Редактор уроку» у loading/error/ready, group «Додати секцію» в empty, назви palette/save/publish/move/delete, textbox «Текст»/«Завдання», upload/alt labels, класи секції й img-preview. Run A/B та сторонні зміни збережено. Модель, контракти, конвертер, правила save/publish/access, залежності та D-слайси не змінено. Комітів немає.

ФАКТ: `oxfmt --check` шести змінених/нових кодових файлів, `oxlint --deny-warnings` п’яти TS/TSX, `tsc --noEmit`, `node scripts/check-architecture.mjs` (208 файлів) і `git diff --check` — exit 0. Наявний `LessonDocumentEditor.test.tsx` не стартував: rstest завершився exit 1 на `listen EPERM: operation not permitted ::1:3000`, ДО виконання тестів. Це не passed. Повний ordered gate, build, e2e й браузерна перевірка не виконувались; за брифом їх виконує Claude.

ДАЛІ: Claude — FE-гейт + повний e2e, браузером звірити lazy-loading та query-loading, load-error→retry, порожній документ→додавання секції→видалення останньої, помилки операцій, dark-admin, клавіатуру й 200% zoom. Попередній QW5-ризик selection/aria-pressed не виправлявся.

ФАКТ: дописування до `docs/lesson-builder/_wip-editor-polish-codex.md` заблоковане sandbox: `patch rejected: writing outside of the project; rejected by user approval settings`. Цей файл збережено в дозволеному frontend/ як готовий розділ для перенесення в основний журнал; цільовий журнал не змінено.
