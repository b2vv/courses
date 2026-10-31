import type {
  DocumentAuthoringErrorCode,
  DocumentAuthoringResult,
  ImageUploadErrorCode,
  ImageUploadResult,
  LessonDocument,
  LessonDocumentDraft,
} from '../domain/document';
import type { DocumentAuthoringRepository } from '../ports/DocumentAuthoringRepository';
import {
  getAdminLessonDocument,
  publishAdminLessonDocument,
  saveAdminLessonDocument,
  uploadAdminLessonImage,
  getTeachLessonDocument,
  saveTeachLessonDocument,
  publishTeachLessonDocument,
  uploadTeachLessonImage,
} from '../../../adapters/api/generated/courses';
import { documentErrorSchema } from './document-schema';
import {
  imageUploadSchema,
  lessonDocumentDraftSchema,
  lessonDocumentUpdatedSchema,
} from './admin-document-schema';
import { mapDocument, toWireDocument } from './document-mapping';
import { createApiTransport, unnamedStatusCode } from '../../../adapters/api/transport';
import type { TransportOptions, WireResponse } from '../../../adapters/api/transport';

function authoringFailure(code: DocumentAuthoringErrorCode): DocumentAuthoringResult<never> {
  return { ok: false, error: { code } };
}

function uploadFailure(code: ImageUploadErrorCode): ImageUploadResult {
  return { ok: false, error: { code } };
}

function named<Code extends string>(
  response: WireResponse,
  allowed: Readonly<Record<number, readonly Code[]>>,
): Code {
  const parsed = documentErrorSchema.safeParse(response.data);
  if (parsed.success) {
    const code = allowed[response.status]?.find(
      (candidate) => candidate === parsed.data.error.code,
    );
    if (code !== undefined) return code;
  }
  return unnamedStatusCode(response.status) as Code;
}

// Every refusal each admin surface can name; a code outside the map degrades to the status-derived
// one rather than being shown as itself, since an unnamed code is contract drift.
const writeRefusals: Readonly<Record<number, readonly DocumentAuthoringErrorCode[]>> = {
  400: ['invalid_lesson_document', 'invalid_request'],
  401: ['authentication_required'],
  403: ['forbidden', 'course_forbidden'],
  404: ['lesson_not_found'],
  409: ['document_conflict'],
  503: ['dependency_unavailable'],
};

const uploadRefusals: Readonly<Record<number, readonly ImageUploadErrorCode[]>> = {
  400: ['invalid_image', 'image_too_many_pixels'],
  401: ['authentication_required'],
  403: ['forbidden'],
  404: ['lesson_not_found'],
  413: ['image_too_large'],
  415: ['unsupported_image'],
  503: ['dependency_unavailable'],
};

type DocumentOperations = {
  readonly get: typeof getAdminLessonDocument;
  readonly save: typeof saveAdminLessonDocument;
  readonly publish: typeof publishAdminLessonDocument;
  readonly uploadImage: typeof uploadAdminLessonImage;
};

export function createApiDocumentAuthoringRepository(
  options: TransportOptions = {},
): DocumentAuthoringRepository {
  return createDocumentAuthoringRepository(
    options,
    {
      get: getAdminLessonDocument,
      save: saveAdminLessonDocument,
      publish: publishAdminLessonDocument,
      uploadImage: uploadAdminLessonImage,
    },
    'admin',
  );
}

export function createApiTeachDocumentAuthoringRepository(
  options: TransportOptions = {},
): DocumentAuthoringRepository {
  return createDocumentAuthoringRepository(
    options,
    {
      get: getTeachLessonDocument,
      save: saveTeachLessonDocument,
      publish: publishTeachLessonDocument,
      uploadImage: uploadTeachLessonImage,
    },
    'teach',
  );
}

/** One repository per browser/SSR request scope; generated transport never escapes this adapter. */
function createDocumentAuthoringRepository(
  options: TransportOptions,
  operations: DocumentOperations,
  labelPrefix: 'admin' | 'teach',
): DocumentAuthoringRepository {
  const send = createApiTransport(options);
  return {
    async getDraft(lessonId, locale, signal) {
      const outcome = await send(
        `${labelPrefix}.document`,
        (init) => operations.get(lessonId, locale === undefined ? undefined : { locale }, init),
        200,
        signal,
      );
      if (!outcome.ok) return authoringFailure(outcome.code as DocumentAuthoringErrorCode);
      if (outcome.status !== 200) return authoringFailure(named(outcome, writeRefusals));
      const parsed = lessonDocumentDraftSchema.safeParse(outcome.data);
      if (!parsed.success) return authoringFailure('invalid_request');
      const draft: LessonDocumentDraft = {
        document: mapDocument(parsed.data.document),
        updatedAt: parsed.data.updatedAt,
      };
      return { ok: true, value: draft };
    },

    async saveDraft(lessonId, document: LessonDocument, expectedUpdatedAt, _locale, signal) {
      // The save takes no locale query: the row is chosen by the document's own locale.
      const outcome = await send(
        `${labelPrefix}.document.save`,
        (init) =>
          operations.save(
            lessonId,
            {
              ...(expectedUpdatedAt === null ? {} : { expectedUpdatedAt }),
              document: toWireDocument(document),
            },
            init,
          ),
        200,
        signal,
      );
      if (!outcome.ok) return authoringFailure(outcome.code as DocumentAuthoringErrorCode);
      if (outcome.status !== 200) return authoringFailure(named(outcome, writeRefusals));
      const parsed = lessonDocumentUpdatedSchema.safeParse(outcome.data);
      if (!parsed.success) return authoringFailure('invalid_request');
      return { ok: true, value: { updatedAt: parsed.data.updatedAt } };
    },

    async publish(lessonId, expectedUpdatedAt, locale, signal) {
      const outcome = await send(
        `${labelPrefix}.document.publish`,
        (init) =>
          operations.publish(
            lessonId,
            { expectedUpdatedAt },
            locale === undefined ? undefined : { locale },
            init,
          ),
        200,
        signal,
      );
      if (!outcome.ok) return authoringFailure(outcome.code as DocumentAuthoringErrorCode);
      if (outcome.status !== 200) return authoringFailure(named(outcome, writeRefusals));
      return { ok: true, value: undefined };
    },

    async uploadImage(lessonId, file, signal) {
      const outcome = await send(
        `${labelPrefix}.image.upload`,
        (init) => operations.uploadImage(lessonId, { file }, init),
        201,
        signal,
      );
      if (!outcome.ok) return uploadFailure(outcome.code as ImageUploadErrorCode);
      if (outcome.status !== 201) return uploadFailure(named(outcome, uploadRefusals));
      const parsed = imageUploadSchema.safeParse(outcome.data);
      if (!parsed.success) return uploadFailure('invalid_image');
      return {
        ok: true,
        value: {
          imageId: parsed.data.id,
          contentType: parsed.data.contentType,
          width: parsed.data.width,
          height: parsed.data.height,
        },
      };
    },
  };
}
