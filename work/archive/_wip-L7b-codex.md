ДАЛІ: Прочитати L7b-план і наявні enrollment/teach шари, потім звірити контракти перед змінами.

2026-10-07 — Крок 1: прочитано `work/brief-codex-L7b.md`, `work/codex-working-model.md` і репозиторний briefing. Межі зафіксовані: лише frontend enrollment roster/block/unblock; без git, `work/tasks.md`, phase report та ручних змін generated.

ДАЛІ: Додати типи, порт, API-адаптер і query/mutation options для roster; після цього — компоненти й route.

2026-10-07 — Крок 2: звірено `docs/enrollment/plan.md` (En06, En08, En12) з L7a-шаром і teach moderation. Контракти існують: generated roster/keyset, block/unblock, root route-context `enrollment`. Виявлено, що generated `displayName` має minLength 1, але UI все одно безпечно анонімізує порожній рядок за вимогою brief.

## Discussion

**Зупинено: контракт 404 для block/unblock відсутній, хоча brief вимагає його мапити.**

- Evidence: `work/brief-codex-L7b.md:24-25` вимагає внесення 404 до enrollment error-union. Але generated API для `blockCourseParticipant` декларативно має тільки 400/401/403/409/500/503 (`frontend/src/adapters/api/generated/courses.ts:637-679`), а для `unblockCourseParticipant` — той самий перелік (`:714-756`). Поточний `EnrollmentErrorCode` не містить жодного 404-коду (`frontend/src/domain/enrollment/enrollment.ts:17-30`).
- Runtime consequence: якщо бекенд надає заявлений brief 404 (наприклад, participant/course no longer exists), клієнт або мусить тихо перетворити його на `invalid_response`, або додати вигаданий/незгенерований контрактний код. Перше маскує бізнес-відмову; друге розходиться з OpenAPI та generated client. Обидва варіанти є багом.
- Proposed correction: reviewing session має звірити backend/OpenAPI і, якщо 404 є intended, додати його з іменованим error code до OpenAPI та регенерувати клієнт; потім явно назвати відповідний код у brief (наприклад, `course_not_found` або `membership_not_found`) і дозволити L7b додати його до error-union. Якщо 404 не є контрактом — прибрати `/404` з L7b brief.
- What I did: зупинився до frontend-змін. Змінено тільки обов'язковий append-only WIP; git, `work/tasks.md`, phase report і `generated/` не чіпались.

ДАЛІ: Очікую рішення reviewing session щодо 404-контракту; після виправленого brief/контракту продовжити L7b з data layer.

ДАЛІ: Реалізувати та перевірити instructor roster UI (Tabs, CourseRoster, i18n), після чого додати/оновити тести.

2026-10-07 — Крок 3: reviewer виправив brief: 404 для block/unblock не є контрактом; стару дискусію закрито цим рішенням. Додано enrollment domain roster types, repository methods, строго валідований API adapter (keyset + block/unblock лише 400/401/403/409/5xx сітка) та infinite-query/mutations з інвалідацією виключно `['enrollment','roster',courseId]`. Generated/OpenAPI не змінювалися.

ДАЛІ: Усунути результати typecheck, додати цільові unit/component тести для roster, потім запустити дозволені перевірки.

2026-10-07 — Крок 4: додано instructor roster як окрему секцію на teach course screen (дозволена brief-альтернатива Tabs): loader prefetch, `useInfiniteQuery`, block/unblock mutations через route-context; presentation не виконує fetch. `CourseRoster` має empty/error+retry, keyset «Показати ще», текстові state-бейджі, locale date, ProgressBar і disable під час mutation. Додано uk/en ключі. Existing test stubs розширено під порт.

ДАЛІ: Запустити typecheck і lint після нових roster tests; потім Rstest, якщо sandbox дозволить порт.

2026-10-07 — Крок 5: додано unit/component tests для 0% progress, empty state, empty-name anonymization, block/unblock action, error retry/keyset paging та scoped roster invalidation. Наявний query-test і route-guard stubs оновлено під розширений порт.

ДАЛІ: Запустити `npm test` (Rstest); якщо sandbox відмовить порт — зафіксувати not-run denial.

2026-10-07 — Крок 6: `cd frontend && npm run typecheck` — GREEN. `npm run lint` — GREEN, включно з architecture boundaries (171 files).

ДАЛІ: Виправити один некоректний test expectation і перезапустити Rstest.

2026-10-07 — Крок 7: `npm test` ЗАПУЩЕНО, але FAIL: 624/625. Єдиний fail — новий `CourseRoster` test очікував «Колишній автор», тоді як наявний переклад `ratings.formerAuthor` має «Колишній учасник». Це помилка тестового очікування, не product-path; sandbox port не блокував запуск.

ДАЛІ: Повторити typecheck/lint після error-rendering hardening; build/e2e лишити reviewing session за brief.

2026-10-07 — Крок 8: після корекції test expectation `cd frontend && npm test` — GREEN, Rstest 625/625. Додатково нормалізовано невідомі transport/query error codes до наявного `enrollment.errors.generic`, щоб UI не показував сирий i18n key.

ДАЛІ: Reviewing session має прогнати Rstest поза sandbox, потім build/e2e (L8) і повний gate; Codex work завершено.

2026-10-07 — Крок 9: фінальний `npm run typecheck` — GREEN; фінальний `npm run lint` — GREEN (architecture boundaries 171 files). Фінальний `npm test` **NOT RUN / DENIAL**: sandbox відмовив bind `::1:3000` з `listen EPERM`. Не зараховано як passed; reviewing session має повторити. Build/e2e не запускались за brief (e2e належить L8).
