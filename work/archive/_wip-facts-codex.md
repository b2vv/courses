# editor-polish — факти стану коду (зібрав Explore-сабагент 2026-10-15, read-only)

## 1. UI-примітиви (усі ІСНУЮТЬ, `presentation/ui/`)
- `FileUpload` controls.tsx:278-286 (label, isDisabled, Aria.FileTriggerProps/onSelect; рендерить Aria.FileTrigger + secondary Button). **Редактор його НЕ вживає** — сирий `<input type=file className="ui-input">` (LessonDocumentEditor.tsx:464-473).
- `TextField` controls.tsx:47-66 (label обов., description, errorMessage, placeholder; root .ui-field, input .ui-input).
- `Skeleton` states.tsx:39-49 (label, lines 1..10). `ErrorState` states.tsx:22-35 (title, retry-union, role=alert). `EmptyState` states.tsx:6-13 (title, children).
- `Tooltip` overlays.tsx:116-123 (children ReactElement, content; delay 500). `Toast` overlays.tsx:127-139 (message, onDismiss; aria-live polite).
- `Button` controls.tsx:8-10 (variant primary|secondary|danger; onPress/isPending/isDisabled).
- Окремого Spinner нема; редактор має власний `LessonDocumentEditorLoading` (import :19).

## 2. Компоненти (рядки): Editor 479, RichText 149, tiptap-document 133, View 143. CSS-in-JS нема — усі класи глобальні в main.css.
- Editor класи: lesson-editor(-header/-actions/-icon-tile/-sections/-section[+data-type]/-section-bar/-section-type/-section-index/-section-actions/-add/-palette/-palette-label/-palette-description/-video/-row/-columns[+data-columns]/-column/-column-bar/-add-column/-image/-image-preview/-upload). alt через TextField (476). Кнопки — ui Button. sectionTypes video|text|image|row|task; колонок max 3 (text|image).
- RichText: lesson-editor-richtext(66)/-toolbar(68 role=toolbar)/-toolbar-group(×4)/-link(131)/-input(146). Внутрішній prose-клас `lesson-prose` через editorProps (54) + role=textbox/aria-multiline/aria-label. Tiptap StarterKit heading [2,3], вимкнено strike/code/codeBlock/blockquote/hr; Link openOnClick:false. Поле лінку — сирий `<input type=url>` (134).
- Reader (View): lesson-document(28 section aria-label="Матеріали уроку")/-text(41)/-task(47 +data-section=task)/-task-label(48)/-image(54,75)/-row(60)/-column(70)/`lesson-prose`(83). Семантичний HTML p/h2/h3/ul/ol/li/a(target=_blank). video-секція → null (плеєр грає).

## 3. Scoped CSS СЬОГОДНІ (main.css, 2600 рядків) — КЛЮЧОВЕ
Урок-блок ~1412-1766. **РЕДАКТОР уже має великий CSS-блок (не «голий MVP»):**
- READER lesson-document*: 1412-1419, task 1479-1486, task-label 1487-1491, image img 1493-1497, row 1499-1503, column 1510-1513.
- СПІЛЬНА prose `:is(.lesson-document,.lesson-editor-input) .lesson-prose`: 1421-1477 (h2/h3, списки, посилання, focus-visible). Спільний контейнер .lesson-document,.lesson-editor 1505-1508 (container inline-size).
- EDITOR lesson-editor*: **1515-1765** — .lesson-editor 1515; стани >.ui-state/.ui-skeleton 1519-1532; header/actions/bars 1536-1547; header 1549-1555; sections 1556-1562; section 1563-1572; section-bar 1573-1576; section-type 1577-1584; section-actions 1585-1587; icon-tile 1588-1597; section-index 1598-1603; palette 1604-1624; richtext/image/row 1626-1632; toolbar 1633-1644; toolbar .ui-button + [aria-pressed=true] 1645-1654; input 1655-1670 (contenteditable 1658-1669); .ui-button 1671-1675; image-preview 1676-1681; upload/link 1682-1717 (::file-selector-button 1693); columns/column 1719-1735; add fieldset+legend 1736-1751; video 1752-1755. Container-query data-columns 2/3: 1757-1765.
- `.ProseMirror`/`.tiptap` селекторів НЕМА (стиль через .lesson-prose + [contenteditable]).

