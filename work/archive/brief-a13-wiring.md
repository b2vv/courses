# Brief A13 / крок 4a — wiring C17 і механізм C18 (фронтенд)

Частина A13 (Phase 8): закрити приймання ASR. Кроки 1–3 — бекенд, образ і фікстура. Цей крок —
лише фронтенд `src/` і unit-тести. E2e для C03/C06/C11/C13/C17/C18 — окремий крок 4b.
Бриф — сесія Courses 1, 2026-09-24; рецензент і гейти — Courses 2.

## Стан, від якого починаєш

У дереві лежить **незакомічений WIP невідомого автора** (6 файлів):
`CourseEditorView.tsx`, `LessonTranscriptField.tsx`, `TranscriptEditor.tsx`,
`TranscriptEditor.test.tsx`, `locales/en.ts`, `locales/uk.ts`. Він підключає C17: `onLoadDraft` з
`LessonDraftField` → стан `loadedDrafts` у `LessonFields` → проп `draftOffer` у `TranscriptEditor`,
який показує пропозицію «завантажити / відхилити» тим самим способом, що crash-recovery, без
ремаунту. **Бери його за основу, але як чужий код: спершу прочитай диф (`git diff -- frontend/src`)
і перевір його так само, як перевіряв би свій.** Що не так — виправ, а в звіті назви окремо.

Відомі діри WIP (перевірено проти коду):
1. **Мова чернетки губиться.** `LessonDraftField` викликає `onLoadDraft(draft.segments,
   draft.language)` (`LessonDraftField.tsx:118`), а WIP бере лише `segments`. `TranscriptEditor`
   зберігає під своїм пропом `language`. Англійська чернетка, завантажена в редактор української
   мови, збережеться як українська.
2. **C18 немає.** `draftOffer` не несе токен чернетки (`draft.updatedAt`, `domain/transcript/draft.ts:9`).
   Знімок recovery (`application/transcript/recovery.ts:30-48`) несе лише `expectedUpdatedAt`
   транскрипту. Тому після перегенерації пропозиція відновлення не може стати «застарілою».
3. Дрібниця: IIFE-спред у JSX `CourseEditorView` (`{...(() => {…})()}`) — заміни звичайною змінною.

## Read first
- `docs/asr/spec.md` §12: **C17** (:1203), **C18** (:1204), **C11** (:1197); і D8-правила, на які
  вони посилаються.
- `frontend/src/application/transcript/recovery.ts` — весь (`TranscriptSnapshot`,
  `TranscriptBaseline`, `recoveryOffer` :88-97, де вже є `offer-stale`).
- `frontend/src/presentation/admin/TranscriptEditor.tsx`, `LessonTranscriptField.tsx`,
  `LessonDraftField.tsx`, `CourseEditorView.tsx` (з WIP).
- `frontend/src/domain/transcript/draft.ts`, `frontend/src/ports/DraftRepository.ts`.
- Наявні тести: `TranscriptEditor.test.tsx`, `LessonDraftField.test.tsx`, тести `recovery`.

## Зроби

1. **Мова.** `draftOffer` несе `language` чернетки. Вирішуй, як редактор з ним поводиться:
   зберігати під мовою чернетки чи показувати розбіжність і вимагати вибору. Правило одне — текст
   ніколи не зберігається під чужою мовою мовчки. `auto` не буває мовою чернетки (C11), не вводь його.
   Якщо модель мов у `LessonTranscriptField` робить це неочевидним, опиши вибір у `## Discussion`.

2. **C18 — механізм.** Коли рядки в редакторі походять із завантаженої чернетки, знімок recovery
   несе токен цієї чернетки. Під час відновлення, якщо поточний токен чернетки уроку інший (її
   перегенерували), пропозиція — `offer-stale`, а не звичайна `offer`. Узагальни наявне порівняння
   в `recoveryOffer` на токен чернетки. Не заводь другий паралельний механізм, бо spec прямо каже:
   «тим самим виводом». Поточний токен чернетки має дійти до редактора через композицію
   (`CourseEditorView` уже знає чернетку через `LessonDraftField`). Знімок без токена чернетки,
   тобто ручне редагування, поводиться як досі.

2a. **Один тип пропозиції.** `loadedDrafts` у `CourseEditorView`, `draftOffer` у
   `LessonTranscriptField` і в `TranscriptEditor` несуть той самий набір: сегменти, мову й токен
   чернетки (`updatedAt`). Тип один, а не три часткові копії.

3. **Unit-тести** (Rstest, поряд із наявними):
   - `recoveryOffer`: знімок із токеном чернетки A + поточна чернетка A → `offer`; поточна
     чернетка B → `offer-stale`; знімок без токена чернетки → як раніше.
   - `TranscriptEditor`: `draftOffer` показується лише коли `recovery.offer` = `none`; «завантажити»
     замінює рядки і з мовою чернетки поводиться за правилом п. 1; «відхилити» кличе
     `onDraftOfferHandled`; прийняття пропозиції не перемонтовує редактор.
   - `CourseEditorView`: завантаження чернетки з `LessonDraftField` доходить до редактора того самого
     уроку й не зачіпає інші уроки.

## NOT
- Не пиши e2e (`frontend/e2e/`) — це крок 4b.
- Не чіпай бекенд, `openapi.yaml`, Orval-`generated/`, draft-адаптер
  (`adapters/api/draft-repository.ts`; його контракт — борг B-20, окремо).
- Не редагуй `work/tasks.md`, `docs/asr/report.md`. Git не чіпай.
- AGENTS.md: без `any` і небезпечних assertion-ів; fetch лише в адаптерах; UI через React Aria
  з кіту `../ui`; нові рядки UI — у `en.ts` і `uk.ts` обидва.

## Перевірки
Прожени в `frontend`: `npm run format:check`, `npm run lint`, `npm run typecheck`, `npm test`.
Результат кожної познач чесно. Docker і e2e не запускаєш.

## Якщо бачиш неточність, що дасть баг
`## Discussion` у звіті: що не так, file:line, що станеться в рантаймі, пропозиція, що зробив тим часом.

## Звіт
`work/report-a13-wiring.md` — окремим розділом «Рев'ю чужого WIP»: що в ньому було правильно, що
змінено і чому.
