# document-module importer map (A-пілот) — 2026-10-16

Джерело: Explore-агент (ripgrep над frontend/src). [MOVE]=у наборі що переїжджає; [REMAIN]=лишається, переписати import.

## Файли document-вертикалі (кандидати на переїзд)
- domain: `domain/document/document.ts` (+test) — імпортує НІЧОГО (domain-правило тримається).
- application: `app/composition/admin-document-queries.ts` (+test) [authoring]; `app/composition/document-queries.ts` [read, НЕ був у 11].
- ports: `ports/DocumentAuthoringRepository.ts`; `ports/DocumentRepository.ts` [read, НЕ був у 11].
- adapters: `adapters/api/document-repository.ts` [read], `admin-document-repository.ts` [authoring];
  залежать від `document-mapping.ts`, `document-schema.ts`, `admin-document-schema.ts` (document-специфічні,
  мають їхати разом) + СПІЛЬНА інфра `generated/courses`, `transport` (лишається спільною).
- ui admin: LessonDocumentEditor(+test), LessonDocumentRichText, LessonDocumentEditorLoading,
  tiptap-document(+test).
- ui learn (read renderer): LessonDocumentView(+test).

## НЕ в модуль
- `DocumentLibrary.tsx` (+test): єдина доменна залежність — `domain/transcript/transcript`; споживач —
  `CourseEditorView.tsx`. Це transcript/course-фіча, НЕ document. Лишається в presentation/admin.

## REMAIN-файли, яким переписати import на `modules/document` (index): 
- ports/DocumentRepository.ts → domain/document (якщо read лишається flat) / у модуль (якщо read теж їде)
- adapters/api/document-mapping.ts → domain/document
- app/composition/document-queries.ts → domain/document (read)
- app/composition/learn.ts → adapters/api/document-repository (read)
- app/composition/admin.ts:20-21,40 → admin-document-repository + admin-document-queries (authoring)
- routes/route-guards.test.tsx:23 → admin-document-queries
- routes/admin.courses.$courseId_.lessons.$lessonId.tsx:6,7,14 → LessonDocumentEditor(×2: type+lazy) + Loading
- routes/learn.$.tsx:19 → LessonDocumentView
- presentation/admin/CourseEditorView.tsx:30 → DocumentLibrary (лишається flat, шлях не міняється якщо DL не їде)

## Вихідні залежності файлів-переселенців на код, що ЛИШАЄТЬСЯ (правити відносні шляхи):
- UI → `../ui`, `../ui/states` (спільний ui-кіт лишається).
- adapters → `generated/courses`, `transport` (спільна інфра лишається в adapters/api).
- admin.ts wiring: `createAdminDocumentQueries(createApiDocumentAuthoringRepository(options))`.

## Відкрите для власника (депт пілоту) — див. питання в сесії.
