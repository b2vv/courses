import type { DocumentErrorCode, DocumentResult } from '../domain/document';
import type { DocumentRepository } from '../ports/DocumentRepository';
import { getLearnLessonDocument } from '../../../adapters/api/generated/courses';
import { documentErrorSchema, lessonDocumentSchema } from './document-schema';
import { mapDocument } from './document-mapping';
import { createApiTransport, unnamedStatusCode } from '../../../adapters/api/transport';
import type { TransportOptions, WireResponse } from '../../../adapters/api/transport';

function failure(code: DocumentErrorCode): DocumentResult<never> {
  return { ok: false, error: { code } };
}

function namedError(
  response: WireResponse,
  allowed: Readonly<Record<number, readonly DocumentErrorCode[]>>,
): DocumentErrorCode {
  const parsed = documentErrorSchema.safeParse(response.data);
  if (parsed.success) {
    const code = allowed[response.status]?.find(
      (candidate) => candidate === parsed.data.error.code,
    );
    if (code !== undefined) return code;
  }
  return unnamedStatusCode(response.status) as DocumentErrorCode;
}

/** One repository per browser/SSR request scope; generated transport never escapes this adapter. */
export function createApiDocumentRepository(options: TransportOptions = {}): DocumentRepository {
  const send = createApiTransport(options);
  return {
    async getLearnLessonDocument(lessonId, locale, signal) {
      const outcome = await send(
        'learn.document',
        (init) =>
          getLearnLessonDocument(lessonId, locale === undefined ? undefined : { locale }, init),
        200,
        signal,
      );
      if (!outcome.ok) return failure(outcome.code as DocumentErrorCode);
      if (outcome.status !== 200)
        return failure(
          namedError(outcome, {
            400: ['invalid_request'],
            401: ['authentication_required'],
            403: ['course_forbidden'],
            404: ['lesson_not_found'],
            503: ['dependency_unavailable'],
          }),
        );
      const parsed = lessonDocumentSchema.safeParse(outcome.data);
      if (!parsed.success) return failure('invalid_request');
      return { ok: true, value: mapDocument(parsed.data) };
    },
  };
}
