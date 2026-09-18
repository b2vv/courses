import type {
  LessonTranscript,
  TranscriptErrorCode,
  TranscriptResult,
} from '../../domain/transcript/transcript';
import type { TranscriptRepository } from '../../ports/TranscriptRepository';
import { getLessonTranscript } from './generated/courses';
import { lessonTranscriptSchema, transcriptErrorSchema } from './transcript-schema';
import { createApiTransport, unnamedStatusCode } from './transport';
import type { TransportOptions, WireResponse } from './transport';

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
            403: ['course_forbidden'],
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
  };
}
