## 2026-09-28 — старт

- Почато виконання `work/brief-d04.md`.
- Коміти не створюю.
- Наступне: прочитати репозиторний брифінг, інструкції реалізації та brief D04.

## 2026-09-28 — мінори D03

- ФАКТ: `POST /api/v1/learn/lessons/{lessonId}/comments` тепер посилається на `PlaybackForbidden`, опис якого містить `Origin`; Orval успішно перегенеровано, змінився лише відповідний generated response type.
- ФАКТ: `application.Service.Update` тепер перевіряє право автора до стану `removed`; додано contract-кейс «чужий на прибраному» → `403 comment_forbidden`, збережено авторський `409 comment_removed`.
- ПЕРЕВІРКА: `gofmt` виконано; `go test ./internal/discussion/...` заблоковано правами на глобальний Go build cache (`~/Library/Caches/go-build/... operation not permitted`). `npm run generate` з `frontend/` пройшов.
- ДАЛІ: додати frontend domain/port та API adapter із runtime-перевіркою wire-відповіді.

## 2026-09-28 — frontend seam і перша інтеграція

- ФАКТ: додано `domain/discussion`, `DiscussionRepository`, Orval-based adapter із явним мапінгом і runtime-перевіркою сторінки/коментаря; generated transport не виходить за adapter.
- ФАКТ: додано Query key, query та publish/update/remove mutations, усі scoped lesson-ом і з інвалідацією після власної мутації.
- ФАКТ: `DiscussionPanel` підключено до lesson route без suspense; панель підтримує корені/відповіді, мітки → `onSeek`, поточний момент, серверні `canEdit`/`canRemove`, курсор «показати ще», loading/empty/error/retry та локалізовані коди помилок.
- ДАЛІ: додати/уточнити unit-тести адаптера, Query і панелі, стилі та пройти frontend format/type/lint; після цього виправити знайдені compile/test проблеми.

## 2026-09-28 — тести й гейти

- ФАКТ: додано adapter-тести на явний мапінг, відмову невалідної wire-відповіді та named `comment_thread_full`; додано panel-тести на timestamp → `onSeek`, «поточний момент» і відсутність edit/remove без серверних capability.
- ФАКТ: порожній backend `nextCursor` мапиться в доменний `null`, тому «показати ще» не з’являється без наступної сторінки.
- ПРОЙШЛО: з `frontend/` `npm run format:check`, `npm run typecheck`, `npm run lint`; `lint` завершився з `Architecture boundaries checked (118 files)`. З `backend/` через локальний дозволений cache `GOCACHE=/private/tmp/courses-go-cache go test ./internal/discussion/...` і `go vet ./...` пройшли.
- BLOCKED: Rstest не стартує в sandbox — `listen EPERM ::1:3000`; повний `npm test`, e2e та `make check` не подаю як пройдені.
- ДАЛІ: фінальний diff-аудит, перевірити OpenAPI/generated drift, зафіксувати обмеження й статус у звіті без коміту.

## 2026-09-28 — завершення D04

- ФАКТ: фінальний diff-аудит пройшов `git diff --check`; зміни OpenAPI й generated узгоджені (`npm run generate` без додаткового drift). Незакомічені PNG у `docs/foundation/screenshots/` були наявні до задачі й не змінювалися мною.
- ФАКТ: code-review за стандартами й специфікацією не виявив додаткових блокувальних розбіжностей; D05/e2e свідомо не виконувався, бо це наступний слайс.
- ПРОЙШЛО: frontend `format:check`, `typecheck`, `lint`; backend `go test ./internal/discussion/...`, `go vet ./...`; генерація Orval.
- BLOCKED: frontend Rstest, e2e і повний `make check` залишаються неперевіреними через sandbox `listen EPERM ::1:3000`; це не позначаю як passed.
- РІШЕННЯ: коміт не створювався, як вимагав запит.

## 2026-09-28 — follow-up K16/K18

