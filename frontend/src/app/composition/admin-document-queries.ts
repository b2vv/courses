import { mutationOptions, queryOptions } from '@tanstack/react-query';
import type { QueryClient } from '@tanstack/react-query';
import type {
  DocumentAuthoringErrorCode,
  ImageUploadErrorCode,
  LessonDocument,
  UploadedImage,
} from '../../domain/document/document';
import type { DocumentAuthoringRepository } from '../../ports/DocumentAuthoringRepository';

export class DocumentAuthoringError extends Error {
  readonly code: DocumentAuthoringErrorCode;
  constructor(code: DocumentAuthoringErrorCode) {
    super(code);
    this.name = 'DocumentAuthoringError';
    this.code = code;
  }
}

export class ImageUploadError extends Error {
  readonly code: ImageUploadErrorCode;
  constructor(code: ImageUploadErrorCode) {
    super(code);
    this.name = 'ImageUploadError';
    this.code = code;
  }
}

export function adminLessonDocumentQueryKey(lessonId: string, locale?: 'uk' | 'en') {
  return ['admin', 'document', lessonId, locale ?? null] as const;
}

export function createAdminDocumentQueries(documents: DocumentAuthoringRepository) {
  return {
    documents,
    document: (lessonId: string, locale?: 'uk' | 'en') =>
      queryOptions({
        queryKey: adminLessonDocumentQueryKey(lessonId, locale),
        queryFn: async ({ signal }) => {
          const result = await documents.getDraft(lessonId, locale, signal);
          if (!result.ok) throw new DocumentAuthoringError(result.error.code);
          return result.value;
        },
      }),
    save: (client: QueryClient) =>
      mutationOptions({
        mutationFn: async ({
          lessonId,
          document,
          expectedUpdatedAt,
          locale,
        }: {
          readonly lessonId: string;
          readonly document: LessonDocument;
          readonly expectedUpdatedAt: string | null;
          readonly locale?: 'uk' | 'en';
        }) => {
          const result = await documents.saveDraft(lessonId, document, expectedUpdatedAt, locale);
          if (!result.ok) throw new DocumentAuthoringError(result.error.code);
          return result.value;
        },
        // Re-read the draft so the next save carries the fresh optimistic-lock token.
        onSuccess: (_value, { lessonId, locale }) =>
          client.invalidateQueries({
            queryKey: adminLessonDocumentQueryKey(lessonId, locale),
            exact: true,
          }),
      }),
    publish: (client: QueryClient) =>
      mutationOptions({
        mutationFn: async ({
          lessonId,
          expectedUpdatedAt,
          locale,
        }: {
          readonly lessonId: string;
          readonly expectedUpdatedAt: string;
          readonly locale?: 'uk' | 'en';
        }) => {
          const result = await documents.publish(lessonId, expectedUpdatedAt, locale);
          if (!result.ok) throw new DocumentAuthoringError(result.error.code);
          return result.value;
        },
        // Publish moves draft into published: the admin draft token changes and every learner's
        // reader for this lesson is now stale.
        onSuccess: (_value, { lessonId, locale }) =>
          Promise.all([
            client.invalidateQueries({
              queryKey: adminLessonDocumentQueryKey(lessonId, locale),
              exact: true,
            }),
            // Every locale of the learner's reader for this lesson is now stale, so match by prefix.
            client.invalidateQueries({ queryKey: ['learn', 'document', lessonId] }),
          ]),
      }),
    uploadImage: () =>
      mutationOptions({
        mutationFn: async ({
          lessonId,
          file,
        }: {
          readonly lessonId: string;
          readonly file: File;
        }): Promise<UploadedImage> => {
          const result = await documents.uploadImage(lessonId, file);
          if (!result.ok) throw new ImageUploadError(result.error.code);
          return result.value;
        },
      }),
  };
}
