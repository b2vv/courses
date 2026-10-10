# Brief P05 — «немає доступу» на маршруті навчання як названий стан

Phase 9. Прочитай спершу: `docs/readiness/spec.md` (§3 D7; §6 R11), `docs/readiness/plan.md` §2
(підозра P05), `frontend/src/routes/learn.tsx`, `frontend/src/routes/learn.$.tsx`,
`frontend/src/presentation/identity/ProtectedRouteStates.tsx`, локалі `en.ts`/`uk.ts`.

## Підозра, яку спершу треба перевірити, а не вгадати

`learn.$.tsx` у `beforeLoad` кличе `ensureRouteQuery(context.learn.course(slug))`. Будь-яка помилка
запиту курсу йде в `errorComponent` батьківського маршруту (`learn.tsx:19` —
`ProtectedRouteFailure`), а він показує `login.failure` («не вдалося перевірити сесію») з
«спробувати знову». Якщо бекенд на чернетку чужого автора відповідає `course_forbidden` (або іншим
кодом відмови доступу), студент бачить збій входу, а не «немає доступу».

1. З'ясуй за кодом: що саме повертає бекенд на читання курсу для навчання стороннім на чужій чернетці
   (`backend/internal/course/adapters/http/…`), як це мапить фронтенд-адаптер і доменний тип
   помилки (`frontend/src/domain/course/course.ts:49-50` — `'forbidden' | 'course_forbidden'`), і
   куди потрапляє помилка на маршруті. Опиши в звіті ланцюжок із file:line.
2. Якщо підозра підтверджується — **виправ**: відмова доступу до курсу рендериться названим станом
   «немає доступу» з людським поясненням і посиланням до каталогу, **без** «спробувати знову» (повтор
   нічого не змінить). Справжній збій перевірки сесії лишається `ProtectedRouteFailure`. Фронтенд
   нічого не вирішує сам — лише відображає рішення бекенду.
3. Якщо підозра НЕ підтверджується (стан уже правильний) — не змінюй продукт, лише додай тести.

## Тести (R11)

- Unit: відмова `course_forbidden` на маршруті навчання → стан «немає доступу»; збій сесії → як
  раніше `ProtectedRouteFailure`.
- E2e в `frontend/e2e/`: за прецедентом `asr.spec.ts` засій через API адміна курс-**чернетку** з
  уроком і неймспейсом на прогін; зі **збереженою сесією студента** (`authStatePath('student')`,
  без нових логінів — бюджет B19 на стелі) відкрий `/learn/<slug>` → видно «немає доступу», немає
  «спробувати знову». Прибери курс у `afterAll`. Статуси й повідомлення звужуй текстом, не
  `getByRole('status')` без фільтра (AGENTS.md §Decisions).

## NOT

Бекенд, контракт, `openapi.yaml`, `docs/`, `work/tasks.md`, git — ні. Нові тексти — в `en.ts` і
`uk.ts`.

## Перевірки

У `frontend`: `npm run format:check`, `npm run lint`, `npm run typecheck`,
`npx playwright test <новий файл або тест> --list`. `npm test` у пісочниці не стартує — чесно.

## Звіт

`work/report-p05.md`
