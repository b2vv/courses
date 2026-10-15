# editor-polish — план (ЯК)

Фаза: **QW2 + D1 + D2** (скоуп переузгоджено 2026-10-15; візуал QW1/3/4/5/6 уже відвантажив
admin-restyle S3 `639c7ca`). Фронт-онлі: лише `presentation/` + scoped CSS у `main.css` + i18n обох
локалей. Бекенд / контракти / домен / схема / Tiptap round-trip НЕ чіпаються.

Специфікація (ЩО/ЧОМУ): [spec.md](spec.md). Факти коду: [_wip-facts-codex.md](_wip-facts-codex.md).
Жива нота: [_wip-spec-flow.md](_wip-spec-flow.md). Гілка: `feat/editor-polish` → один фінальний PR у
`main` (owner merge). Імплементація — Codex (роль Courses 1); рев'ю зрізів — Courses 2.

---

## 0. Reuse-first інвентар (нічого нового, крім drop-зони на наявному примітиві)

| Потреба | Уже є | Де |
|---|---|---|
| Клік-завантаження файлу | `FileUpload` (FileTrigger+Button) | `ui/controls.tsx:278-286` |
| Drag-drop зона | `Aria.DropZone` + `isFileDropItem` | `react-aria-components` (рантайм-підтверджено, НЕ нова залежність) |
| Поле з помилкою/описом | `TextField` (`errorMessage`, `ref`) | `ui/controls.tsx:47-66` |
| Чесний статус-месидж | `Toast` (`<output>` role=status, aria-live polite) | `ui/overlays.tsx`; приклад ужитку `AdminShell.tsx:145` |
| Помилка-alert | `ErrorState` (role=alert) | `ui/states.tsx:22-35` |
| 44px / фокус-ринг / токени | `--control-min-size` (2.75rem), `--color-*`, `--radius-panel`, `--space-1..6,8` | `tokens.css`; **`--space-7` відсутній — не вживати** |

Нових рантайм-залежностей — **нуль** (spec §Out-of-scope). Нових доменних типів — нуль.

---

## 1. Blast radius / спільний стан (гейтимо повним e2e, не точковим)

- **`FileUpload` вживається ще в `LessonVideoField.tsx:242,279`** (відео-аплоад). → drop-зону додаємо
  **опційним** prop (`onDrop?` / `dropLabel?`); без нього примітив поводиться як раніше → відео не
  зачеплено. Регрес відео перевіряємо його тестами + e2e.
- **`main.css` спільний** (глобальні класи, imports tailwind — не чіпати, лише дописати scoped-блоки).
  Пастка admin-restyle «admin-специфіку тягнуть у глобальний примітив і звуть scoped» — drop-зона має
  бути нейтральним примітивом (не admin-only хак).
- **i18n: обидві локалі** (`locales/uk.ts` + `en.ts`) — нові ключі додати в обидві, інакше e2e/типи
  червоні.
- **e2e-контракт НЕ ламати** (`e2e/lesson-builder.spec.ts`): `getByLabel('Завантажити зображення')
  .setInputFiles` (123), `.lesson-editor-image-preview` toBeVisible (128), `getByRole('alert')
  toHaveCount 0` після save/publish (142, 226), reorder → `activeElement` лишається button не body (192).

---

## 2. Зрізи (commits на `feat/editor-polish`)

### S1 — QW2: image-upload через примітив + drop-зона
**Файли:** `ui/controls.tsx` (розширити `FileUpload`), `admin/LessonDocumentEditor.tsx` (`ImageField`
462-474 → примітив), `main.css` (drop-зона scoped), `locales/{uk,en}.ts` (нові ключі), `ui/ui.test.tsx`
(тест drop-зони примітиву).

**ЯК:**
1. `FileUpload` отримує опційні props: `onDrop?: (file: File) => void`, `dropLabel?: string`. Коли
   задані — обгортає Button у `Aria.DropZone` (`className="ui-dropzone"`), у `onDrop` розбирає
   `items`/`isFileDropItem`, **фільтрує за MIME** (jpeg/png/webp — як native `accept` у клік-шляху),
   бере перший прийнятний файл, решту/невалідний тип ігнорує (клієнтський фільтр, не лише серверний).
   Коли не задані — рендер як зараз (відео не зачеплено). Клік-шлях (FileTrigger) лишається завжди.
