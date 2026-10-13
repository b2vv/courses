import { queryOptions } from '@tanstack/react-query';
import type { DocumentErrorCode } from '../../domain/document/document';
import type { DocumentRepository } from '../../ports/DocumentRepository';

export class DocumentQueryError extends Error {
  readonly code: DocumentErrorCode;
  constructor(code: DocumentErrorCode) {
    super(code);
    this.name = 'DocumentQueryError';
    this.code = code;
  }
}

export function lessonDocumentQueryKey(lessonId: string, locale?: 'uk' | 'en') {
  return ['learn', 'document', lessonId, locale ?? null] as const;
}

export function createDocumentQueries(document: DocumentRepository) {
  return {
    document,
    lessonDocument: (lessonId: string, locale?: 'uk' | 'en') =>
      queryOptions({
        queryKey: lessonDocumentQueryKey(lessonId, locale),
        queryFn: async ({ signal }) => {
          const result = await document.getLearnLessonDocument(lessonId, locale, signal);
          if (!result.ok) throw new DocumentQueryError(result.error.code);
          return result.value;
        },
      }),
  };
}
