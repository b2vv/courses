import { afterAll, afterEach, beforeAll, expect, test } from '@rstest/core';
import { http, HttpResponse } from 'msw';
import { setupServer } from 'msw/node';
import { apiOrigin } from './api-origin';
import { createApiTranscriptRepository } from './transcript-repository';

const lessonId = '30000000-0000-4000-8000-000000000001';
const documentId = '60000000-0000-4000-8000-000000000001';
const route = `${apiOrigin()}/api/v1/learn/lessons/:lessonId/transcript`;

const body = {
  lessonId,
  language: 'uk',
  stale: false,
  segments: [
    {
      id: '70000000-0000-4000-8000-000000000001',
      startMs: 0,
      endMs: 1000,
      text: 'перша репліка',
      speaker: 'Лектор',
      documentIds: [documentId],
    },
    {
      id: '70000000-0000-4000-8000-000000000002',
      startMs: 1000,
      endMs: 2000,
      text: 'друга репліка',
      documentIds: [],
    },
  ],
  documents: [{ id: documentId, label: 'Довідка', url: 'https://example.test/a' }],
};

const server = setupServer();
beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

test('a transcript reads with its library carried once', async () => {
  server.use(http.get(route, () => HttpResponse.json(body)));
  const result = await createApiTranscriptRepository().getLessonTranscript(lessonId);
  expect(result.ok).toBe(true);
  if (!result.ok) return;
  expect(result.value.segments).toHaveLength(2);
  expect(result.value.segments[0]?.documentIds).toEqual([documentId]);
  expect(result.value.segments[0]?.speaker).toBe('Лектор');
  // The second segment carries no speaker: an absent optional must stay absent rather than
  // becoming an empty string that the UI would render as a nameless speaker label.
  expect(result.value.segments[1]?.speaker).toBeUndefined();
  expect(result.value.documents).toHaveLength(1);
});

test('an absent transcript is a value, not an error', async () => {
  server.use(
    http.get(route, () =>
      HttpResponse.json({ lessonId, language: null, stale: false, segments: [], documents: [] }),
    ),
  );
  const result = await createApiTranscriptRepository().getLessonTranscript(lessonId);
  expect(result.ok).toBe(true);
  if (!result.ok) return;
  expect(result.value.language).toBeNull();
  expect(result.value.segments).toHaveLength(0);
});

test('a body about another lesson is refused instead of rendered', async () => {
  // Without this the transcript of one lesson could render under another and every timestamp
  // would point at the wrong video.
  server.use(
    http.get(route, () =>
      HttpResponse.json({ ...body, lessonId: '30000000-0000-4000-8000-000000000009' }),
    ),
  );
  const result = await createApiTranscriptRepository().getLessonTranscript(lessonId);
  expect(result.ok).toBe(false);
});

test('a segment whose end precedes its start is refused, not silently hidden', async () => {
  server.use(
    http.get(route, () =>
      HttpResponse.json({
        ...body,
        segments: [{ ...body.segments[0], startMs: 900, endMs: 100 }],
      }),
    ),
  );
  const result = await createApiTranscriptRepository().getLessonTranscript(lessonId);
  expect(result.ok).toBe(false);
});

test('named refusals keep their code', async () => {
  server.use(
    http.get(route, () =>
      HttpResponse.json({ error: { code: 'course_forbidden' } }, { status: 403 }),
    ),
  );
  const result = await createApiTranscriptRepository().getLessonTranscript(lessonId);
  expect(result.ok).toBe(false);
  if (result.ok) return;
  expect(result.error.code).toBe('course_forbidden');
});

test('the language parameter reaches the wire', async () => {
  let seen: string | null = null;
  server.use(
    http.get(route, ({ request }) => {
      seen = new URL(request.url).searchParams.get('language');
      return HttpResponse.json(body);
    }),
  );
  await createApiTranscriptRepository().getLessonTranscript(lessonId, 'uk');
  expect(seen).toBe('uk');
});
