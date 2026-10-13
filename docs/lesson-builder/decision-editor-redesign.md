# Рішення: редизайн редактора уроку — adopt-with-changes

- **Дата:** 2026-10-14
- **Автор:** Courses 2 (рев'ю-роль), на передачу від Courses 1
- **Статус:** прийнято; інтеграція — окремим слайсом на модулі `lesson-builder`
- **Закриває альтернативи:** «взяти код прототипу як є» і «відхилити редизайн»

## Контекст

Власник замовив у зовнішнього design-AI редизайн сторінки редактора уроку
(`/admin/courses/:courseId/lessons/:lessonId`), бо після editor-polish (QW1–QW6, PR #21)
вона все ще виглядає голо. Прийшов робочий React + React Aria + Tiptap прототип «lesson-studio».
Courses 1 відрендерив його, зібрав скрини й прогнав порівняльну оцінку A (прототип) vs B
(поточний код) — `/tmp/.../scratchpad/design-eval-colleague.md`. Courses 2 звірив несучі
твердження оцінки з кодом (нижче «Перевірка»).

## Рішення

**Беремо візуальну мову A, не код A.** Прототип — цільовий вигляд (три зони, типова ідентичність
секції icon-tile + caps-лейбл + індекс, картка з рамкою/тінню, палітра-тайли з описом, pressed-стан
тулбара, reader-foreshadow), але в прод він іде **перекладеним на наявний компонент**
`LessonDocumentEditor.tsx` — реальні токени/ваги, TanStack Query save/publish/upload, i18n,
доменні типи `LessonSection`. Код прототипу як є в прод не йде.

## Чому не код прототипу (перевірено у звірці)

- **Паралельний token-шар:** `dist/styles.css:3` «Proposed tokens: required values absent from
  brief»; **27 розрізнених чужих родин** `--type-*/--size-*/--weight-*/--leading-*/--tracking-*/
  --duration-*/--layer-*`, яких у проді немає (система іменує лише `--color-*/--space-*/--radius-*/
  --shadow-*/--control-min-size/--card-padding`).
- **Сміттєві літерали:** `--size-zero:0`, `--color-transparent:transparent` — індирекція без цінності.
- **Дрейф ваг:** прототип 500/600/700 ≠ реальні 650/700/750 (`main.css:1377,1518,1587,1649`).
- **CTA-дивергенція:** прототип переписує primary через `color-mix` («AA correction: glossy blue»),
  інший, тьмяніший бренд-блакитний, ніж системний `--btn-primary-bg/-highlight/-glow`.
- **Регрес закритого a11y-дефекту:** `src.jsx` ставить `isDisabled` на move-кнопки (й palette/row) —
  це рівно наш закритий gotcha `isdisabled-on-focused-button-drops-focus` (React Aria `isDisabled`
  → native `disabled` → фокус у `<body>`). Поточний B **свідомо** це обходить:
  `LessonDocumentEditor.tsx:112-124` тримає кнопки enabled, вихід за край = no-op, з коментарем.
- **Прото-шасі:** власний AdminShell (sidebar/topbar/breadcrumbs/footer/brand) — у проді вже є;
  save/publish/upload на localStorage; відеоблок placeholder; 200% zoom не перевірявся (README §Обмеження).

## Acceptance (MUST перед інтеграцією)

1. Прибрати `isDisabled` на move/video-tile/row-колонках → enabled + no-op, як у B:112-124.
2. Звести всю візуалку на реальні токени; видалити паралельний `--type-*/--size-*/--weight-*/
   --leading-*/--tracking-*` шар (лишити лише справді відсутні — названим токеном зі значенням);
   ваги 500/600/700 → 650/700/750.
3. Primary-CTA з системних `--btn-primary-bg/-highlight/-glow`, не `color-mix`-ремікс.
4. Видалити `--size-zero`/`--color-transparent` → `0`/`transparent`.
5. Зняти AdminShell прототипу (sidebar/topbar/breadcrumbs/footer/brand) — у проді вже є.
6. Підключити до реальних save/publish/upload (Query, конфлікт-token), i18n і доменних типів —
   не localStorage.

## SHOULD

- Перевірити 200% zoom без горизонтального скролу в реальному браузері.
- Картка секції = `background` + `border` + `shadow`. Токен уже є (`--card-shadow-hover`,
  `--shadow-floating`); поточна `.lesson-editor-section` плаский `--color-surface` без тіні
  (`main.css:1503-1511`). Це **захищений борг (SHOULD), не жорстке порушення стандарту**: секція —
  форм-контейнер, не навігаційна картка каталогу; але спертися на наявну elevation-мову варто.
- `editor ↔ reader` coherence; segmented «Перегляд» рендерить реальний reader-компонент.

## NICE

- Нумерація секцій + лічильник (лише якщо outline реально є на admin-поверхні).
- Статус-рядок із реального стану мутації; task-callout акцент-бордер.
- Декор drag-handle лишити як зарезервований простір (`aria-hidden`), **не** натякати на drag до
  drag-фази (drag-полотно — окрема майбутня фаза, AGENDA §розвилки 5).

## Нюанси рев'ю (уточнення до оцінки A/B)

- B **уже несе** `aria-pressed='true'` на тулбарі (`main.css:1547`) — правильна a11y-семантика.
  MUST щодо pressed-стану — **суто візуальний** (filled замість underline-хака), не «додати семантику».
- «Картка без тіні порушує §5.2» — `§5.2` це нумерація брифа design-AI, не репо-стандарт; тому в
  SHOULD, не MUST (згода з оцінкою).

## Гейт і процес

- Крос-зрізна зміна spільного CSS + редактора → **повний e2e**, не таргетний (урок:
  зміна спільного/персистованого стану ганяється всім e2e-набором). Порядок — `make check`.
- **CI наразі заблоковано білінгом** (GitHub Actions job `verify` не стартує) — зелений CI
  покладатися не можна; хто мерджить — ганяє ordered `make check` локально й кладе run-id у звіт.
- **Pre-launch** (`pre-launch-cut-freely`, власник 2026-10-08): продукт не запущено → візуальну мову
  редактора міняємо прямо, без shim/guard для старих даних.

## Наступний хід

Courses 1 (кодить через Codex) загортає це слайсом на `lesson-builder`: spec-flow (крос-зріз FE)
або серія `to-tickets` з MUST як acceptance, після чого рев'ю + ordered `make check` (Courses 2).
