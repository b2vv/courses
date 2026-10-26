import type { LessonTranscript, TranscriptErrorCode, TranscriptResult } from '../domain/transcript';
import type { TranscriptRepository } from '../ports/TranscriptRepository';
import {
  addCourseDocument,
  getLessonTranscript,
  listCourseDocuments,
  removeCourseDocument,
  saveLessonTranscript,
} from '../../../adapters/api/generated/courses';
import {
  courseDocumentsSchema,
  lessonTranscriptSchema,
  savedTranscriptSchema,
  transcriptErrorSchema,
} from './transcript-schema';
import { createApiTransport, unnamedStatusCode } from '../../../adapters/api/transport';
import type { TransportOptions, WireResponse } from '../../../adapters/api/transport';

function failure(code: TranscriptErrorCode): TranscriptResult<never> {
  return { ok: false, error: { code } };
}

function namedError(
  response: WireResponse,
  allowed: Readonly<Record<number, readonly TranscriptErrorCode[]>>,
): TranscriptErrorCode {
  const parsed = transcriptErrorSchema.safeParse(response.data);
  if (parsed.success) {
    const code = allowed[response.status]?.find(
      (candidate) => candidate === parsed.data.error.code,
    );
    if (code !== undefined) return code;
  }
  return unnamedStatusCode(response.status) as TranscriptErrorCode;
}

/** One repository per browser/SSR request scope; generated transport never escapes this adapter. */
export function createApiTranscriptRepository(
  options: TransportOptions = {},
): TranscriptRepository {
  const send = createApiTransport(options);
  return {
    async getLessonTranscript(lessonId, language, signal) {
      const outcome = await send(
        'learn.transcript',
        (init) =>
          getLessonTranscript(lessonId, language === undefined ? undefined : { language }, init),
        200,
        signal,
      );
      if (!outcome.ok) return failure(outcome.code as TranscriptErrorCode);
      if (outcome.status !== 200)
        return failure(
          namedError(outcome, {
            400: ['invalid_request'],
            401: ['authentication_required'],
            403: ['course_forbidden', 'course_purchase_required', 'email_not_verified'],
            404: ['lesson_not_found'],
            503: ['dependency_unavailable'],
          }),
        );
      const parsed = lessonTranscriptSchema.safeParse(outcome.data);
      // The lesson in the body must be the lesson that was asked for. Without this a cached or
      // mismatched response would render under the wrong lesson and every timestamp would lie.
      if (!parsed.success || parsed.data.lessonId !== lessonId) return failure('invalid_request');
      const transcript: LessonTranscript = {
        lessonId: parsed.data.lessonId,
        language: parsed.data.language,
        updatedAt: parsed.data.updatedAt,
        stale: parsed.data.stale,
        segments: parsed.data.segments.map((segment) => ({
          id: segment.id,
          startMs: segment.startMs,
          endMs: segment.endMs,
          text: segment.text,
          ...(segment.speaker === undefined ? {} : { speaker: segment.speaker }),
          ...(segment.confidence === undefined ? {} : { confidence: segment.confidence }),
          documentIds: segment.documentIds,
        })),
        documents: parsed.data.documents.map((document) => ({
          id: document.id,
          label: document.label,
          url: document.url,
        })),
      };
      return { ok: true, value: transcript };
    },

    async saveLessonTranscript(lessonId, draft, signal) {
      const outcome = await send(
        'learn.transcript.save',
        (init) =>
          saveLessonTranscript(
            lessonId,
            {
              language: draft.language,
              ...(draft.expectedUpdatedAt === undefined
                ? {}
                : { expectedUpdatedAt: draft.expectedUpdatedAt }),
              segments: draft.segments.map((segment) => ({
                id: segment.id,
                startMs: segment.startMs,
                endMs: segment.endMs,
                text: segment.text,
                speaker: segment.speaker ?? null,
                confidence: segment.confidence ?? null,
                documentIds: [...segment.documentIds],
              })),
            },
            init,
          ),
        200,
        signal,
      );
      if (!outcome.ok) return failure(outcome.code as TranscriptErrorCode);
      if (outcome.status !== 200) return failure(namedError(outcome, writeRefusals));
      const parsed = savedTranscriptSchema.safeParse(outcome.data);
      if (!parsed.success) return failure('invalid_request');
      return { ok: true, value: { updatedAt: parsed.data.updatedAt } };
    },

    async listCourseDocuments(courseId, signal) {
      const outcome = await send(
        'learn.transcript.documents',
        (init) => listCourseDocuments(courseId, init),
        200,
        signal,
      );
      if (!outcome.ok) return failure(outcome.code as TranscriptErrorCode);
      if (outcome.status !== 200) return failure(namedError(outcome, writeRefusals));
      const parsed = courseDocumentsSchema.safeParse(outcome.data);
      if (!parsed.success) return failure('invalid_request');
      return { ok: true, value: parsed.data.documents };
    },

    async addCourseDocument(courseId, document, signal) {
      const outcome = await send(
        'learn.transcript.documents.add',
        (init) => addCourseDocument(courseId, document, init),
        201,
        signal,
      );
      if (!outcome.ok) return failure(outcome.code as TranscriptErrorCode);
      if (outcome.status !== 201) return failure(namedError(outcome, writeRefusals));
      return { ok: true, value: undefined };
    },

    async removeCourseDocument(courseId, documentId, signal) {
      const outcome = await send(
        'learn.transcript.documents.remove',
        (init) => removeCourseDocument(courseId, documentId, init),
        204,
        signal,
      );
      if (!outcome.ok) return failure(outcome.code as TranscriptErrorCode);
      if (outcome.status !== 204) return failure(namedError(outcome, writeRefusals));
      return { ok: true, value: undefined };
    },
  };
}

// Every refusal the write surface can name. A code outside this map degrades to the status-derived
// one rather than being shown to the author as itself: an unnamed code is a contract drift, and
// rendering it as a message would make the drift look like a product state.
const writeRefusals: Readonly<Record<number, readonly TranscriptErrorCode[]>> = {
  400: [
    'invalid_request',
    'transcript_segment_limit',
    'transcript_text_limit',
    'transcript_document_limit',
  ],
  401: ['authentication_required'],
  403: ['course_forbidden'],
  404: ['lesson_not_found'],
  409: [
    'transcript_modified',
    'transcript_exists',
    'transcript_document_referenced',
    'transcript_document_exists',
  ],
  503: ['dependency_unavailable'],
};
