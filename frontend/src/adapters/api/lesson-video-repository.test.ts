import { afterAll, afterEach, beforeAll, describe, expect, test } from '@rstest/core';
import { http, HttpResponse } from 'msw';
import { setupServer } from 'msw/node';
import { createLessonVideos } from '../../application/admin/lesson-video';
import type { VideoErrorCode } from '../../domain/media/video';
import { apiOrigin } from './api-origin';
import { createApiLessonVideoRepository } from './lesson-video-repository';

const courseId = '10000000-0000-4000-8000-000000000001';
const lessonId = '30000000-0000-4000-8000-000000000001';
const videoId = '50000000-0000-4000-8000-000000000001';
const api = `${apiOrigin()}/api/v1/admin`;
const envelope = (code: string) => ({
  error: { code, message: 'x', requestId: 'a'.repeat(32) },
});
const server = setupServer();
beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

type Failure = {
  readonly name: string;
  readonly expected: VideoErrorCode;
  invoke(): Promise<unknown>;
};

describe('lesson video API repository', () => {
  test('maps R23 JSON business codes through matching use cases', async () => {
    const videos = createLessonVideos(createApiLessonVideoRepository());
    const failures: Failure[] = [
      {
        name: 'mint conflict',
        expected: 'upload_in_progress',
        invoke: () => {
          server.use(
            http.post(`${api}/lessons/${lessonId}/video`, () =>
              HttpResponse.json(envelope('upload_in_progress'), { status: 409 }),
            ),
          );
          return videos.mint(lessonId, 3, 'lesson.mp4');
        },
      },
      {
        name: 'ownership refusal',
        expected: 'course_forbidden',
        invoke: () => {
          server.use(
            http.get(`${api}/courses/${courseId}/videos`, () =>
              HttpResponse.json(envelope('course_forbidden'), { status: 403 }),
            ),
          );
          return videos.listByCourse(courseId);
        },
      },
      {
        name: 'opaque server failure',
        expected: 'dependency_unavailable',
        invoke: () => {
          server.use(
            http.get(`${api}/courses/${courseId}/videos`, () =>
              HttpResponse.json(envelope('internal_error'), { status: 500 }),
            ),
          );
          return videos.listByCourse(courseId);
        },
      },
    ];

    for (const failure of failures) {
      const result = await failure.invoke();
      expect(result, failure.name).toEqual({ ok: false, error: { code: failure.expected } });
      server.resetHandlers();
    }
  });

  test('parses generated ticket and list responses into domain values', async () => {
    const video = {
      lessonId,
      videoId,
      state: 'created',
      generation: 1,
      originalName: 'lesson.mp4',
      receivedBytes: 0,
      declaredBytes: 3,
      updatedAt: '2026-09-06T12:00:00.123456Z',
    };
    server.use(
      http.post(`${api}/lessons/${lessonId}/video`, () =>
        HttpResponse.json(
          {
            videoId,
            uploadUrl: `/uploads/${videoId}`,
            chunkBytes: 8 * 1024 * 1024,
            maxBytes: 2 * 1024 * 1024 * 1024,
            expiresAt: '2026-09-07T12:00:00.123456Z',
          },
          { status: 201 },
        ),
      ),
      http.get(`${api}/courses/${courseId}/videos`, () => HttpResponse.json({ items: [video] })),
    );
    const videos = createLessonVideos(createApiLessonVideoRepository());
    expect(await videos.mint(lessonId, 3, 'lesson.mp4')).toMatchObject({
      ok: true,
      value: { videoId },
    });
    expect(await videos.listByCourse(courseId)).toEqual({ ok: true, value: [video] });
  });
});
