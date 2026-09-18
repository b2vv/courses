import type {
  LessonTranscript,
  TranscriptDocument,
  TranscriptResult,
  TranscriptSegment,
} from '../domain/transcript/transcript';

/** The whole transcript of one lesson in one language: the write replaces, it does not merge. */
export type TranscriptDraft = {
  readonly language: string;
  /** Absent when the author believes no transcript exists yet. */
  readonly expectedUpdatedAt?: string;
  readonly segments: readonly TranscriptSegment[];
};

export interface TranscriptRepository {
  getLessonTranscript(
    lessonId: string,
    language?: string,
    signal?: AbortSignal,
  ): Promise<TranscriptResult<LessonTranscript>>;

  saveLessonTranscript(
    lessonId: string,
    draft: TranscriptDraft,
    signal?: AbortSignal,
  ): Promise<TranscriptResult<{ readonly updatedAt: string }>>;

  listCourseDocuments(
    courseId: string,
    signal?: AbortSignal,
  ): Promise<TranscriptResult<readonly TranscriptDocument[]>>;

  addCourseDocument(
    courseId: string,
    document: TranscriptDocument,
    signal?: AbortSignal,
  ): Promise<TranscriptResult<void>>;

  removeCourseDocument(
    courseId: string,
    documentId: string,
    signal?: AbortSignal,
  ): Promise<TranscriptResult<void>>;
}