## 4. Токени (визначені)
tokens.css: --space-1..6 (46-51), --space-8 (52). **--space-7 ВІДСУТНІЙ** (шкала 6→8 — НЕ вживати --space-7). --color-background/surface/text/muted/border/hairline/focus (3-16, dark @media 79-89 + [data-theme=dark] 99-109). --radius-panel :43 (0.75rem). **min-контроль = --control-min-size :59 (2.75rem=44px)** (НЕ --control-min-height). --font-body 37-39, --font-display 40.
tokens.admin.css [data-surface=admin]: --font-display 19, --radius-control 20, --card-padding 21, --card-radius 22, --surface-gap 25; палітра dark безумовно 29-69 (--radius-panel/--space-* успадковані).

## 5. e2e — НЕ ламати (lesson-builder.spec.ts, 242 рядки)
- Т1 (97-166): heading 'Редактор уроку'; group 'Додати секцію' → button 'Текст'/'Зображення'/'Завдання'/'Відео'; textbox name='Текст'/'Завдання' (rich-text accessible name = тип); getByLabel('Завантажити зображення').setInputFiles; **locator('.lesson-editor-image-preview')**; getByLabel('Опис зображення'); button 'Опублікувати'/'Зберегти чернетку'; **getByRole('alert') toHaveCount 0** після save/publish; reader region 'Матеріали уроку' + getByText + img name=alt.
- Т2 (168-206): **locator('.lesson-editor-section') toHaveCount**; button 'Пересунути вгору'; Enter рухає; sections.first() data-type='image'; activeElement лишається button; **axe wcag2a/2aa/21aa @200% == []**.
- Т3 (208-241): reader heading 'Зміст курсу'; неопублікований текст toHaveCount 0.
- Інші спеки lesson-editor*/lesson-document* НЕ асертять (лише плеєр .lesson-player-controls в accessibility.spec 284-349).

## 6. Admin dark (tokens.admin.css)
[data-surface=admin] скоуп (admin+instructor контейнер); color-scheme:dark (:30); dark-палітра БЕЗУМОВНО на скоупі 29-69, ігнорує prefers-color-scheme/[data-theme] (owner 2026-10-09, коментар 10-13).

---
## ВИСНОВОК (Courses 1, 2026-10-15): преміса аналізу 2026-10-13 ПРОТУХЛА
Editor-CSS (1515-1765) — це робота **admin-restyle S3** (`639c7ca`, 2026-10-14), що відбулась ПІСЛЯ
аналізу. Більшість візуального QW-набору вже відвантажено:
- QW1 межі/картки секцій — **є** (section 1563, section-bar, icon-tile, borders).
- QW4 палітра tiles+підписи — **є** (palette 1604-1624, add fieldset+legend 1736-1751).
- QW5 toolbar групи + active-state — **є** (toolbar-group, [aria-pressed=true] 1645-1654).
- QW6 loading/error/empty editor — **є** (>.ui-state/.ui-skeleton 1519-1532 + власний Loading).
- QW3 reader prose + task — **є** (.lesson-prose 1421-1477, task 1479-1491).
Лишилось, що S3 ймовірно НЕ закрив (візуальний прохід, не поведінка):
- **QW2 image: сирий `<input type=file>` замість примітиву FileUpload** (є лише ::file-selector-button стиль) — треба звірити оком, чи це проблема.
- **D1 чесні save/publish/upload-статуси** (поведінка) — ймовірно відкрито.
- **D2 URL-помилка не має видаляти посилання + фокус після add/delete/reorder** (поведінка) — ймовірно відкрито.
ДІЯ: звірити поточний вигляд оком (host-Playwright скриншоти editor+reader) + прочитати CSS/поведінку,
тоді переузгодити скоуп фази з власником — бо половину QW вже зроблено в S3.
