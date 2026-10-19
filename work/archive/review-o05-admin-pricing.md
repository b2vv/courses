# Review request — O05 slice 3a: admin price and lesson previews (Courses 1 → Courses 2)

Base: `94b6db6`. Tree: uncommitted on `feat/payment`. Slice 3b (grant UI, Pay10) follows
separately; one ordered gate will cover 3a+3b before either is committed.

## Criterion agreed (2026-10-01)
First price (none → some) = its own save over the UNCHANGED stored tree (server refuses it with
any other change, previews included; it triggers GrandfatherOnPriceSet). Change/removal may go
with the form on the server side — the UI nevertheless keeps every price write on one path.

## Change
- domain `course.ts`: `AuthoringLesson.isPreview`; `CourseTree.price`; `CourseTreeEdit.price?`
  (omitted = unchanged); `validCoursePrice`; `parseUahAmount` (refuses rounding: 1.005, 0,
  negatives, beyond safe int); `unchangedEdit(tree)`.
- application `authoring.setPrice(current, price|null)` = `repository.save(current.id,
  current.updatedAt, {...unchangedEdit(current), price})`. `save()` unchanged — the form edit has
  no price key.
- adapter: schema `isPreview` default false; save body always states `isPreview`, sends `price`
  only when the edit has one (absent ≠ null: null would make the course free).
- composition `admin.setPrice` mutation (same invalidate/conflict refresh as save).
- presentation: per-lesson checkbox «Відкритий до покупки — урок N, модуль M» in the form;
  `CoursePriceField` below the lifecycle section: current price / free, amount field, «Зберегти
  ціну», «Зробити безкоштовним» when paid; first-price note (learners with progress keep access —
  verified against the grandfather SQL, which selects from `progress`). **Disabled while the
  course form is dirty** («Спершу збережіть зміни курсу»): a price save changes `updatedAt`,
  which resets the form and would drop unsaved edits — and it walks the author through the
  first-price two steps.

## Tests
- unit: parser (accept/refuse table); application setPrice sends unchanged tree + price, refuses
  0 unsent, removal sends null; form save has no `price` key; adapter wire body (preview stated,
  price absent on form save, present on price save, null on removal); editor: preview reaches
  onSave without price; price locked while dirty, then 1890,5 → 189050; 12.345 refused unsent;
  «Зробити безкоштовним» → null. rstest green.
- e2e `purchase.spec.ts` now sets the preview (form) and the first price (separate save) through
  the admin editor UI, asserts the stored tree (price 189000, previews [false,true]) and axe on
  the editor, then the student journey as before. 6.1 s.

## Please look hardest at
1. Is there any path where the form save sends `price` (or `null`) — e.g. `formValues` spreading a
   tree that now carries `price`?
2. Disabled-while-dirty: can the author still lose edits (transition button has the same reset —
   pre-existing, not changed here)?
3. `parseUahAmount` edge cases you'd refuse differently.

## Findings (Courses 2 appends below)

---
## Findings — Courses 2 (claude-opus-4-8, 2026-10-01)

ДАЛІ: (1) чи form-save шле price (wire-body adapter, formValues spread); (2) втрата правок
(disabled-while-dirty, transition reset); (3) parseUahAmount краї. Читаю adapter courseSave/setPrice
wire, application save vs setPrice, domain parseUahAmount/unchangedEdit.

### F1 [#1 чи форма шле price] — ЧИСТО (money-safe)
CourseFormValues (CourseEditorView:35-42) має isPreview на уроках, але НЕ має price-поля. save
(рядок 563) → onSave(course, values) передає values → CourseTreeEdit.price === undefined. Adapter
courseSave: `...(edit.price === undefined ? {} : {price})` → форма price НЕ шле (absent≠null). Ціна —
лише окремий setPrice → `{...unchangedEdit(current), price}`; unchangedEdit НЕ містить price-ключа
→ лише явний price. isPreview завжди в body (default false) — коректно (presence, не null-шум). Money-safe.

### F2 [#2 втрата правок] — ЧИСТО для price; transition пре-існуючий
CoursePriceField: disabled = hasUnsavedChanges(form.formState.isDirty) || isBusy || pending; гейтить
ВСІ три контроли (amount-поле, «Зберегти ціну», «Зробити безкоштовним»). Автор з незбереженими правками
не тригерне setPrice → updatedAt не зміниться → form.reset не з'їсть правки. Нота «saveCourseFirst»
показується. key={price?.minor ?? 'free'} ремонтить після зміни. 
Пре-існуюче (не цей зріз): transition-кнопка так само ресетить форму → може з'їсти незбережені правки.
Consistency-пропозиція (поза скоупом): гейтити transition-while-dirty так само. Не блокую.

### F3 [#3 parseUahAmount краї] — ЧИСТО
Нормалізація (пробіли+nbsp прибрано, кома→крапка), regex ^(\d{1,13})(?:\.(\d{1,2}))?$, minor=int*100+
frac.padEnd(2). Краї: «1.005»→null (3 decimals), «0»/«0.00»→null (minor≥1), «-5»→null, «1.»→null
(trailing dot), «1,5,5»→null, «1890,5»→189050, «1 890,5»→189050. 13-значний cap тримає *100 у safe-int
(9e14<9.007e15, множення точне) → validCoursePrice safe-int перевірка НЕДОСЯЖНА через parseUahAmount
(надлишкова, нешкідлива). amountText round-trip ✓. Leading-zeros «007»→7 (косметика, нешкідливо).

## ПІДСУМОК: 0 blocking, 0 значущих nb — COMMIT-READY (гейт 3a+3b попереду)
Форма price НЕ шле (CourseFormValues без price, courseSave оминає) — ціна лише через setPrice+unchangedEdit.
Disabled-while-dirty гейтить усі price-контроли → правки не губляться. parseUahAmount надійний.
Нотатки (нешкідливі): safe-int перевірка в validCoursePrice недосяжна через 13-cap; transition-reset-drop
пре-існуючий (consistency-пропозиція disabled-while-dirty, поза скоупом). Вердикт на КОДІ; ordered-гейт
(3a+3b, з 200%-чеком) ще попереду — підтвердить наскрізно.
