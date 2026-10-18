import { afterAll, afterEach, beforeAll, describe, expect, rstest, test } from '@rstest/core';
import { QueryClient, QueryObserver } from '@tanstack/react-query';
import { http, HttpResponse } from 'msw';
import { setupServer } from 'msw/node';
import { createApiLessonVideoRepository } from '../../modules/media';
import { apiOrigin } from '../../adapters/api/api-origin';
import { createLessonVideos } from '../../modules/media';
import type { LessonVideo } from '../../modules/media';
import type { LessonVideoRepository } from '../../modules/media';
import {
  adminCourseVideosQueryKey,
  createAdminVideoQueries,
  invalidateAfterVideoWrite,
  videoPollMs,
  videoRefetchInterval,
} from './admin-video-queries';
import { lessonDraftQueryKey } from './admin-draft-queries';

const courseId = '10000000-0000-4000-8000-000000000001';
const lessonId = '30000000-0000-4000-8000-000000000001';
const otherLessonId = '30000000-0000-4000-8000-000000000002';
const videoId = '50000000-0000-4000-8000-000000000001';
const updatedAt = '2026-09-06T12:00:00.123456Z';
const server = setupServer();
beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

function video(state: LessonVideo['state']): LessonVideo {
  return {
    lessonId,
    videoId,
    state,
    generation: 1,
    originalName: 'lesson.mp4',
    receivedBytes: 12,
    declaredBytes: 12,
    updatedAt,
  };
}

function videoForLesson(
  lesson: string,
  generation: number,
  state: LessonVideo['state'],
): LessonVideo {
  return { ...video(state), lessonId: lesson, generation };
}

function repository(overrides: Partial<LessonVideoRepository> = {}): LessonVideoRepository {
  const base: LessonVideoRepository = {
    mint: async () => ({ ok: false, error: { code: 'dependency_unavailable' } }),
    listByCourse: async () => ({ ok: true, value: [] }),
    retry: async () => ({ ok: true, value: video('queued') }),
    transfer: async () => ({ ok: true, value: null }),
    cancelUpload: async () => ({ ok: true, value: null }),
  };
  return { ...base, ...overrides };
}

describe('admin video query boundary', () => {
  test('P11 polls processing to ready, then stops; created alone never polls', async () => {
    rstest.useFakeTimers();
    try {
      let reads = 0;
      server.use(
        http.get(`${apiOrigin()}/api/v1/admin/courses/${courseId}/videos`, () => {
          reads += 1;
          return HttpResponse.json({ items: [video(reads === 1 ? 'processing' : 'ready')] });
        }),
      );
      const queries = createAdminVideoQueries(createLessonVideos(createApiLessonVideoRepository()));
      const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
      const observer = new QueryObserver(client, queries.list(courseId));
      const processing = new Promise<void>((resolve) => {
        const unsubscribe = observer.subscribe((result) => {
          if (result.data?.[0]?.state === 'processing') {
            unsubscribe();
            resolve();
          }
        });
      });
      const unsubscribePolling = observer.subscribe(() => {});
      await processing;
      await rstest.advanceTimersByTimeAsync(videoPollMs);
      expect(reads).toBe(2);
      expect(
        client.getQueryData<readonly LessonVideo[]>(adminCourseVideosQueryKey(courseId)),
      ).toEqual([video('ready')]);
      await rstest.advanceTimersByTimeAsync(videoPollMs * 2);
      expect(reads).toBe(2);
      unsubscribePolling();

      reads = 0;
      server.use(
        http.get(`${apiOrigin()}/api/v1/admin/courses/${courseId}/videos`, () => {
          reads += 1;
          return HttpResponse.json({ items: [video('created')] });
        }),
      );
      const createdClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
      const createdObserver = new QueryObserver(createdClient, queries.list(courseId));
      const created = new Promise<void>((resolve) => {
        const unsubscribe = createdObserver.subscribe((result) => {
          if (result.data?.[0]?.state === 'created') {
            unsubscribe();
            resolve();
          }
        });
      });
      const unsubscribeCreated = createdObserver.subscribe(() => {});
      await created;
      await rstest.advanceTimersByTimeAsync(videoPollMs * 2);
      expect(reads).toBe(1);
      unsubscribeCreated();
      expect(videoRefetchInterval([video('created')], true)).toBe(videoPollMs);
    } finally {
      rstest.useRealTimers();
    }
  });

  test('uses a stable course key and invalidates only that course after a write', async () => {
    const client = new QueryClient();
    const key = adminCourseVideosQueryKey(courseId);
    const other = adminCourseVideosQueryKey('10000000-0000-4000-8000-000000000002');
    client.setQueryData(key, [video('created')]);
    client.setQueryData(other, []);
    await invalidateAfterVideoWrite(client, courseId);
    expect(client.getQueryState(key)?.isInvalidated).toBe(true);
    expect(client.getQueryState(other)?.isInvalidated).toBe(false);
  });

  test('invalidates only the lesson draft whose ready video generation changed', async () => {
    let reads = 0;
    const queries = createAdminVideoQueries(
      createLessonVideos(
        repository({
          listByCourse: async () => {
            reads++;
            return {
              ok: true,
              value:
                reads === 1
                  ? [
                      videoForLesson(lessonId, 1, 'ready'),
                      videoForLesson(otherLessonId, 1, 'ready'),
                    ]
                  : reads === 2
                    ? [
                        videoForLesson(lessonId, 2, 'processing'),
                        videoForLesson(otherLessonId, 1, 'ready'),
                      ]
                    : [
                        videoForLesson(lessonId, 2, 'ready'),
                        videoForLesson(otherLessonId, 1, 'ready'),
                      ],
            };
          },
        }),
      ),
    );
    const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
    client.setQueryData(lessonDraftQueryKey(lessonId), null);
    client.setQueryData(lessonDraftQueryKey(otherLessonId), null);
    const observer = new QueryObserver(client, queries.list(courseId));
    await observer.refetch();
    await observer.refetch();
    expect(client.getQueryState(lessonDraftQueryKey(lessonId))?.isInvalidated).toBe(false);
    await observer.refetch();

    expect(client.getQueryState(lessonDraftQueryKey(lessonId))?.isInvalidated).toBe(true);
    expect(client.getQueryState(lessonDraftQueryKey(otherLessonId))?.isInvalidated).toBe(false);
  });

  test('every video mutation exposes the named invalidation boundary', async () => {
    const queries = createAdminVideoQueries(createLessonVideos(repository()));
    const client = new QueryClient();
    expect(queries.mint(client).onSuccess).toBeDefined();
    expect(queries.retry(client).onSuccess).toBeDefined();
    expect(queries.transfer(client).onSettled).toBeDefined();
    expect(queries.cancelUpload(client).onSuccess).toBeDefined();
  });
});
