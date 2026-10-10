# Report A14

## C08 — lesson deletion during recognition

Added `C08 lesson deletion removes the draft before completion` to the existing PostgreSQL
integration harness in `backend/internal/transcript/adapters/postgres/draft_integration_test.go`.
The scenario inserts a ready asset, claims a running draft, deletes the lesson, verifies the
cascaded `asr_drafts` row is gone, and verifies completion with the claim token returns
`domain.ErrNotFound`. The existing unit `TestProcessNotFoundOnComplete` supplies the worker-side
proof that this result is a silent success without retry; the new test supplies the real-Postgres
side.

## C16 — red `/readyz` without model weights

Added `backend/cmd/asr-worker/main_test.go`. It uses the production `httpserver.NewHandler` and
the real `modelReady` callback, then checks the same `/readyz` route for a missing `model.bin`, a
directory named `model.bin`, and a regular model file: 503, 503, and 200 respectively.

## C19 — author-facing publication notice

The course editor now reports per-lesson transcript knowledge from the existing lazy transcript
queries. While publishing a draft course with at least one saved lesson whose transcript is
missing or not yet known, it renders a non-blocking semantic status output with English and
Ukrainian copy explaining that transcripts appear after speech in lesson videos is recognized.
The publish button remains enabled. When every saved lesson has a transcript, the notice is absent.
Added focused `CourseEditorView` unit coverage for both cases and extended P20 to assert the
notice before publication; existing publication flow remains unchanged.

## Verification

- `gofmt` — run on changed Go files.
- `go vet ./...` — passed.
- `GOCACHE=/tmp/courses-go-cache go test ./cmd/asr-worker/... ./internal/transcript/...` — passed.
- `GOCACHE=/tmp/courses-go-cache go test -tags integration -run '^$' ./internal/transcript/...` — passed as compile-only selection; integration cases were not executed because Docker is unavailable in this sandbox.
- `npm run format:check` — passed.
- `npm run lint` — passed.
- `npm run typecheck` — passed.
- `npx playwright test e2e/admin.spec.ts --list --workers=1` — passed; 8 tests listed.
- Rstest UI execution was not run: the sandbox denied its listener with `listen EPERM ::1:3000`.
- Full integration, e2e execution, and `npm test` were not run per the working model's sandbox restriction.

No git commands or commits were made. `work/tasks.md` and `docs/asr/report.md` were not edited.

## Правка після рев'ю: статуси в admin.spec

Reviewing-сесія виявила, що пошук `getByRole('status', { name: ... })` не працює для
`status`, бо текст status-регіону не формує доступне ім’я без явного label. У
`frontend/e2e/admin.spec.ts` усі п’ять пошуків статусу звужено текстовим
`getByRole('status').filter({ hasText: ... })`; C19 використовує регулярний вираз для
`Відсутні транскрипти`. Продуктовий код не змінювався.

Перевірки після правки:

- `npm run format:check` — passed.
- `npm run lint` — passed.
- `npm run typecheck` — passed.
- `npx playwright test e2e/admin.spec.ts --list --workers=1` — passed; 8 tests listed.