2. `ImageField` викидає сирий `<label><input type=file>`; вживає `<FileUpload label={t('…uploadImage')}
   aria-label={t('…uploadImage')} onSelect={…} onDrop={onUpload} dropLabel={t('…dropHint')}
   acceptedFileTypes={['image/jpeg','image/png','image/webp']} />`. `onSelect` бере `FileList[0]`, як
   робив `onChange`.
3. **Ризик-контракт — РОЗВ'ЯЗАНО емпірично (2026-10-15, T1 гейт):** гіпотеза «aria-label на input»
   СПРОСТОВАНА probe-тестом: RAC `FileTrigger` викликає `filterDOMProps(rest,{global:true})`, а `global`
   пропускає лише `id`/`data-*`, **НЕ `aria-*`** (для них потрібен `labelable:true`, якого FileTrigger не
   вмикає) → input лишається без accessible name, хоч aria-label передано. **Рішення (датоване, §7):**
   accessible name несе видима **Button** (текст «Завантажити зображення») — це коректніша a11y, ніж
   прихований лейблений input; e2e таргетить прихований input директивно
   (`.lesson-editor-image input[type="file"]`). Мертвий `aria-label` з usage прибрано.
4. **Невалідний тип при drop** (клієнтський фільтр у кроці 1) ігнорується тихо; якщо попри фільтр
   невалідний файл дійде (клік-шлях), серверна валідація дає `unsupported_image`/`invalid_image`
   → наявний `ErrorState` (role=alert). Жодної нової error-поверхні.
5. CSS `.ui-dropzone`: рамка-dashed `--color-border`, `--radius-panel`, padding `--space-3/4`,
   `min-height` ≥ `--control-min-size`, `[data-drop-target] → --color-focus`; фокус-ринг наявний.
   alt-поле (`TextField`) і preview (`.lesson-editor-image-preview`) лишаються.

**i18n нові:** `administration.document.dropHint` («Перетягніть зображення сюди або оберіть файл»).
(Ключа «fileHint» НЕ додаємо — без визначеного споживача; додати лише коли буде відрендерений.)

**a11y:** клавіатура (FileTrigger-кнопка фокусується, Enter/Space), 44px (кнопка + зона),
видимий drop-таргет не лише кольором (рамка), збережені назви.

**Тести:** `ui.test.tsx` — FileUpload з `onDrop` рендерить DropZone і викликає onDrop з файлом;
без `onDrop` — лише Button (відео-режим). e2e `lesson-builder` зелений (setInputFiles по новому контролу).

---

### S2 — D1: per-image upload-стан + чесний save/publish-статус
**Файли:** `admin/LessonDocumentEditor.tsx`, `locales/{uk,en}.ts`, `LessonDocumentEditor.test.tsx`.

**ЯК:**
1. **Per-image upload** (A-BH1 «не спільний»). Зараз `uploading={upload.isPending}` тече в КОЖНЕ
   `ImageField` (272,347,405) → усі показують «Завантажуємо…» під час одного аплоаду. Фікс: тримати
   `uploadingKey: string | null` (id секції/колонки, що вантажиться). `uploadInto` приймає `key`,
   ставить `setUploadingKey(key)` до mutate, а в `onSettled` чистить **лише якщо збігається**:
   `setUploadingKey(k => (k === key ? null : k))` — інакше resolve першого аплоаду скине індикатор
   другого (один `useMutation`, послідовні виклики). `ImageField`/`RowBody` отримують
   `uploading = uploadingKey === ownKey` замість глобального прапора. Ключ = `section.id` (для row —
   `column.id`), стабільний.
