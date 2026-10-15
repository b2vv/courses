# editor-polish — tasks (для Codex-імплементації)

План (ЯК): [plan.md](plan.md). Гілка: `feat/editor-polish` (відгалузити від `main` @ `d6ec304`).
Один коміт на таск. Роль: Codex пише код (Courses 1 веде), Courses 2 рев'ю на ключових зрізах,
фінальний PR — owner merge.

## Пастки Codex (дотримати на КОЖНОМУ таску)
- **Format:** Codex пропускає форматування → `npm run format` перед комітом кожного таску (format-check =
  1-й крок гейту, валить на невідформатованому) [[codex-skips-oxfmt-gate-catches]].
- **Impl ↕ тести — окремими прогонами** (великий зріз упирається в ліміт, тести випадають): спершу
  реалізація, resume за id для тестів [[codex-split-impl-and-tests]].
- **Shared-зміни — повним гейтом:** `main.css`, спільний `FileUpload`, обидві локалі зачіпають поза
  зрізом → після T1 і наприкінці ганяти **повний** `make e2e`, не точковий lesson-builder
  [[integration-tag-skipped-by-go-test-and-e2e]].
- **i18n обидві локалі** (`locales/uk.ts` + `en.ts`) — кожен новий ключ в обидві, інакше типи/e2e червоні.

---

## T1 — QW2: image-upload через примітив FileUpload + drop-зона
**Зріз плану:** S1. **Файли:** `ui/controls.tsx`, `admin/LessonDocumentEditor.tsx` (`ImageField`),
`main.css`, `locales/{uk,en}.ts`, `ui/ui.test.tsx`, e2e `lesson-builder.spec.ts` (не ламати).

**Impl (прогін 1):**
1. `FileUpload` (controls.tsx:278-286): додати опційні props `onDrop?: (file: File)=>void`,
   `dropLabel?: string`. Коли задані — обгорнути в `Aria.DropZone` (`className="ui-dropzone"`), у `onDrop`
   через `isFileDropItem` відфільтрувати за MIME (jpeg/png/webp), узяти перший прийнятний файл. Форвардити
   `aria-label` на прихований input (filterDOMProps). Без `onDrop` — поведінка як зараз (відео).
2. `ImageField` (464-473): прибрати сирий `<label><input type=file>`; вжити `<FileUpload
   label=uploadImage aria-label=uploadImage onSelect dropLabel=dropHint
   acceptedFileTypes=['image/jpeg','image/png','image/webp'] onDrop=onUpload />`.
3. CSS `.ui-dropzone`: dashed `--color-border`, `--radius-panel`, padding `--space-3/4`, `min-height ≥
   --control-min-size`, `[data-drop-target] → --color-focus`, фокус-ринг. **Не вживати `--space-7`** (нема).
4. i18n: `administration.document.dropHint` (uk+en). НЕ додавати `fileHint`.

**Тести (прогін 2):** `ui.test.tsx` — FileUpload з `onDrop` рендерить DropZone і викликає onDrop;
без `onDrop` — лише Button. **Критично:** e2e `getByLabel('Завантажити зображення').setInputFiles`
(e2e:123) мусить лишитись зелений — доказ, що `aria-label` осів на input.

**Гейт T1:** format → lint → typecheck → unit → **повний e2e** (shared: controls+css+i18n).
**Готово коли:** A-ED2 — drop+клік працюють, preview обмежене, e2e:123/128 зелені, відео не зламане
(LessonVideoField тести).

---

## T2 — D1: per-image upload-стан + чесний save/publish + no-text-loss фікс
**Зріз плану:** S2 (вкл. B1-фікс + ib1/ib2). **Файли:** `admin/LessonDocumentEditor.tsx`,
`locales/{uk,en}.ts`, `LessonDocumentEditor.test.tsx`.

**Impl (прогін 1):**
1. **Per-image upload:** `uploadingKey: string|null`; `uploadInto(key, file, apply)` ставить
   `setUploadingKey(key)` до mutate, у `onSettled` чистить лише при збігу `setUploadingKey(k=>k===key?null:k)`.
   `ImageField`/`RowBody` отримують `uploading = uploadingKey===ownKey` (ownKey = section.id / column.id).
