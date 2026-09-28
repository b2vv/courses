import type {
  DraftErrorCode,
  DraftResult,
  DraftState,
  LessonDraft,
} from '../../domain/transcript/draft';
import type { DraftRepository } from '../../ports/DraftRepository';
import { getLessonDraft, regenerateLessonDraft } from './generated/courses';
import type {
  DraftSegment as GeneratedDraftSegment,
  LessonDraft as GeneratedLessonDraft,
} from './generated/models';
import { createApiTransport, unnamedStatusCode } from './transport';
import type { TransportOptions, WireResponse } from './transport';

function failure(code: DraftErrorCode): DraftResult<never> {
  return { ok: false, error: { code } };
}

function isRecord(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null;
}

function hasOnlyKeys(value: Record<string, unknown>, keys: readonly string[]): boolean {
  return Object.keys(value).every((key) => keys.includes(key));
}

function isUuid(value: unknown): value is string {
  return (
    typeof value === 'string' &&
    /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i.test(value)
  );
}

function runeLength(value: string): number {
  return Array.from(value).length;
}

function isLanguage(value: unknown): value is string {
  return (
    typeof value === 'string' &&
    value.length >= 2 &&
    value.length <= 35 &&
    /^[a-z][a-z0-9]{1,7}(?:-[a-z0-9]{1,8})*$/.test(value)
  );
}

function isOptionalString(value: unknown): value is string | undefined {
  return value === undefined || typeof value === 'string';
}

function isGeneratedDraftSegment(value: unknown): value is GeneratedDraftSegment {
  if (
    !isRecord(value) ||
    !hasOnlyKeys(value, ['id', 'startMs', 'endMs', 'text', 'speaker', 'confidence'])
  ) {
    return false;
  }
  return (
    isUuid(value.id) &&
    typeof value.startMs === 'number' &&
    Number.isInteger(value.startMs) &&
    value.startMs >= 0 &&
    typeof value.endMs === 'number' &&
    Number.isInteger(value.endMs) &&
    value.endMs >= value.startMs &&
    typeof value.text === 'string' &&
    runeLength(value.text) >= 1 &&
    runeLength(value.text) <= 1000 &&
    isOptionalString(value.speaker) &&
    (value.speaker === undefined ||
      (runeLength(value.speaker) >= 1 && runeLength(value.speaker) <= 100)) &&
    (value.confidence === undefined ||
      (typeof value.confidence === 'number' &&
        Number.isFinite(value.confidence) &&
        value.confidence >= 0 &&
        value.confidence <= 1))
  );
}

function isGeneratedLessonDraft(value: unknown): value is GeneratedLessonDraft {
  if (
    !isRecord(value) ||
    !hasOnlyKeys(value, [
      'lessonId',
      'state',
      'language',
      'failure',
      'stale',
      'updatedAt',
      'assetId',
      'generation',
      'segments',
    ])
  ) {
    return false;
  }
  const validState =
    value.state === null ||
    value.state === 'queued' ||
    value.state === 'running' ||
    value.state === 'ready' ||
    value.state === 'failed' ||
    value.state === 'no_speech';
  const validLanguage = value.language === null || isLanguage(value.language);
  const validFailure =
    value.failure === null ||
    (typeof value.failure === 'string' &&
      runeLength(value.failure) >= 1 &&
      runeLength(value.failure) <= 200);
  const validUpdatedAt =
    value.updatedAt === null || (typeof value.updatedAt === 'string' && value.updatedAt.length > 0);
  const validGeneration =
    value.generation === null ||
    (typeof value.generation === 'number' &&
      Number.isInteger(value.generation) &&
      value.generation >= 1);
  return (
    isUuid(value.lessonId) &&
    validState &&
    validLanguage &&
    validFailure &&
    typeof value.stale === 'boolean' &&
    validUpdatedAt &&
    (value.assetId === null || isUuid(value.assetId)) &&
    validGeneration &&
    Array.isArray(value.segments) &&
    value.segments.length <= 5000 &&
    value.segments.every(isGeneratedDraftSegment)
  );
}

function domainState(state: GeneratedLessonDraft['state']): DraftState | null | undefined {
  switch (state) {
    case null:
      return null;
    case 'queued':
      return 'pending';
    case 'running':
    case 'ready':
    case 'failed':
    case 'no_speech':
      return state;
    default:
      return undefined;
  }
}

function mapDraft(raw: unknown, lessonId: string): LessonDraft | undefined {
  if (!isGeneratedLessonDraft(raw) || raw.lessonId !== lessonId) return undefined;
  const state = domainState(raw.state);
  if (state === undefined) return undefined;
  return {
    lessonId: raw.lessonId,
    state,
    language: raw.language,
    failure: raw.failure,
    stale: raw.stale,
    updatedAt: raw.updatedAt,
    assetId: raw.assetId,
    generation: raw.generation,
    segments: raw.segments.map((segment) => ({
      id: segment.id,
      startMs: segment.startMs,
      endMs: segment.endMs,
      text: segment.text,
      ...(segment.speaker === undefined ? {} : { speaker: segment.speaker }),
      ...(segment.confidence === undefined ? {} : { confidence: segment.confidence }),
    })),
  };
}

function namedError(
  response: WireResponse,
  allowed: Readonly<Record<number, readonly DraftErrorCode[]>>,
): DraftErrorCode {
  const data = response.data;
  const error = isRecord(data) ? data.error : undefined;
  if (isRecord(error) && typeof error.code === 'string') {
    const code = allowed[response.status]?.find((candidate) => candidate === error.code);
    if (code !== undefined) return code;
  }
  return unnamedStatusCode(response.status) as DraftErrorCode;
}

/** One repository per browser/SSR request scope; generated transport stays inside this adapter. */
export function createApiDraftRepository(options: TransportOptions = {}): DraftRepository {
  const send = createApiTransport(options);
  return {
    async getDraft(lessonId, signal) {
      const outcome = await send(
        'admin.draft.get',
        (init) => getLessonDraft(lessonId, init),
        200,
        signal,
      );
      if (!outcome.ok) return failure(outcome.code as DraftErrorCode);
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
      const draft = mapDraft(outcome.data, lessonId);
      return draft === undefined ? failure('invalid_response') : { ok: true, value: draft };
    },

    async regenerate(lessonId, language, signal) {
      const request = language === undefined || language === '' ? undefined : { language };
      const outcome = await send(
        'admin.draft.regenerate',
        (init) => regenerateLessonDraft(lessonId, request, init),
        202,
        signal,
      );
      if (!outcome.ok) return failure(outcome.code as DraftErrorCode);
      if (outcome.status !== 202)
        return failure(
          namedError(outcome, {
            400: ['invalid_request'],
            401: ['authentication_required'],
            403: ['origin_forbidden', 'course_forbidden'],
            404: ['lesson_not_found'],
            409: ['already_running'],
            503: ['dependency_unavailable'],
          }),
        );
      return { ok: true, value: undefined };
    },
  };
}