2. **Save/publish success** (A-BH1 «показують success саме виконаної операції»). Додати
   `status: {kind:'saved'|'published'} | null` state. `onSave.onSuccess` → `setStatus({kind:'saved'})`;
   `onPublish` після publish → `setStatus({kind:'published'})`. Рендер `Toast` (role=status, НЕ alert —
   тому `getByRole('alert') toHaveCount 0` лишається зелений) з `message=t('…savedToast'/'…publishedToast')`,
   `onDismiss` чистить. Помилки лишаються `ErrorState` (role=alert). **save/publish isPending** лишаємо
   на кнопках (вже є `busy`), але success тепер видимий.
3. **No-text-loss** (A-BH1 остання теза) — **ФІКС коду, НЕ «вже коректно»** (GATE 2 B1, підтвердили
   обидва рецензенти + звірка). Баг: `onSave.onSuccess` (186) робить `setEdited({ ...draft, token })`,
   де `draft` (112) — **стале замикання** з рендеру, в якому викликано `onSave`. Якщо автор редагує
   ПІД ЧАС pending save, новий рендер робить `setEdited(новіСекції)`, але in-flight `onSuccess`
   безумовно перезаписує стан `{...стале draft, token}` і **відкочує** правку — рівно той A-BH1-сценарій,
   що фаза має закрити. Те саме в `onPublish` (202, 204, двічі). Той самий клас — `uploadInto.onSuccess`
   → `apply` → `updateSection` → `setSections` (131, 164) читає стале `draft.sections`.
   **Фікс:** усі `setEdited` в async-колбеках — функційні від поточного стану, не від стале `draft`:
   - token-апдейти (186, 202, 204): `setEdited(cur => (cur ? { ...cur, token } : cur))`;
   - `setSections` переписати на функційний updater, щоб async-шляхи читали поточні секції:
     `const setSections = (fn) => setEdited(cur => { const b = cur ?? draft; return { ...b,
     sections: typeof fn === 'function' ? fn(b.sections) : fn }; })` (синхронні add/remove/move з масивом
     лишаються коректні);
   - **upload-apply мусить бути дельта-функційним за id, не стейл-масивом за index:** колбек `apply`
     застосовує лише свою дельту до ПОТОЧНОЇ секції, знайденої за `id` (не `index` — reorder під час
     аплоаду зсуває index): `setSections(prev => prev.map(s => s.id === targetId ? { ...s, imageId,
     contentType } : s))`. Інакше `setSections(функційний)`, але переданий стейл-масив від `draft.sections.map`
     усе одно відкотить (напр. alt, змінений під час аплоаду). Для row-колонки — вкладена дельта за
     `column.id`.
   Тест пишемо **після** фіксу (інакше чесний тест упаде), і він мусить **реально редагувати під час
   pending save**, не лише перевіряти token.

**i18n нові:** `administration.document.savedToast` («Чернетку збережено»),
`administration.document.publishedToast` («Урок опубліковано»).

**Тести:** unit — (a) аплоад у секцію A не вмикає «Завантажуємо…» у секції B; (b) після save
зʼявляється status-повідомлення (скоупити запит за accessible name / hasText, НЕ голий
`getByRole('status')` — у редакторі ≥2 status-регіони: success-Toast + reorder-live, GATE 2 N2) і НЕ
зʼявляється alert; (c) **правка секції під час pending save зберігається** після resolve (B1-регрес);
(d) паралельний аплоад двох зображень: resolve першого не гасить індикатор другого (N5). e2e
`lesson-builder` зелений (alert count 0 після save/publish).

---

### S3 — D2a: невалідний URL не видаляє посилання
**Файли:** `admin/LessonDocumentRichText.tsx`, `locales/{uk,en}.ts`, `LessonDocumentEditor.test.tsx`.

**ЯК (`applyLink` 59-64 — зараз невалідний URL робить `unsetLink()` + закриває поле + чистить):**
1. Валідний (`httpUrl`): `setLink`, закрити поле, очистити — як зараз.
2. **Невалідний: НЕ `unsetLink`**; `setLinkError(true)`, поле лишається відкритим, `linkUrl`
   НЕ чиститься, фокус повертається в поле (ref на input). (A-BH2 «лишається відкритим із поясненням,
   посилання не зникає».)
