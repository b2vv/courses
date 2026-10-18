# A-батч — PR8 (наступний): transcript — СКОУПИНГ (read-only)

## Слід (більший модуль)
- domain/transcript: transcript.ts, draft.ts, webvtt (+ transcript.test, webvtt.test).
- application/transcript: recovery.ts(+test).
- adapters/api: transcript-repository.ts(+test), transcript-schema.ts, draft-repository.ts (+ draft-repository.test?).
- ports: TranscriptRepository.ts, DraftRepository.ts, TranscriptHistoryStore.ts.

## Рішення
- **feature-local, contracts НЕ треба:** flat app/ports споживачі домену — лише ports (переїдуть). draft = transcript-draft (ASR), живе в transcript.
- UI → transcript/ui: TranscriptPanel (learn), TranscriptEditor, LessonTranscriptField, LessonDraftField, transcript-draft-offer (admin — транскриптові поля/редактор).
- **Лишити flat** (репойнт на barrel): CourseEditorView (admin editor), DocumentLibrary (document-модуль; споживає transcript). Перевірити зв'язність при виконанні.
- TranscriptHistoryStore — browser store (localStorage) порт; адаптер storage? Перевірити, де його реалізація (adapters/storage?), і чи тягнути в модуль.
- Нагадування: adapter-тести під node (конфіг ловить modules/*/adapters).

## ВИКОНАТИ ПІСЛЯ: зеленого гейта purchase + merge, від свіжого main.
