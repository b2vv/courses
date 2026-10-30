// Public surface of the `document` feature module. Everything another module, the composition
// roots or the routes may use is re-exported here; nothing outside this folder imports the module's
// internals directly (enforced by scripts/check-architecture.mjs — no deep cross-module import).

// Domain: lesson-document types, codes and pure helpers.
export * from './domain/document';

// Ports: the I/O contracts this feature depends on.
export type { DocumentAuthoringRepository } from './ports/DocumentAuthoringRepository';
export type { DocumentRepository } from './ports/DocumentRepository';

// Application: query/mutation factories and their error types.
export {
  createAdminDocumentQueries,
  createTeachDocumentQueries,
  teachLessonDocumentQueryKey,
  adminLessonDocumentQueryKey,
  DocumentAuthoringError,
  ImageUploadError,
} from './application/admin-document-queries';
export {
  createDocumentQueries,
  lessonDocumentQueryKey,
  DocumentQueryError,
} from './application/document-queries';

// Adapters: the HTTP implementations wired by the composition roots.
export {
  createApiDocumentAuthoringRepository,
  createApiTeachDocumentAuthoringRepository,
} from './adapters/admin-document-repository';
export { createApiDocumentRepository } from './adapters/document-repository';

// UI. `LessonDocumentEditor` is deliberately NOT value-exported here: it pulls in Tiptap, and the
// lesson route loads it through a dynamic `import('./ui/LessonDocumentEditor')` so that dependency
// stays out of the shared bundle. A value re-export would drag Tiptap into every eager importer of
// this barrel (e.g. the composition roots that only need the queries). Only the plain `LessonMeta`
// type is surfaced (type-only re-exports are erased and carry no runtime weight).
export type { LessonMeta } from './ui/LessonDocumentEditor';
export { LessonDocumentEditorLoading } from './ui/LessonDocumentEditorLoading';
export { LessonDocumentView } from './ui/LessonDocumentView';