3. Поле лінку (сирий `<input type=url>` 134-139) → `TextField` з `errorMessage` (показ при
   `linkError`) + `ref` (для фокусу) + `type` лишаємо url через проп. Помилка чиститься при повторному
   вводі (`onChange` → `setLinkError(false)`).
4. **Видалення посилання — окрема явна дія** (A-BH2 остання теза): коли на виділенні активний лінк
   (`editor.isActive('link')`), показати кнопку «Видалити посилання» → `unsetLink()`. Прибрати
   неявний `unsetLink` з гілки «невалідний».

**i18n нові:** `administration.document.linkInvalid` («Адреса має починатися з http:// або https://»),
`administration.document.removeLink` («Видалити посилання»).

**Тести:** unit — (a) невалідний URL: поле відкрите, errorMessage видимий, `isActive('link')` без змін,
`linkUrl` збережено, фокус у полі; (b) валідний URL: лінк застосовано, поле закрито; (c) кнопка
«Видалити посилання» знімає лінк. (jsdom: фокус перевіряти `document.activeElement`.)

---

### S4 — D2b: фокус-ладування add / delete / reorder + оголошення позиції
**Файли:** `admin/LessonDocumentEditor.tsx`, `locales/{uk,en}.ts`, `LessonDocumentEditor.test.tsx`,
`e2e/lesson-builder.spec.ts` (додати асерт, не ламати наявні).

**ЯК (A-BH3):**
1. **add** → фокус у новій секції. `addSection` після вставки — перевести фокус на **редаговане поле**
   нового блоку (не на першу кнопку тулбара): для text/task — rich-text textbox (`[role=textbox]`),
   для image — file-кнопку, для video/row — першу кнопку дій секції. Механізм: `pendingFocus` ref = id
   новоствореної секції; `useEffect` на зміну sections знаходить `[data-section-id=…]` і фокусує його
   редагований контрол за цим пріоритетом, чистить ref. (Imperative після рендеру — DOM-вузол існує лише
   після вставки.) (GATE 2 N6: «перший контрол у DOM» = кнопка Bold, не поле — тому ціль задано явно.)
2. **delete** → фокус на сусідній блок або палітру, **не body**. `removeSection(index)` до видалення
   обчислює ціль: секція `index-1` (її контрол дій) якщо є, інакше наступна, інакше кнопка палітри
   «Текст». Той самий `pendingFocus`-механізм по id/селектору; палітра має стабільний якір.
   (Пастка [[isdisabled-on-focused-button-drops-focus]]: не вимикати кнопки — фокус злітає; тут вузол
   зникає при видаленні, тож фокус переносимо явно.)
3. **reorder** → фокус лишається на тому самому контролі. Стабільний `key={section.id}` вже зберігає
   DOM-вузол кнопки ↑/↓ (тому наявний e2e 192 зелений). Додати **оголошення руху й межі** (GATE 2 B2 —
   spec A-BH3 вимагає «позиція/межа», план НЕ звужує): `aria-live` polite region (`role=status`,
   візуально-прихований). **Рішення (див. §Рішення GATE 2):** оголошуємо ОБА випадки —
   успішний рух `t('…movedTo', {pos, total})`, а no-op на краю — `t('…moveBlocked', {edge})`
   («секція вже перша/остання»), бо для клавіатури/скрінрідера тиша на краю = незрозуміло, чому дія не
   спрацювала. Live-region скоупити за name (не голий `getByRole('status')`, GATE 2 N2); НЕ role=alert
   (щоб e2e alert count 0 жив).
4. Кожна `lesson-editor-section` отримує `data-section-id={section.id}` для адресації фокусу.

**i18n нові:** `administration.document.movedTo` («Секцію переміщено: позиція {{pos}} з {{total}}»),
`administration.document.moveBlocked` («Секція вже {{edge}}» — edge: «перша»/«остання»).

**Тести:** unit — (a) add → `activeElement` = редаговане поле нової секції (не кнопка тулбара, N6);
(b) delete середньої → фокус у сусідній секції (не body); (c) delete останньої → фокус на палітрі;
(d) reorder успішний → live-region містить нову позицію; (e) reorder на краю (no-op) → live-region
містить «вже перша/остання» (B2). Live-region читати за name/hasText, не голим `getByRole('status')`.
e2e — додати перевірку, що після delete `activeElement` ≠ body (новий асерт поряд із наявним
reorder-192; наявні НЕ чіпати).

---

## 3. Side-effects / міграція / rollback

- **Міграція даних — нема** (фронт-онлі, формат документа незмінний). Жодних змін схеми/контрактів.
- **Side-effects:** спільний `FileUpload` (опц. prop — відео backward-compatible), спільний `main.css`
  (дописані scoped-блоки, наявні не редагуються без потреби), обидві локалі. Усе — гейт повним e2e.
- **Rollback:** кожен зріз — окремий коміт на гілці; відкат = revert коміту. Фінальний PR не merge-иться,
  поки owner не дав візуальний сигнал (Q3). Гілка не чіпає `main` до merge.
- **Toast lifecycle:** success-Toast самознятний по `onDismiss`; не блокує; не role=alert.

## 4. Конституція (інлайн — skip делегації `acceptance-spec`, зріз малий, назвав у `_wip-spec-flow.md` §ПОРІГ)

- **Шари (AGENTS §Architecture):** зміни лише в `presentation/` (ui-примітив + два admin-компоненти) +
  CSS + i18n. Домен/застосунок/порти/адаптери/контракти — не торкаємось. ✅
- **TypeScript:** strict, без `any`; зовнішні дані (drop `items`) валідуємо через `isFileDropItem`
  перед читанням файлу. ✅
- **Дизайн-токени:** семантичні токени, 44px (`--control-min-size`), видимий фокус, dark (editor
  dark-only — перевірка в admin-dark), 200% зум без h-scroll (reader не чіпаємо цією фазою — лише
  editor-поведінка). **`--space-7` не вживати** (відсутній). ✅
- **Без нових залежностей** (DropZone — уже в RAC). ✅
- **Глобально-лічене (AGENTS §Decisions):** drop-зона — нейтральний примітив, не admin-only; новий
  `role=status` live-region — скоупити читання (e2e шукає alert, не status). ✅

## 5. Acceptance (інлайн; перевірка — спостережувана)

| # | Критерій | Спостережуваний чек |
|---|---|---|
| A-ED2 | Image через доступне поле + drop; довге імʼя/alt не розпирає; preview обмежене | e2e: `getByLabel('Завантажити зображення').setInputFiles` зелений; drop файлу → preview; host-скрин вузького екрана |
| A-BH1 | Чесний статус save/publish; upload прив'язаний до конкретного зображення; правка не губиться | unit: аплоад у A не чіпає B; save→status за name без alert; **правка секції під pending save збережена (B1-фікс)**; паралельні аплоади (N5). e2e: alert count 0 |
| A-BH2 | Невалідний URL лишає поле відкритим + пояснення, лінк не зникає; видалення — окрема дія | unit: невалідний→errorMessage+лінк цілий+фокус у полі; кнопка «Видалити посилання» знімає лінк |
| A-BH3 | add→фокус у редагованому полі блоку; delete→сусід/палітра не body; reorder→той самий контроль + **позиція І межа оголошені** | unit a–e; e2e: delete→activeElement≠body; наявний reorder-192 зелений |
| A-GATE | Ordered `make check` зелений; shared-зміни доведені ПОВНИМ e2e; a11y в браузері (44px bbox, фокус) | GATE_EXIT=0; host-Playwright скрини editor-dark; bbox drop-зони/кнопок |

## 6. Верифікація (порядок AGENTS §Verification)

`make format-check` → `lint` → `typecheck` → `test` → `integration` → `build` → **`e2e` (ПОВНИЙ, не
точковий lesson-builder)**. Плюс host-Playwright скрини editor (admin-dark) для A-ED2/A-BH візуалу +
bbox 44px. CI Actions заблоковано білінгом → гейт **локальний** (named у handoff), owner розрулює білінг
окремо.

## 7. GATE 2 — вердикт і рішення (2026-10-15)

Метод: два незалежні свіжі gate-reviewer (Opus 4.8) — `plan-critique` (4 лінзи) + `plan-defense`
(5 питань), нейтральні формулювання (обхід safeguard). Робочі файли: `_wip-gate2-critique.md`,
`_wip-gate2-defense.md`. **Перший прохід: CHANGES-NEEDED** (critique 2 blocking, defense 1 blocking —
B1 підтвердили ОБИДВА незалежно → не поділена сліпа пляма; звірено з кодом).

**Закрито в цьому плані:**
- **B1 (blocking, обидва)** — no-text-loss стале замикання → S2 крок 3 переписано на функційний
  `setEdited(cur=>…)` у 186/202/204 + `setSections` як updater (upload-apply шлях). Тест пишеться
  після фіксу й реально править під pending.
- **B2 (blocking, critique)** — reorder-оголошення → S4 крок 3 розвʼязує «АБО» на користь
  оголошення І руху, І межі; acceptance A-BH3 відновлено до «позиція І межа».
- **N1** (e2e-label) → S1 крок 2/3: `aria-label` на input явно, не «за потреби».
- **N4** (drop mime) → S1 крок 1/4: клієнтський MIME-фільтр + названий серверний fallback.
- **N5** (паралельні аплоади) → S2 крок 1: `setUploadingKey(k=>k===key?null:k)` + тест (d).
- **N6** (ціль add-фокусу) → S4 крок 1: редаговане поле, не кнопка тулбара.
- **N7** (спекулятивний `fileHint`) → прибрано зі S1 i18n.

**Рішення GATE 2 (датовані артефакти, AGENTS §Decisions):**
- 2026-10-15, Courses 1: **reorder на краю (no-op) оголошується** (`moveBlocked`), не замовчується —
  бо тиша на межі для скрінрідера нечитабельна; закриває A-BH3 «межа оголошується» без звуження.
- 2026-10-15, Courses 1 (T1 гейт): **RAC FileTrigger не лейблить прихований input** (filterDOMProps
  `global` не пропускає aria-*, доведено probe-тестом + e2e-timeout). Accessible name несе видима Button;
  e2e `lesson-builder.spec.ts:123` оновлено з `getByLabel(...)` на
  `.lesson-editor-image input[type="file"]` — RAC-семантика, не обхід. A-ED2 «підписи» виконує Button.
- 2026-10-15, Courses 1: **drop-зона лишається в QW2** (N3) — це owner-стер 2026-10-15 («в RAC є
  DropZone»), зафіксований у `_wip-spec-flow.md` §переузгодження; spec.md A-ED2/QW2 оновлено, щоб
  відбити це (spec↔plan узгоджено), drop на СПІЛЬНОМУ `FileUpload` — опційний prop (відео не зачеплено).

**N2** (дві role=status) і скоуп status-запитів — враховано в описах тестів S2/S4 (читати за name).

**Re-check: PASS** (2026-10-15, третій gate-reviewer — B1/B2 закриті без нової діри). Три не-блокуючі
impl-ноти для Codex (гейт зелений і без них):
- **ib1** (B1 token-seed): форма `setEdited(cur => cur ? {...cur, token} : {...base, token})` сидить token
  синхронно і без overlay (інакше null-гілка покладається на async-рефетч від `invalidateQueries` —
  самолікується, але синхронний seed чистіший). A-BH1 стосується `sections`, не `token` — тому це покраща,
  не вимога.
- **ib2** (row upload-apply): щоб дельта-за-id спрацювала в row-колонці, колбек `apply` мусить знати
  `section.id` І `column.id` (зараз `(imageId,contentType)=>void` замикає index). Протягнути обидва id
  у сигнатуру apply — інакше B1-фікс для row лишиться стейл.
- **ib3** (reorder re-announce): повторне ↑ на краю дає ідентичний `moveBlocked`-текст → aria-live може
  не переоголосити. Форсити переоголошення (лічильник-ключ або скидання тексту) — a11y-нюанс.
