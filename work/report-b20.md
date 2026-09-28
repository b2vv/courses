# B20 — журнал виконання

- 2026-09-28: Роботу розпочато. Перший запис зроблено до читання брифу та будь-яких змін у коді; коміт створювати не буду.

- 2026-09-28 — ФАКТ: B20 стосується лише транспортного контракту й фронтового адаптера: три ASR-draft endpoint-и вже існують у backend handler, але відсутні в OpenAPI; домен і backend-поведінку змінювати не можна.
- РІШЕННЯ: Спочатку звірю фактичні handler/application/domain error paths і DTO, потім додам OpenAPI, запущу генерацію, переведу `draft-repository.ts` на Orval із явним мапінгом, видалю ручну zod-схему лише якщо її повністю замінено, і виконаю дозволені локальні гейти без Docker/e2e.
- ДАЛІ: Прочитати повні реалізації `draft_handler.go`, application draft use cases, доменний draft, поточний фронтовий adapter/schema та генераційний конфіг; зафіксувати реальні response/error коди перед змінами.

- 2026-09-28 — ФАКТ: Реальний backend path дає `Get`: 200 (у т.ч. без draft), 400 `invalid_request`, 401 `authentication_required`, 403 `course_forbidden`, 404 `lesson_not_found`, 503 `dependency_unavailable`; `Regenerate`: 202, 400, 401, 403 `origin_forbidden`/`course_forbidden`, 404 `lesson_not_found`, 409 `already_running`, 503; `Publish`: 200, 400, 401, 403 `origin_forbidden`/`course_forbidden`, 404 `lesson_not_found`, 409 `transcript_modified`/`transcript_exists`, 503. `asset_not_found` backend не повертає.
- ФАКТ: DTO завжди має поля `lessonId,state,language,failure,stale,updatedAt,assetId,generation,segments`; absent draft кодує nullable-поля й порожні segments. Backend domain обмежує failure 200 runes, segments 5000, text 1000, speaker 100, language 2–35; state enum = `queued|running|ready|failed|no_speech`.
- РІШЕННЯ: OpenAPI відображатиме `queued`; фронт мапитиме generated `queued` у наявний domain/UI `pending`, щоб не змінювати доменну поведінку. Generated models будуть runtime-перевірені вузьким type guard у репозиторії, бо Orval генерує TypeScript types, а не runtime validation.
- ДАЛІ: Додати три path-и та схеми `LessonDraft`, `DraftSegment`, два optional request body і response `updatedAt`; додати відсутній у спільному ErrorResponse код `already_running`, потім запустити Orval.

- 2026-09-28 — ФАКТ: OpenAPI доповнено трьома operationId (`getLessonDraft`, `regenerateLessonDraft`, `publishLessonDraft`), DTO/request schemas і `already_running` у ErrorResponse enum. Успішні draft-поля описані як required nullable там, де backend серіалізує `null`; `queued` залишено authoritative backend state.
- ДАЛІ: Запустити `make generate`, перевірити diff generated-файлів і переглянути фактичні Orval signatures/models перед переписуванням адаптера.

- 2026-09-28 — ФАКТ: `make generate` заблоковано sandbox: sqlc-частина запускає Docker і отримала `permission denied while trying to connect to the Docker API`; через `make` до Orval не дійшло. Це не вважаю проходженням генерації.
- РІШЕННЯ: Запущу frontend Orval окремо (`cd frontend && npm run generate`), бо він не потребує Docker; sqlc-generated tree не чіпатиму.
- ДАЛІ: Перегенерувати Orval і перевірити, що OpenAPI схема прийнята.

- 2026-09-28 — ФАКТ: `npm run generate` пройшов; Orval створив draft models/client для всіх трьох endpoint-ів, включно з `LessonDraftState.queued`, optional regenerate/publish requests і `publishLessonDraft` client. Адаптер переведено на `getLessonDraft`/`regenerateLessonDraft`; `draft-schema.ts` видалено.
- РІШЕННЯ: Runtime guard у адаптері відкидає невідомі поля/стани, неправильні UUID, межі сегментів, `endMs < startMs`, невалідні мову/failure/confidence; мапінг `queued → pending` є єдиним перекладом transport→domain.
- ДАЛІ: Запустити frontend typecheck/lint і виправити лише помилки, що виникли від цієї міграції.

- 2026-09-28 — ФАКТ: Після одного typecheck-виправлення frontend `npm run typecheck` і `npm run lint` пройшли; architecture boundaries — 113 files. `git diff --check` також чистий.
- ДАЛІ: Запустити frontend unit suite; паралельно перевірити `gofmt -l`, `go vet ./...` і `go build ./...` з окремим GOCACHE, як вимагає бриф.

- 2026-09-28 — ФАКТ: `gofmt -l backend` порожній; `GOCACHE=/private/tmp/courses-go-cache go vet ./...` і `go build ./...` завершилися успішно. `npm run test` не стартував через sandbox `listen EPERM` на `::1:3000`; це blocked, не passed.
- ФАКТ: `npm run format:check`, `npm run typecheck`, `npm run lint` пройшли після форматування adapter-а; повторний `npm run generate` пройшов. `git diff --check` чистий, посилань на видалений `draft-schema` або allow-list `asset_not_found` у draft adapter немає.
- РІШЕННЯ: Додав `DraftConflict` response component для точного 409 regenerate і використав `AdminMutationForbidden` для same-origin draft mutations; випадкову зміну старого logout response під час патчу виправлено й перевірено diff.
- ДАЛІ: Фінальна read-only перевірка diff/status, зафіксувати список змін і обмеження. Коміт не створювати.

- 2026-09-28 — ФАКТ: Фінальна перевірка підтвердила три generated API operations і generated models (`LessonDraft`, `DraftSegment`, requests, `DraftConflictResponse`), чистий `git diff --check`, HEAD без нового коміту (`ba5279a`). Незв’язані зміни PNG/AGENDA/CTO-RESEARCH збережено й не чіпано.
- РІШЕННЯ: Роботу залишаю в робочому дереві без коміту, як вимагав бриф; Docker/sqlc і frontend unit suite залишаються явно позначеними blocked у цьому журналі.
- ДАЛІ: B20 виконано; повний `make check` запускає Courses поза цією sandbox-сесією.

- 2026-09-28 — ФАКТ: Review diff завершено: зміни обмежені OpenAPI/generated client, draft adapter і видаленням дубльованої schema; backend/domain/transport не змінені. Виявлену неточність опису `DraftConflict` виправлено, Orval і typecheck повторно пройшли.
- ДАЛІ: Немає незавершених змін у межах B20; очікується зовнішній Docker/e2e/full-check прогін Courses.
