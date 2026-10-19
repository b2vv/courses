# editor V-пілот — план реалізації (layout fidelity)

**Дата:** 2026-10-15/16. Гілка `feat/editor-layout-fidelity`. Власник: E6 — дротувати метадату (повний).
Спека: `_wip-proto-lesson-spec.md`. Diff: `_wip-editor-diff.md`. Токени наші = proto 1:1
(`--control-min-size`, `--space-tile-min`, `--color-faint`, `--font-utility` mono). Proto dark-only —
переношу ГЕОМЕТРІЮ, не палітру (власник: кольори ок).

## Рішення чесності (датовані 2026-10-16, Courses 1)
- **E6 title-block** = read-only `<h2>` назва уроку (у нашому editor назва редагується деінде — редаговане
  поле було б фейком без save-path). Eyebrow «Назва» + caption «{курс} · {модуль}». Дані з `admin.detail(courseId)`.
- **E8 future-handle (декоративний drag)** — ОМИНАЮ: grip-dots = афорданс drag, якого нема (reorder через
  up/down). Фейковий афорданс порушує чесність (design-pass §3). Прапорець власнику.
- **E3 primary-кнопка** — лише chevron; розмір/заливку НЕ чіпаю (палітра ок; _ORIGIN рішення про розмір не дію).
- **E4 Перегляд** — функційний: segmented перемикає на наявний `LessonDocumentView(document)` (read-renderer).

## Зміни по файлах
1. **i18n** uk.ts+en.ts `administration.document`: eyebrow, intro, modeLabel/modeEditor/modePreview,
   savedStatus/unsavedStatus, titleLabel, previewEmpty, linkHeading/linkCancel/linkConfirm/linkClose.
2. **LessonDocumentRichText.tsx**: toolbar→B/I/H2/H3 текст-спани + list/ordered/link іконки, Tooltip на кожній,
   toolbar-group-start на H2+list, toolbar-caption (тип секції) справа. Link inline→Dialog (Aria DialogTrigger/
   ModalOverlay/Modal/Dialog, tooltip-trigger, .ui-dialog-actions Cancel+Confirm). ЗБЕРЕГТИ PR#33: invalid лишає
   відкритим+focus, valid apply, no-text-loss, remove-link, announce.
3. **LessonDocumentEditor.tsx**: prop `meta?`. PageHeader (eyebrow+h1+intro+actions, chevron). mode state
   edit|preview. `.lesson-editor-meta` (segmented + save-status dot+count). title-block (meta). preview→
   LessonDocumentView. section-bar: .section-identity (icon-tile+type+padded `<small>01</small>`)+actions
   (up/down/delete як quiet icon-buttons, delete не danger-red, boundary aria-disabled не native). dirty-tracking
   для save-status.
4. **route** admin.courses.$courseId_.lessons.$lessonId.tsx: useQuery(admin.detail(courseId))→resolveLessonMeta→meta.
5. **CSS** main.css `.lesson-editor-*`: eyebrow, headings, meta, save-status(+dot), title-block, section-identity,
   padded-index, toolbar(icon+caption+group), video empty center. Primitives (segmented/modal/tooltip) вже стилізовані.
6. **tests** LessonDocumentEditor.test.tsx: оновити під новий DOM, ЗБЕРЕГТИ поведінкові асерти (link через діалог).

## Гейт
`make check` (видалити `frontend/.proto-render` ПЕРЕД). Браузер: proto↔prod 1440×900 + 390×844, axe@200%, клавіатура.
