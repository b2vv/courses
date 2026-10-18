// Public surface of the `transcript` feature module (lesson transcripts, ASR drafts, the WebVTT
// reader/editor and unsaved-line recovery). Everything another module, the composition roots or the
// routes may use is re-exported here; nothing outside this folder imports the module's internals
// directly (enforced by scripts/check-architecture.mjs — no deep cross-module import).

// Domain: transcript + draft types, codes and pure helpers.
export * from './domain/transcript';
export * from './domain/draft';

// Application: recovery offer/snapshot/baseline and capture directives.
export * from './application/recovery';

// Ports: the I/O contracts this feature depends on.
export type { TranscriptDraft, TranscriptRepository } from './ports/TranscriptRepository';
export type { DraftRepository } from './ports/DraftRepository';
export type { TranscriptHistoryStore } from './ports/TranscriptHistoryStore';

// Adapters: the HTTP and history-store implementations wired by the composition roots.
export { createApiTranscriptRepository } from './adapters/transcript-repository';
export { createApiDraftRepository } from './adapters/draft-repository';
export { createMemoryTranscriptHistory } from './adapters/transcript-history-memory';
export { createIndexedDbTranscriptHistory } from './adapters/transcript-history-indexed-database';

// UI.
export { TranscriptPanel } from './ui/TranscriptPanel';
export { TranscriptEditor } from './ui/TranscriptEditor';
export type { TranscriptEditorActions } from './ui/TranscriptEditor';
export { LessonTranscriptField } from './ui/LessonTranscriptField';
export type { CourseEditorTranscripts } from './ui/LessonTranscriptField';
export { LessonDraftField } from './ui/LessonDraftField';
export type { CourseEditorDrafts, LessonDraftFieldActions } from './ui/LessonDraftField';
export type { TranscriptDraftOffer } from './ui/transcript-draft-offer';
export { idleCaptureMs, useTranscriptRecovery } from './ui/use-transcript-recovery';
export type { TranscriptRecovery, TranscriptRecoveryOptions } from './ui/use-transcript-recovery';
