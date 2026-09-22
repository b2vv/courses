import type { DraftErrorCode, DraftResult, LessonDraft } from '../../domain/transcript/draft';
import type { DraftRepository } from '../../ports/DraftRepository';
import { apiOrigin } from './api-origin';
import { draftErrorSchema, lessonDraftSchema } from './draft-schema';
import { createApiTransport, unnamedStatusCode } from './transport';
import type { TransportOptions, WireResponse } from './transport';

function failure(code: DraftErrorCode): DraftResult<never> {
  return { ok: false, error: { code } };
}

function namedError(
  response: WireResponse,
  allowed: Readonly<Record<number, readonly DraftErrorCode[]>>,
): DraftErrorCode {
  const parsed = draftErrorSchema.safeParse(response.data);
  if (parsed.success) {
    const code = allowed[response.status]?.find(
      (candidate) => candidate === parsed.data.error.code,
    );
    if (code !== undefined) return code;
  }
  return unnamedStatusCode(response.status) as DraftErrorCode;
}

async function fetchJson(url: string, init: RequestInit): Promise<WireResponse> {
  const res = await fetch(url, init);
  const body = [202, 204, 205, 304].includes(res.status) ? null : await res.text();
  const data = body ? JSON.parse(body) : {};
  return { data, status: res.status };
}

export function createApiDraftRepository(options: TransportOptions = {}): DraftRepository {
  const send = createApiTransport(options);
  return {
    async getDraft(lessonId, signal) {
      const outcome = await send(
        'admin.draft.get',
        (init) =>
          fetchJson(`${apiOrigin()}/api/v1/admin/lessons/${lessonId}/draft`, {
            ...init,
            method: 'GET',
          }),
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
      const parsed = lessonDraftSchema.safeParse(outcome.data);
      if (!parsed.success || parsed.data.lessonId !== lessonId) return failure('invalid_response');
      const draft: LessonDraft = {
        lessonId: parsed.data.lessonId,
        state: parsed.data.state,
        language: parsed.data.language,
        failure: parsed.data.failure,
        stale: parsed.data.stale,
        updatedAt: parsed.data.updatedAt,
        assetId: parsed.data.assetId,
        generation: parsed.data.generation,
        segments: parsed.data.segments.map((s) => ({
          id: s.id,
          startMs: s.startMs,
          endMs: s.endMs,
          text: s.text,
          ...(s.speaker === undefined ? {} : { speaker: s.speaker }),
          ...(s.confidence === undefined ? {} : { confidence: s.confidence }),
        })),
      };
      return { ok: true, value: draft };
    },

    async regenerate(lessonId, language, signal) {
      const body =
        language === undefined || language === '' ? undefined : JSON.stringify({ language });
      const outcome = await send(
        'admin.draft.regenerate',
        (init) =>
          fetchJson(`${apiOrigin()}/api/v1/admin/lessons/${lessonId}/draft/regenerate`, {
            ...init,
            method: 'POST',
            ...(body === undefined
              ? {}
              : {
                  body,
                  headers: (() => {
                    const h = new Headers(init.headers);
                    h.set('Content-Type', 'application/json');
                    return h;
                  })(),
                }),
          }),
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
            404: ['lesson_not_found', 'asset_not_found'],
            409: ['already_running'],
            503: ['dependency_unavailable'],
          }),
        );
      return { ok: true, value: undefined };
    },
  };
}
