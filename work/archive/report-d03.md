## 2026-09-28 — старт

- Почав виконання `work/brief-d03.md`.
- Коміти не створюю.

## 2026-09-28 — контекст і рішення

- ФАКТ: D01–D02 уже містять домен, порти та PostgreSQL-адаптер. Адаптер навмисно схлопує причини для `UpdateBody`, а `Remove` ідемпотентний, тому D03 мусить робити get-then-authorize перед мутаціями.
- РІШЕННЯ: HTTP буде тонким шаром над чотирма use case; усі чотири маршрути матимуть `Cache-Control: no-store`, GET — лише `authenticate`, мутації — `sameOrigin(authenticate(...))`.
- ДАЛІ: реалізувати application-контракт і покрити його unit-тестами перед HTTP/OpenAPI.

## 2026-09-28 — application / HTTP / контракт

- ФАКТ: додано `backend/internal/discussion/application`: список, публікація кореня/відповіді, зміна і прибирання; mutation flow читає запис до перевірки прав, тому розрізняє `comment_not_found`, `comment_forbidden`, `comment_removed`, а повторне прибирання дозволеного запису лишається 204.
- ФАКТ: додано HTTP routes `GET/POST /api/v1/learn/lessons/{lessonId}/comments` і `PATCH/DELETE /api/v1/learn/comments/{commentId}`. GET обгорнутий лише authentication; мутації — same-origin + authentication; JSON має ліміт, `DisallowUnknownFields`, duplicate-key/trailing-content перевірки; усі відповіді no-store.
- ФАКТ: додано OpenAPI schemas/operations та error enum; `npm run generate` пройшов, generated Orval client/models оновлені. Discussion handler підключений у `cmd/api/main.go`.
- ФАКТ: `GOCACHE=/private/tmp/courses-go-cache go test ./internal/discussion/... ./cmd/api` і HTTP contract tests — green.
- ДАЛІ: прогнати `gofmt`, vet, build, frontend typecheck/lint; потім звірити коди хендлера з OpenAPI й зафіксувати межі гейтів.

## 2026-09-28 — перевірка

- ФАКТ: `gofmt -l backend`, `go vet ./...`, `go build ./...`, `go test ./internal/discussion/... ./platform/architecture/...`, `go test ./...` (крім sandbox-only lifecycle bind), `npm run typecheck`, `npm run lint`, `npm run generate`, `git diff --check` — виконано; D03-тести зелені.
- БЛОКЕР: повний Go `./...` має один не-D03 збій `platform/httpserver/TestShutdownDrainsAnInFlightResponse`: sandbox забороняє `listen tcp 127.0.0.1:0` (`operation not permitted`). Integration, `npm test`, e2e та повний `make check` не запускаю: бриф прямо залишає їх для Courses поза sandbox.
- ЗВІРКА B-20: хендлер реально віддає `authentication_required`, `course_forbidden`, `lesson_not_found`, `comment_cursor_invalid`, `comment_body_invalid`, `comment_position_invalid`, `comment_parent_invalid`, `comment_thread_full`, `comment_forbidden`, `comment_not_found`, `comment_removed`, `dependency_unavailable`, `invalid_request`; усі є в `ErrorResponse.enum` і мають статусний шлях у чотирьох операціях. Origin middleware реально віддає `origin_forbidden`; він також є в enum.
- ДАЛІ: провести code review diff, виправити лише підтверджені дефекти, перевірити робоче дерево й завершити append-звіт без коміту.

## 2026-09-28 — завершення

- РЕВ'Ю: standards — порушень, що ловляться репозиторними гейтами, не знайдено; spec — D03 виконано. Виправлено capability edge case: `canRemove=false` для вже прибраного коментаря, хоча дозволений повторний DELETE як і раніше повертає 204.
- ФАКТ: фінальний `go test ./internal/discussion/... ./platform/architecture/...`, `npm run typecheck`, `npm run lint`, `gofmt` і `git diff --check` — green.
- ФАКТ: Orval generated diff містить лише нові D03 operation/model exports; UI/adapters D04 не додавались.
- СТАН: D03 реалізовано; комітів не створено. Інтеграційний/e2e/full `make check` лишаються зовнішнім гейтом Courses; повний Go `./...` має лише sandbox bind blocker, описаний вище.

## 2026-09-28 — ревізія Courses: план виправлення

- РІШЕННЯ: перенести `assertContract` у `serve(...)`, щоб кожна відповідь тестів D03 автоматично проходила OpenAPI response validation із фактичним method/path/body; для навмисно невалідних request body/query передавати `validRequest=false`.
- ДАЛІ: додати повні табличні K10–K14, окремо перевірити статус і `error.code`, потім прогнати `gofmt` та `go test ./internal/discussion/...`.

## 2026-09-28 — contract coverage K10–K14

- ФАКТ: `serve(...)` тепер відновлює consumed request body і сам викликає `assertContract` для кожної відповіді. `validRequest=false` використано для malformed JSON; response validation виконується завжди.
- K10 → `TestK10GetComments`: 200; 401 `authentication_required`; 403 draft `course_forbidden`; 404 `lesson_not_found`; 400 `comment_cursor_invalid`.
- K11 → `TestK11PostComments`: 201 root; 201 reply; 400 `comment_body_invalid`; 400 `comment_position_invalid`; 400 `comment_parent_invalid`; 409 `comment_thread_full`; 403 `course_forbidden`; додатково 400 `invalid_request` із `validRequest=false`.
- K12 → `TestK12PatchComments`: 200 author із ненульовим `editedAt`; 403 `comment_forbidden`; 404 `comment_not_found`; 409 `comment_removed`.
- K13 → `TestK13DeleteComments`: 204 author; 204 admin; 204 course owner; 403 `comment_forbidden`; 204 repeat; 404 `comment_not_found`.
- K14 → `TestK14CapabilitiesAndRemovedAuthorForm`: capabilities для author / other student / admin / course owner; removed root з reply має `body:null`, `removed:true`; deleted author має `author.displayName:null`.
- ОБМЕЖЕННЯ: `origin_forbidden` не тестується в цьому handler harness, бо `passThrough` навмисний; його продукує identity `sameOrigin` middleware і він лишається в OpenAPI/error enum.
- ФАКТ: `gofmt -w internal/discussion/adapters/http/handler_test.go` і `GOCACHE=/private/tmp/courses-go-cache go test ./internal/discussion/...` — green.
