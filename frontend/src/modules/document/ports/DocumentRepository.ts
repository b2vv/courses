import type { DocumentResult, LessonDocument } from '../domain/document';

export interface DocumentRepository {
  /**
   * The published document of one lesson in one locale. An absent publication is not a failure:
   * the endpoint returns an empty valid document, which renders the same as a lesson with no body.
   */
  getLearnLessonDocument(
    lessonId: string,
    locale?: 'uk' | 'en',
    signal?: AbortSignal,
  ): Promise<DocumentResult<LessonDocument>>;
}