- ДАЛІ: дописати лише frontend-тести для K16 у `DiscussionPanel.test.tsx` і K18 у наявному learn-route тесті; продуктову поведінку не змінюю.

## 2026-09-28 — K16/K18 тести

- ФАКТ: `DiscussionPanel.test.tsx` тепер окремо доводить K16: loading status у межах panel, empty, error + retry callback, reply з `parentId` кореня без `positionMs`, removed comment без тексту/дій, deleted author, `editedAt`, cursor-only «показати ще» та людський текст `comment_thread_full`.
- ФАКТ: K18 додано в `route-guards.test.tsx`: тестова фабрика маршруту приймає transcript/discussion repositories і media factory; discussion repository повертає `dependency_unavailable`, transcript успішний, після монтування перевіряються player, transcript і panel error без `LearnRouteFailure`.
- РІШЕННЯ: зміни `route-guards.test.tsx` потрібні лише для dependency injection у тестовий harness; production-код і продуктова поведінка не змінювалися.
- ПРОЙШЛО: з `frontend/` `npm run format:check`, `npm run typecheck`, `npm run lint` (`Architecture boundaries checked (118 files)`).
- BLOCKED: rstest у sandbox не стартує через `listen EPERM ::1:3000`; 5 D04-тестів підтверджені зовнішнім прогоном Courses, локально не дублюю як passed.
- ДАЛІ: немає; коміт не створюю.

## 2026-09-28 — виправлення дефектів A–D

- ФАКТ: ручний cursor/page accumulator у `learn.$.tsx` замінено на TanStack `useInfiniteQuery` з lesson-scoped key, `initialPageParam`, `getNextPageParam` і `fetchNextPage`; прибрано effects та всі два route-level `oxlint-disable`. Інвалідація мутацій тепер перезапитує завантажені сторінки узгоджено.
- ФАКТ: `DiscussionQueryError.code` використовується через `instanceof`; приведення неперевірених `error.message` до `DiscussionErrorCode` прибрано.
- ФАКТ: reply form тепер справжній `<form>` із submit, має окрему доступну назву `Send reply` / `Надіслати відповідь`; головна форма лишила `Post comment` / `Опублікувати`. K16-тест оновлено.
- ФАКТ: show-more тест демонтує першу панель перед другим render.
- ФАКТ: додано regression test у `discussion-queries.test.tsx`: після завантаження двох сторінок і власної публікації перевіряються новий корінь першим і відсутність duplicate IDs.
- ПРОЙШЛО: `npm run format:check`, `npm run typecheck`, `npm run lint`; lint без нових `oxlint-disable`, architecture `118 files`.
- ДАЛІ: rstest проганяє Courses поза sandbox; коміт не створюю.

## 2026-09-28 — follow-up review завершено

- ФАКТ: `CommentCard` отримав явний `canReply`; він `true` лише для кореня, а для replies — `false`, тому відповідь більше не може послати власний `id` як `parentId`.
- ФАКТ: reply regression test перевіряє відсутність `Reply` у картці відповіді та наявність кнопки на корені; наявний тест публікації з `parentId` кореня збережено.
- ФАКТ: прибрано недосяжний `comment_not_found` із `publishErrors`.
- ПРОЙШЛО: `npm run format:check`, `npm run typecheck`, `npm run lint`; architecture `118 files`.
- РІШЕННЯ: коміт не створюю.

## 2026-09-28 — follow-up review D04

- ДАЛІ: обмежити reply UI кореневими коментарями (`canReply`), додати тест на відсутність Reply у картці відповіді, прибрати недосяжний publish error `comment_not_found`, потім пройти frontend gates.

## 2026-09-28 — follow-up: дефекти після розширеного прогону

- ДАЛІ: замінити ручний накопичувач discussion pages у learn route на `useInfiniteQuery`, прибрати його effects/disable-коментарі, типізувати query error, виправити reply form/label і стабілізувати show-more тест.
