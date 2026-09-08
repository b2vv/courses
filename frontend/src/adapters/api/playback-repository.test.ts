import { afterAll, afterEach, beforeAll, expect, test } from '@rstest/core';
import { http, HttpResponse } from 'msw';
import { setupServer } from 'msw/node';
import { apiOrigin } from './api-origin';
import { createApiPlaybackRepository } from './playback-repository';

const lessonId = '30000000-0000-4000-8000-000000000001';
const playbackId = '0123456789abcdef0123456789abcdef';
const course = {
  id: '10000000-0000-4000-8000-000000000001',
  slug: 'go-basics',
  title: 'Go basics',
  summary: 'Practice',
  description: 'Learn Go.',
  status: 'published',
  createdAt: '2026-09-01T00:00:00Z',
  updatedAt: '2026-09-01T00:00:00Z',
  publishedAt: '2026-09-01T00:00:00Z',
  continueLessonId: lessonId,
  modules: [
    {
      id: '20000000-0000-4000-8000-000000000001',
      title: 'Foundations',
      lessons: [
        {
          id: lessonId,
          title: 'First lesson',
          videoState: 'ready',
          position: 12,
          completed: false,
          clientSeq: 4,
        },
      ],
    },
  ],
} as const;
const session = {
  playbackId,
  mediaPath: `/media/p/${playbackId}/hls/50000000-0000-4000-8000-000000000001/master.m3u8`,
  clientSeq: 4,
  heartbeatSeconds: 2,
  ownershipTtlSeconds: 6,
} as const;

const server = setupServer();
beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

test('maps the authenticated learning model and validates the requested slug', async () => {
  server.use(
    http.get(`${apiOrigin()}/api/v1/learn/courses/go-basics`, () => HttpResponse.json(course)),
  );

  expect(await createApiPlaybackRepository().getCourse('go-basics')).toEqual({
    ok: true,
    value: {
      id: course.id,
      slug: course.slug,
      title: course.title,
      summary: course.summary,
      description: course.description,
      status: course.status,
      modules: course.modules,
      continueLessonId: course.continueLessonId,
    },
  });

  server.use(
    http.get(`${apiOrigin()}/api/v1/learn/courses/go-basics`, () =>
      HttpResponse.json({ ...course, slug: 'another-course' }),
    ),
  );
  expect(await createApiPlaybackRepository().getCourse('go-basics')).toEqual({
    ok: false,
    error: { code: 'invalid_response' },
  });
});

test('mints only after an explicit play action and maps the exact session', async () => {
  let requestBody: unknown;
  server.use(
    http.post(`${apiOrigin()}/api/v1/playback/sessions`, async ({ request }) => {
      requestBody = await request.json();
      return HttpResponse.json(session, { status: 201 });
    }),
  );

  const repository = createApiPlaybackRepository();
  expect(requestBody).toBeUndefined();
  expect(await repository.play(lessonId)).toEqual({ ok: true, value: session });
  expect(requestBody).toEqual({ lessonId });
});

test.each([
  [403, 'playback_forbidden'],
  [404, 'lesson_not_found'],
] as const)('maps play status %i to %s', async (status, code) => {
  server.use(
    http.post(`${apiOrigin()}/api/v1/playback/sessions`, () =>
      HttpResponse.json({ error: { code, message: code, requestId: 'test' } }, { status }),
    ),
  );
  expect(await createApiPlaybackRepository().play(lessonId)).toEqual({
    ok: false,
    error: { code },
  });
});

test.each([
  [204, null],
  [403, 'playback_forbidden'],
  [404, 'lesson_not_found'],
  [409, 'playback_taken_over'],
  [410, 'playback_session_expired'],
  [503, 'dependency_unavailable'],
] as const)('maps heartbeat status %i without collapsing recovery events', async (status, code) => {
  let requestBody: unknown;
  server.use(
    http.post(`${apiOrigin()}/api/v1/playback/heartbeat`, async ({ request }) => {
      requestBody = await request.json();
      return code === null
        ? new HttpResponse(null, { status })
        : HttpResponse.json({ error: { code, message: code, requestId: 'test' } }, { status });
    }),
  );
  const heartbeat = {
    playbackId,
    lessonId,
    positionSeconds: 19,
    clientSeq: 5,
    reason: 'tick' as const,
  };
  const result = await createApiPlaybackRepository().heartbeat(heartbeat);
  expect(result).toEqual(
    code === null ? { ok: true, value: undefined } : { ok: false, error: { code } },
  );
  expect(requestBody).toEqual(heartbeat);
});
