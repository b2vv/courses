# Brief A13 / крок 4e — «поточний ассет» (7-й дефект) і дві правки e2e

Частина A13 (Phase 8). Знайдено в e2e-прогоні 4d (Courses 2, 2026-09-24) і розібрано за кодом і
трасою сесією Courses 1. Бриф — Courses 1; рецензент, перезбірка образів і ре-прогін — Courses 2.

## (A) Дефект: селектори ASR беруть будь-який ready-ассет, а не поточний

Після заміни відео урок свідомо тримає **два** ready-ассети без `evicted_at`: попередню генерацію
evict-ять, лише коли ready-ассетів стає більше двох
(`backend/internal/media/adapters/postgres/publish.go:79-96`). Канонічне «поточне» визначення вже є:
`readyGenerationForLesson` (`backend/internal/transcript/adapters/postgres/repository.go:20-26`,
`ORDER BY generation DESC LIMIT 1`); ним рахується застарілість. Два селектори ASR його не
використовують:

1. `PendingRecognitions` (`adapters/postgres/draft.go`) вибирає **кожен** ready-ассет, чия
   генерація відрізняється від генерації чернетки. Із двома ready-ассетами це нескінченний цикл:
   gen1 ≠ чернетка gen2 → розпізнати gen1; тоді gen2 ≠ чернетка gen1 → розпізнати gen2; і знову.
   У живому стеку (старий образ без C09) один урок розпізнано 5 разів поспіль, speech і no_speech
   чергувались. З фіксом C09 стара генерація вже не перезаписує, але селектор усе одно ставить у
   чергу старий ассет.
2. `LessonAsset` (`adapters/postgres/draft.go`, його кличе `Regenerate`) — `QueryRow` без
   `ORDER BY`/`LIMIT`. За двох ready-ассетів Postgres повертає будь-який, тож ручна перегенерація
   може поставити в чергу **стару** генерацію. З C09 вона мовчки відсічеться, і UI висітиме в
   «черзі» до 15-хвилинної стелі; без C09 розпізналося б старе відео.

**Фікс:** обидва селектори беруть лише поточний ассет уроку — найновішу ready-генерацію без
`evicted_at` (те саме визначення, що `readyGenerationForLesson`). Для `PendingRecognitions` — по
одному ассету на урок (напр. `DISTINCT ON (lesson_id) … ORDER BY lesson_id, generation DESC` або
підзапит), далі та сама умова «чернетки поточної генерації немає». Для `LessonAsset` —
`ORDER BY generation DESC LIMIT 1`. Визначення «поточного» не дублюй третьою копією, якщо можна
перевикористати наявне.

**Integration-тести** (`draft_integration_test.go`, справжній Postgres): урок із двома ready-ассетами
без `evicted_at` (gen1, gen2):
- чернетка gen1 → `PendingRecognitions` повертає лише gen2;
- чернетка gen2 → не повертає нічого (циклу немає);
- `LessonAsset` повертає gen2.

## (B) E2e: Select у React Aria — це `button`, а не `combobox` (C18, C11)

`frontend/e2e/asr.spec.ts` `selectLanguage` шукає `getByRole('combobox', …)`, але тригер `Select` у
кіті — `Aria.Button` (`frontend/src/presentation/ui/controls.tsx:84`). Траса C18 показує, що селект
«Мова розпізнавання» видимий, а локатор його не знаходить. Unit-тест уже шукає правильно:
`getByRole('button', { name: /Recognition language/i })` (`LessonDraftField.test.tsx:206`). Заміни
роль на `button`.

## (C) E2e: C11 не ставить задачу сам (C11)

C11 — єдиний тест, що не кличе `startRecognition` перед першим `waitForDraft` (~:180). Він
покладається на автопостановку, а її при відкритій сторінці не видно (узгоджена межа 4c). Додай
`startRecognition`. Заодно зроби `startRecognition` стійким: зараз він клікає лише якщо кнопка вже
видима (`isVisible` без очікування) і тихо пропускає клік, коли поле ще не відрендерилось. Дочекайся,
поки з'явиться або кнопка «Згенерувати/Generate», або стан у процесі (`[data-draft-state]`
pending/running — його поставила автопостановка), і клікай лише в першому випадку. Не клікай
«Перегенерувати», якщо чернетка вже є. Стійкий `startRecognition` кличуть і вже зелені C13, C03, C06, C17 — переконайся за кодом
кожного, що нова версія їх не ламає (у C13 чернетки спершу немає: кнопка «Generate» з'явиться, і клік
має відбутися).

## NOT
- Не чіпай `media` (утримання двох ready-ассетів — свідомий дизайн відтворення), міграції,
  фронтенд `src/`, `openapi.yaml`. Git, `tasks.md`, `report.md` — ні.

## Перевірки
Бекенд: `gofmt`, `go vet ./internal/transcript/...`, `go test ./internal/transcript/...`,
`go test -tags integration -run '^$' ./internal/transcript/...`. Фронтенд: `npm run format:check`,
`npm run lint`, `npm run typecheck`, `npx playwright test e2e/asr.spec.ts --list`. Integration і
e2e без Docker не запускаєш — чесно познач. Е2e жене Courses 2 після перезбірки **api і
asr-worker** (зміни в бекенді) та recreate.

## Звіт
`work/report-a13-current-asset.md`
