import { afterAll, afterEach, beforeAll, expect, test } from '@rstest/core';
import { QueryClient } from '@tanstack/react-query';
import { http, HttpResponse } from 'msw';
import { setupServer } from 'msw/node';
import { primeAv1Probe, takeAv1ProbeResult } from '../../app/composition/av1-probe';
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
  price: null,
  paid: false,
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
const originalWindow = Object.getOwnPropertyDescriptor(globalThis, 'window');
const originalMediaSource = Object.getOwnPropertyDescriptor(globalThis, 'MediaSource');

function restoreGlobal(name: 'window' | 'MediaSource', descriptor: PropertyDescriptor | undefined) {
  if (descriptor === undefined) {
    Reflect.deleteProperty(globalThis, name);
    return;
  }
  Object.defineProperty(globalThis, name, descriptor);
}

beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
afterEach(() => {
  server.resetHandlers();
  restoreGlobal('window', originalWindow);
  restoreGlobal('MediaSource', originalMediaSource);
});
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

  // A paid course the viewer is entitled to carries its price; the learning
  // model does not need it, so it is accepted at the boundary and not mapped.
  server.use(
    http.get(`${apiOrigin()}/api/v1/learn/courses/go-basics`, () =>
      HttpResponse.json({ ...course, price: { minor: 49900, currency: 'UAH' }, paid: true }),
    ),
  );
  const paid = await createApiPlaybackRepository().getCourse('go-basics');
  expect(paid.ok).toBe(true);

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

test('P12 carries the real page-scope probe on a denied mint without repeating it on retry', async () => {
  const requestBodies: unknown[] = [];
  let mintAttempts = 0;
  let probeCalls = 0;
  Object.defineProperty(globalThis, 'window', {
    configurable: true,
    value: { location: { origin: 'http://api.test' } },
  });
  Object.defineProperty(globalThis, 'MediaSource', {
    configurable: true,
    value: {
      isTypeSupported: () => {
        probeCalls += 1;
        return true;
      },
    },
  });
  server.use(
    http.post(`${apiOrigin()}/api/v1/playback/sessions`, async ({ request }) => {
      requestBodies.push(await request.json());
      mintAttempts += 1;
      if (mintAttempts === 1) {
        return HttpResponse.json(
          {
            error: {
              code: 'playback_forbidden',
              message: 'Playback is not authorized',
              requestId: 'test',
            },
          },
          { status: 403 },
        );
      }
      return HttpResponse.json(session, { status: 201 });
    }),
  );
  const queryClient = new QueryClient();
  primeAv1Probe(queryClient);
  const repository = createApiPlaybackRepository({}, () => takeAv1ProbeResult(queryClient));

  expect(await repository.play(lessonId)).toEqual({
    ok: false,
    error: { code: 'playback_forbidden' },
  });
  expect(await repository.play(lessonId)).toEqual({ ok: true, value: session });

  expect(probeCalls).toBe(1);
  expect(requestBodies, `P12 mint request bodies: ${JSON.stringify(requestBodies)}`).toEqual([
    { lessonId, av1Support: 'supported' },
    { lessonId },
  ]);
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
