import { afterAll, afterEach, beforeAll, describe, expect, test } from '@rstest/core';
import { http, HttpResponse } from 'msw';
import { setupServer } from 'msw/node';
import { apiOrigin } from '../../../adapters/api/api-origin';
import { createApiDraftRepository } from './draft-repository';

const lessonId = '20000000-0000-4000-8000-000000000001';
const trackId = '60000000-0000-4000-8000-000000000001';
const otherTrackId = '60000000-0000-4000-8000-000000000002';
const api = `${apiOrigin()}/api/v1/admin/lessons/${lessonId}/drafts`;
const wireDraft = {
  trackId,
  trackLanguage: 'uk',
  trackMultilingual: true,
  state: 'ready',
  language: 'uk',
  failure: null,
  stale: false,
  updatedAt: '2026-11-15T12:00:00.123456Z',
  assetId: '50000000-0000-4000-8000-000000000001',
  generation: 1,
  segments: [
    {
      id: '70000000-0000-4000-8000-000000000001',
      startMs: 0,
      endMs: 1000,
      text: 'Вітаю',
      speaker: 'Lecturer',
      confidence: 0.9,
    },
  ],
};
const invalid = { ok: false, error: { code: 'invalid_response' } };
const server = setupServer();
beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('draft API repository', () => {
  test('returns all drafts in server order with every track snapshot', async () => {
    const other = { ...wireDraft, trackId: otherTrackId, language: 'en', trackLanguage: 'en' };
    server.use(http.get(api, () => HttpResponse.json({ drafts: [wireDraft, other] })));
    expect(await createApiDraftRepository().listDrafts(lessonId)).toEqual({
      ok: true,
      value: [
        { ...wireDraft, lessonId },
        { ...other, lessonId },
      ],
    });
  });

  test('returns an empty list when no tracks have drafts', async () => {
    server.use(http.get(api, () => HttpResponse.json({ drafts: [] })));
    expect(await createApiDraftRepository().listDrafts(lessonId)).toEqual({ ok: true, value: [] });
  });

  test.each([
    { name: 'unknown envelope key', response: { drafts: [wireDraft], extra: true } },
    { name: 'missing drafts', response: {} },
    { name: 'unknown first-item key', response: { drafts: [{ ...wireDraft, extra: true }] } },
    {
      name: 'unknown later-item key',
      response: { drafts: [wireDraft, { ...wireDraft, extra: true }] },
    },
    {
      name: 'invalid later-item snapshot',
      response: { drafts: [wireDraft, { ...wireDraft, trackLanguage: 'UK' }] },
    },
    {
      name: 'invalid later-item boolean',
      response: { drafts: [wireDraft, { ...wireDraft, trackMultilingual: 'true' }] },
    },
    {
      name: 'missing track snapshot',
      response: { drafts: [{ ...wireDraft, trackMultilingual: undefined }] },
    },
    { name: 'invalid track id', response: { drafts: [{ ...wireDraft, trackId: 'bad' }] } },
    { name: 'null state', response: { drafts: [{ ...wireDraft, state: null }] } },
    { name: 'null asset', response: { drafts: [{ ...wireDraft, assetId: null }] } },
    { name: 'null generation', response: { drafts: [{ ...wireDraft, generation: null }] } },
    { name: 'null timestamp', response: { drafts: [{ ...wireDraft, updatedAt: null }] } },
    { name: 'invalid timestamp', response: { drafts: [{ ...wireDraft, updatedAt: 'bad' }] } },
    {
      name: 'unknown segment key',
      response: {
        drafts: [{ ...wireDraft, segments: [{ ...wireDraft.segments[0], extra: true }] }],
      },
    },
  ])('rejects the whole list for $name', async ({ response }) => {
    server.use(http.get(api, () => HttpResponse.json(response)));
    expect(await createApiDraftRepository().listDrafts(lessonId)).toEqual(invalid);
  });

  test('maps queued to pending and permits RFC3339 timestamps with no fractional seconds', async () => {
    server.use(
      http.get(api, () =>
        HttpResponse.json({
          drafts: [
            {
              ...wireDraft,
              state: 'queued',
              language: null,
              updatedAt: '2026-11-15T12:00:00Z',
              segments: [],
            },
          ],
        }),
      ),
    );
    expect(await createApiDraftRepository().listDrafts(lessonId)).toMatchObject({
      ok: true,
      value: [{ trackId, state: 'pending', language: null }],
    });
  });

  test('regenerates exactly the requested shown track with no body', async () => {
    const bodies: string[] = [];
    server.use(
      http.post(`${api}/${otherTrackId}/regenerate`, async ({ request }) => {
        bodies.push(await request.text());
        return new HttpResponse(null, { status: 202 });
      }),
    );
    expect(await createApiDraftRepository().regenerate(lessonId, otherTrackId)).toEqual({
      ok: true,
      value: undefined,
    });
    expect(bodies).toEqual(['']);
  });

  test.each([
    { status: 404, code: 'track_not_found', expected: 'track_not_found' },
    { status: 404, code: 'lesson_not_found', expected: 'lesson_not_found' },
    { status: 409, code: 'already_running', expected: 'already_running' },
    { status: 503, code: 'track_not_found', expected: 'dependency_unavailable' },
  ])('maps $status $code to $expected', async ({ status, code, expected }) => {
    server.use(
      http.post(`${api}/${trackId}/regenerate`, () =>
        HttpResponse.json(
          {
            error: { code, message: 'x', requestId: 'a'.repeat(32) },
          },
          { status },
        ),
      ),
    );
    expect(await createApiDraftRepository().regenerate(lessonId, trackId)).toEqual({
      ok: false,
      error: { code: expected },
    });
  });
});

describe('transcript inventory', () => {
  const url = `${apiOrigin()}/api/v1/admin/lessons/${lessonId}/transcripts`;
  const item = { language: 'en-us', updatedAt: wireDraft.updatedAt, stale: true, isPrimary: false };
  test.each([
    { items: [] },
    { items: [item] },
    { items: [item, { ...item, language: 'uk', isPrimary: true }] },
  ])('parses inventory $items', async ({ items }) => {
    server.use(http.get(url, () => HttpResponse.json({ transcripts: items })));
    expect(await createApiDraftRepository().listTranscripts(lessonId)).toEqual({
      ok: true,
      value: items,
    });
  });
  test.each([
    {},
    { transcripts: null },
    { transcripts: [item], extra: true },
    { transcripts: [item, { ...item, extra: true }] },
    { transcripts: [{ ...item, language: 'EN' }] },
    { transcripts: [{ ...item, updatedAt: 'bad' }] },
    { transcripts: [{ ...item, stale: 'true' }] },
    { transcripts: [{ ...item, isPrimary: undefined }] },
  ])('rejects malformed inventory %j', async (body) => {
    server.use(http.get(url, () => HttpResponse.json(body)));
    expect(await createApiDraftRepository().listTranscripts(lessonId)).toEqual(invalid);
  });
  test.each([
    [400, 'invalid_request'],
    [401, 'authentication_required'],
    [403, 'course_forbidden'],
    [404, 'lesson_not_found'],
    [503, 'dependency_unavailable'],
  ] as const)('maps inventory error %s %s', async (status, code) => {
    server.use(http.get(url, () => HttpResponse.json({ error: { code } }, { status })));
    expect(await createApiDraftRepository().listTranscripts(lessonId)).toEqual({
      ok: false,
      error: { code },
    });
  });
});
