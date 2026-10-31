import type {
  DocumentAuthoringResult,
  ImageUploadResult,
  LessonDocument,
  LessonDocumentDraft,
} from '../domain/document';

export interface DocumentAuthoringRepository {
  getDraft(
    lessonId: string,
    locale?: 'uk' | 'en',
    signal?: AbortSignal,
  ): Promise<DocumentAuthoringResult<LessonDocumentDraft>>;

  saveDraft(
    lessonId: string,
    document: LessonDocument,
    expectedUpdatedAt: string | null,
    locale?: 'uk' | 'en',
    signal?: AbortSignal,
  ): Promise<DocumentAuthoringResult<{ readonly updatedAt: string }>>;

  publish(
    lessonId: string,
    expectedUpdatedAt: string,
    locale?: 'uk' | 'en',
    signal?: AbortSignal,
  ): Promise<DocumentAuthoringResult<void>>;

  uploadImage(lessonId: string, file: File, signal?: AbortSignal): Promise<ImageUploadResult>;
}
