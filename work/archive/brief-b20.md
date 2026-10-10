# Бриф B-20 — draft-ендпоінти в OpenAPI, фронт на Orval

РОБОЧИЙ ФАЙЛ: work/report-b20.md
Дописуй ПІСЛЯ КОЖНОГО КРОКУ (append, не rewrite). Перший запис — до першого інструменту:
рядок `ДАЛІ:` з тим, що збираєшся перевіряти. Далі — ФАКТ / РІШЕННЯ / ДАЛІ / ПИТАННЯ.
Фінальна відповідь = стислий переказ файлу.

## Задача

Закрити борг B-20 (`work/backlog.md` §B-20): три draft-ендпоінти ASR зареєстровані в
`backend/internal/transcript/adapters/http/draft_handler.go`, але **відсутні в
`backend/openapi/openapi.yaml`**. Через це фронт ходить рукописним адаптером із власними
zod-схемами (`frontend/src/adapters/api/draft-repository.ts`, `draft-schema.ts`) — контракт
живе у двох НЕзвірених копіях. Треба описати ендпоінти в OpenAPI, перегенерувати Orval, і
перевести фронт-адаптер на згенероване з явним мапінгом у домен, щоб `check-generated` ловив дрейф.

## Три ендпоінти (авторитет — бекенд, НЕ поточний фронт)

1. `GET  /api/v1/admin/lessons/{lessonId}/draft` → 200 `draftDTO` (форма — у `draft_handler.go`, тип `draftDTO`/`draftSegmentDTO`). `Cache-Control: no-store`.
2. `POST /api/v1/admin/lessons/{lessonId}/draft/regenerate` → 202, тіло `{ "language"?: string }`. Проходить `sameOrigin`.
3. `POST /api/v1/admin/lessons/{lessonId}/draft/publish` → 200 `{ "updatedAt": string }`, тіло `{ "expectedUpdatedAt"?: string }`. Проходить `sameOrigin`.

## Коди помилок — вивести АВТОРИТЕТНО з бекенду, не з фронту

Джерело істини:
- `authenticate` middleware → 401 `authentication_required` (усі три).
- `sameOrigin` middleware → 403 `origin_forbidden` (regenerate, publish; НЕ get).
- `writeTranscriptError` (`handler.go:119`) — мапінг доменних помилок: `ErrForbidden`→403 `course_forbidden`, `ErrNotFound`→404 `lesson_not_found`, `ErrTranscriptModified`→409 `transcript_modified`, `ErrTranscriptExists`→409 `transcript_exists`, `ErrAlreadyRunning`→409 `already_running`, `ErrDependencyUnavailable`→503 `dependency_unavailable`, `ErrInvalidRequest`/default→400 `invalid_request` (+ segment/text/document limits — лише для write-транскрипту, навряд для draft).
- Для КОЖНОГО ендпоінта пройди його реальний шлях: handler → `application.Drafting.{GetDraft,Regenerate,Publish}` → які доменні помилки повертаються. Опиши в OpenAPI лише ті коди/статуси, що ендпоінт РЕАЛЬНО може віддати.

**⚠️ Підозра на дрейф (звір і назви у звіті):** поточний рукописний фронт `draft-repository.ts`
дозволяє для `regenerate` 404 `asset_not_found`, а `writeTranscriptError` мапить `ErrNotFound`
лише в `lesson_not_found`. З'ясуй, чи `Regenerate` реально віддає `asset_not_found` (окремим
шляхом), чи це мертвий код у фронт-allow-list (клас C10). Так само звір усі інші коди
frontend-allow-list проти бекенду. Якщо знайдеш розбіжність — це і є цінність B-20; **зафіксуй
у звіті, але не «виправляй» бекенд без окремого рішення** (борг про контракт, не про поведінку).

## Що зробити

1. **OpenAPI** (`backend/openapi/openapi.yaml`): додай три path-описи в стилі наявних (див.
   `/api/v1/admin/lessons/{lessonId}/transcript` і `/api/v1/admin/courses/{id}` — reuse
   компонента `ErrorResponse`, `Origin`-хедер де є `sameOrigin`, patterns/enums). Додай у
   `components/schemas` схему чернетки (`LessonDraft` + segment) з тими ж обмеженнями, що в
   доменних DTO (state enum `pending|running|ready|failed|no_speech`, segments max 5000,
   text max 1000, confidence 0..1, тощо — звір із `draft-schema.ts` і DTO).
2. **Регенерація:** `make generate` (sqlc + `cd frontend && npm run generate`). `check-generated`
   має бути зеленим (`git diff --exit-code` по generated-теках + жодних untracked). Якщо
   пісочниця блокує щось — познач у звіті, я доганю гейт.
3. **Фронт-адаптер:** перепиши `draft-repository.ts` на згенерований Orval-клієнт
   (`frontend/src/adapters/api/generated/`), із ЯВНИМ мапінгом у доменний `LessonDraft`
   (`domain/transcript/draft`). Збережи порт `DraftRepository` (getDraft, regenerate) незмінним і
   збережи named-error семантику (ті самі доменні `DraftErrorCode`). Прибери `draft-schema.ts`,
   якщо Orval-моделі його заміняють (мапінг лишається у репозиторії). `publish` фронтом НЕ
   споживається (порт його не має) — його клієнт Orval згенерує, але адаптер під нього НЕ пиши.
3a. Не чіпай `transport.ts`/`api-origin.ts` без потреби; не змінюй домен і поведінку бекенду.
4. **Гейти в пісочниці:** `gofmt -l`, `go vet ./...` (з `GOCACHE=/private/tmp/courses-go-cache`),
   `go build ./...`; `cd frontend && npm run typecheck && npm run lint`. НЕ ганяй docker/e2e —
   це на мені (Courses). Познач у звіті, що з гейтів пройшло, що заблоковано пісочницею.

## Межі / застереження

- Це контракт транспорту, не поведінка. Бекенд-логіку й домен не міняй.
- Не комітити. Диф лишаєш у дереві; я ганяю повний `make check` і віддаю на рев'ю Courses 2.
- AGENTS.md §Architecture: generated-моделі лишаються в адаптерах, мапінг у домен явний.
- Якщо натрапиш на відсутній контракт або двозначність — стоп, ПИТАННЯ у звіт, не вгадуй.
