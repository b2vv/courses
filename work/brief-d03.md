# Бриф D03 — application, HTTP і контракт `discussion`

РОБОЧИЙ ФАЙЛ: work/report-d03.md
Дописуй ПІСЛЯ КОЖНОГО КРОКУ (append, не rewrite). Перший запис — до першого інструменту:
рядок `ДАЛІ:` з тим, що збираєшся перевіряти. Далі — ФАКТ / РІШЕННЯ / ДАЛІ / ПИТАННЯ.
Фінальна відповідь = стислий переказ файлу.

## Контекст

Phase 7b. Спека `docs/discussion/spec.md` (D3, D4, D5; K10–K15), план `docs/discussion/plan.md`
(§1 D03, **§2 — обидва пункти D03**). Домен (D01) і порти/адаптер (D02) уже в дереві:
`backend/internal/discussion/{domain,ports,adapters/postgres}`. Прочитай усе це перед кодом.
Цей слайс — application, HTTP-адаптер, OpenAPI, Orval-регенерація, підключення в `cmd/api`.
**Без** фронтенд-адаптера й UI (D04).

## Що зробити

1. **`backend/internal/discussion/application`** — чотири use case: список, опублікувати, змінити,
   прибрати. Шаблон — `transcript/application` (`read.go`, `write.go`).
   - Читання й публікація: факти уроку з порту → `domain.Course.CanBeViewedBy`; анонім/без права —
     доменні помилки, які HTTP мапить у коди спеки D3.
   - **Get-then-authorize (план §2, D03):** адаптер D02 схлопує причини — `UpdateBody` фільтрує
     `id AND author_id AND removed_at IS NULL`, `Remove` безумовний за `id`. Тож для зміни й прибирання
     application **спершу читає коментар із фактами курсу**, розрізняє: невідомий → `comment_not_found`
     (404); прибраний при зміні → `comment_removed` (409); не має права → `comment_forbidden` (403); і лише
     тоді викликає `UpdateBody`/`Remove`. Повторне прибирання вже прибраного власником права — 204.
   - Можливості глядача `canEdit`/`canRemove` — обчислює сервер через `CanBeEditedBy`/`CanBeRemovedBy`
     (спека D4), фронт їх не повторює.
2. **HTTP-адаптер `backend/internal/discussion/adapters/http`** — шляхи спеки D5:
   `GET/POST /api/v1/learn/lessons/{lessonId}/comments`, `PATCH/DELETE /api/v1/learn/comments/{commentId}`.
   `Register(mux, authenticate, sameOrigin)` як у `playback` (`cmd/api/main.go:109`): `GET` — лише
   `authenticate`; мутації — `authenticate` + `sameOrigin`. `Cache-Control: no-store`. Тіло запиту —
   `DisallowUnknownFields`, ліміт розміру, один JSON-об'єкт (як `draft_handler.go`). Коди помилок — рівно
   таблиця спеки D3 (включно з `comment_cursor_invalid`, `comment_thread_full`); для читання чужої
   чернетки — 403 `course_forbidden`, як у транскрипті (ADR 010). Автор — лише `displayName` (null для
   видаленого), без email і без id автора.
3. **OpenAPI** (`backend/openapi/openapi.yaml`) — чотири операції в стилі наявних (`ErrorResponse`,
   `Origin` для мутацій, `NoStore`, схеми). Нові коди — в enum `ErrorResponse`. **Урок B-20:** кожен код, який
   віддає хендлер, є в OpenAPI, і кожен код в OpenAPI для цих шляхів має реальний шлях у хендлері —
   звір обидва напрямки й запиши у звіт таблицею.
4. **Orval:** `cd frontend && npm run generate`. Згенерований клієнт лишається невикористаним до D04 — це
   нормально. Нові generated-файли не забудь (їх треба буде застейджити для `check-generated`).
5. **Підключення:** `cmd/api/main.go` — репозиторій, application, handler, `Register` поряд із
   `transcriptHandler`.
6. **Contract-тести** (`adapters/http/handler_test.go`, фейковий порт як у `transcript` — там само є
   `assertContract`, що валідує запит **і** відповідь проти `openapi.yaml`): K10–K14 — кожен статус і код
   з таблиці D3, `canEdit`/`canRemove` для автора / іншого студента / адміна / власника курсу, форма
   прибраного (без тіла), видалений автор (`displayName: null`), повторне прибирання — 204, `DELETE`
   невідомого — 404 (саме те, що схлопнув би адаптер без Get-then-authorize). Кожна відповідь проходить
   `assertContract`.

## Гейти в пісочниці

`gofmt -l backend`; з `backend`: `GOCACHE=/private/tmp/courses-go-cache go vet ./...`, `go build ./...`,
`go test ./internal/discussion/... ./platform/architecture/...`; `cd frontend && npm run typecheck && npm run
lint`. `npm test`, integration, e2e і повний `make check` — на мені (у пісочниці не стартують): позначай blocked.

## Межі

Не комітити. Домен/порти/адаптер D02 змінювати лише якщо без цього не виходить — тоді аргумент у звіт.
Фронтенд-код, крім згенерованого Orval, — D04. Двозначність — стоп і ПИТАННЯ у звіт.