2. **B1 no-text-loss (ОБОВ'ЯЗКОВО фікс, не лише тест):**
   - token-апдейти (186, 202, 204): `setEdited(cur => cur ? {...cur, token} : {...base, token})` (ib1 —
     синхронний seed без overlay);
   - `setSections` → функційний updater: `(fn) => setEdited(cur => { const b = cur ?? draft; return
     {...b, sections: typeof fn==='function' ? fn(b.sections) : fn}; })`;
   - `updateSection` в upload-apply → дельта-функційна **за id, не index**: `setSections(prev =>
     prev.map(s => s.id===targetId ? {...s, imageId, contentType} : s))`; для row — вкладена дельта за
     `column.id` (ib2: протягнути section.id+column.id у `apply`).
3. **Save/publish success:** `status: {kind:'saved'|'published'}|null`; onSave→saved, onPublish→published;
   рендер `Toast` (role=status) `savedToast`/`publishedToast`, onDismiss чистить. Помилки лишаються
   `ErrorState` (role=alert).
4. i18n: `savedToast`, `publishedToast` (uk+en).

**Тести (прогін 2):** unit — (a) аплоад у A не вмикає «Завантажуємо…» у B; (b) save→status (читати за
**name/hasText**, не голий `getByRole('status')`) без alert; (c) **правка секції під час pending save
збережена** (реально редагувати під pending — B1-регрес); (d) паралельні аплоади: resolve першого не
гасить індикатор другого.

**Гейт T2:** format → lint → typecheck → unit → **повний e2e** (alert count 0 після save/publish).
**Готово коли:** A-BH1 — per-image стан, success-статус, правка не губиться.

---

## T3 — D2a: невалідний URL не видаляє посилання
**Зріз плану:** S3. **Файли:** `admin/LessonDocumentRichText.tsx`, `locales/{uk,en}.ts`,
`LessonDocumentEditor.test.tsx`.

**Impl (прогін 1):** `applyLink` (59-64):
1. Валідний (`httpUrl`): `setLink`, закрити, очистити — як зараз.
2. Невалідний: **НЕ `unsetLink`**; `setLinkError(true)`, поле відкрите, `linkUrl` збережено, фокус у
   поле (ref). `onChange` поля → `setLinkError(false)`.
3. Поле лінку (сирий `<input type=url>` 134-139) → `TextField` з `errorMessage`(linkInvalid при помилці)
   + `ref`.
4. Кнопка «Видалити посилання» (показ при `editor.isActive('link')`) → `unsetLink()`. Прибрати неявний
   `unsetLink` з гілки «невалідний».
5. i18n: `linkInvalid`, `removeLink` (uk+en).

**Тести (прогін 2):** unit — (a) невалідний URL: поле відкрите, errorMessage видимий, лінк цілий
(`isActive('link')` без змін), `linkUrl` збережено, фокус у полі (`document.activeElement`); (b) валідний:
застосовано, поле закрито; (c) «Видалити посилання» знімає лінк.

**Гейт T3:** format → lint → typecheck → unit (RichText локальний, не shared-CSS → повний e2e на фіналі).
**Готово коли:** A-BH2.

---

## T4 — D2b: фокус add/delete/reorder + оголошення руху й межі
**Зріз плану:** S4 (вкл. ib3). **Файли:** `admin/LessonDocumentEditor.tsx`, `locales/{uk,en}.ts`,
`LessonDocumentEditor.test.tsx`, e2e `lesson-builder.spec.ts` (додати асерт, не ламати).

**Impl (прогін 1):**
1. Кожна `lesson-editor-section` → `data-section-id={section.id}`.
2. `pendingFocus` ref + `useEffect` на зміну sections: фокусує контрол секції за id.
   - **add** → фокус у **редаговане поле** нового блоку: text/task→`[role=textbox]`, image→file-кнопка,
     video/row→перша кнопка дій (N6).
   - **delete** → ціль ДО видалення: секція index-1, інакше наступна, інакше кнопка палітри «Текст»
     (не body).
3. **reorder оголошення:** `aria-live` polite (`role=status`, візуально-прихований). Успішний рух →
   `movedTo{pos,total}`; no-op на краю → `moveBlocked{edge}`. Форсити переоголошення ідентичного тексту
   (лічильник-ключ, ib3). Скоуп читання за name (не голий status).
4. i18n: `movedTo`, `moveBlocked` (uk+en).

**Тести (прогін 2):** unit — (a) add→`activeElement`=редаговане поле (не кнопка тулбара); (b) delete
середньої→фокус у сусідній (не body); (c) delete останньої→фокус на палітрі; (d) reorder успішний→
live-region має нову позицію; (e) reorder на краю→live-region має «вже перша/остання». e2e — після
delete `activeElement`≠body (новий асерт, наявний reorder-192 не чіпати).

**Гейт T4:** format → lint → typecheck → unit → **повний e2e**.
**Готово коли:** A-BH3.

---

## Фінальний гейт (перед PR)
1. **Ordered повний** `make check`: format-check → lint → typecheck → test → integration → build → e2e
   (ПОВНИЙ). GATE_EXIT=0 [[background-exit-code-is-last-pipe]] (читати GATE_EXIT, не pipe).
2. **Браузер-a11y** (host-Playwright, бо CDP не дає холістичного SSR-Suspense): скрини editor у
   admin-dark; bbox ≥44px drop-зони/кнопок; фокус-ринг drop-зони; computed-контраст у фактичних токенах
   (не лише axe на зеленому гейті — A-GATE).
3. **Courses 2 рев'ю** ключових зрізів (QW2 примітив, B1-фікс, D2 фокус-пастки).
4. **Один фінальний PR** у `main`; owner візуальний огляд (editor dark) → owner merge (Q3).
   CI Actions заблоковано білінгом → гейт локальний (named), owner розрулює білінг окремо.
