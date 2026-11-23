import { MutationObserver, QueryClient } from '@tanstack/react-query';
import { expect, test } from '@rstest/core';
import type { LearningCourse, PlaybackResult } from '../../modules/playback';
import type { CourseRepository } from '../../modules/catalog';
import type { PlaybackRepository } from '../../modules/playback';
import {
  createContinueReads,
  createLearnQueries,
  learningCourseQueryKey,
  refreshAfterHeartbeat,
} from './learn-queries';

// This test reads only the learning course; the review side of the course repository is unused here.
const unusedCourses: CourseRepository = {
  listPublished: async () => ({ ok: false, error: { code: 'dependency_unavailable' } }),
  getPublished: async () => ({ ok: false, error: { code: 'dependency_unavailable' } }),
  listReviews: async () => ({ ok: false, error: { code: 'dependency_unavailable' } }),
  reportReview: async () => ({ ok: false, error: { code: 'dependency_unavailable' } }),
  ownReview: async () => ({ ok: false, error: { code: 'dependency_unavailable' } }),
  putReview: async () => ({ ok: false, error: { code: 'dependency_unavailable' } }),
  deleteReview: async () => ({ ok: false, error: { code: 'dependency_unavailable' } }),
  listLearningReviews: async () => ({ ok: false, error: { code: 'dependency_unavailable' } }),
};

const course: LearningCourse = {
  id: '10000000-0000-4000-8000-000000000001',
  slug: 'go-basics',
  title: 'Go basics',
  summary: 'Practice',
  description: 'Learn Go.',
  status: 'published',
  price: null,
  continueLessonId: null,
  modules: [],
};

function repository(read: () => Promise<PlaybackResult<LearningCourse>>): PlaybackRepository {
  return {
    getCourse: read,
    play: async () => ({ ok: false, error: { code: 'lesson_not_found' } }),
    heartbeat: async () => ({ ok: false, error: { code: 'playback_session_expired' } }),
  };
}

test('uses one learning-course key and lets Query deduplicate concurrent route reads', async () => {
  let reads = 0;
  const learn = createLearnQueries(
    repository(async () => {
      reads += 1;
      return { ok: true, value: course };
    }),
    unusedCourses,
  );
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const query = learn.course(course.slug);

  expect(query.queryKey).toEqual(learningCourseQueryKey(course.slug));
  const [first, second] = await Promise.all([
    client.fetchQuery(query),
    client.fetchQuery(learn.course(course.slug)),
  ]);
  expect(first).toBe(course);
  expect(second).toBe(course);
  expect(reads).toBe(1);
});

test('a review write or delete refreshes both the own review and the learning list', async () => {
  const client = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const invalidated: unknown[] = [];
  const original = client.invalidateQueries.bind(client);
  client.invalidateQueries = (filters, options) => {
    invalidated.push(filters?.queryKey);
    return original(filters, options);
  };
  const ok = async () => ({ ok: true as const, value: undefined });
  const learn = createLearnQueries(
    repository(async () => ({ ok: true, value: course })),
    { ...unusedCourses, deleteReview: ok },
  );
  await new MutationObserver(client, learn.deleteReview(client)).mutate(
    '10000000-0000-4000-8000-000000000001',
  );
  expect(invalidated).toEqual([
    ['learn', 'review', '10000000-0000-4000-8000-000000000001'],
    ['learn', 'reviews', '10000000-0000-4000-8000-000000000001'],
    ['learn', 'course'],
    ['catalog'],
    ['course'],
  ]);
});

test.each([
  ['nav', true],
  ['ended', true],
  ['tick', false],
  ['pause', false],
  ['seek', false],
  ['hide', false],
] as const)(
  'C3b a %s heartbeat refreshes the course and review eligibility: %s',
  async (reason, refresh) => {
    const client = new QueryClient();
    client.setQueryData(learningCourseQueryKey('go'), { slug: 'go' });
    client.setQueryData(['learn', 'review', 'course-1', 'user-1'], {
      eligibility: 'review_not_started',
    });
    client.setQueryData(learningCourseQueryKey('other'), { slug: 'other' });
    await refreshAfterHeartbeat(client, { slug: 'go', courseId: 'course-1', reason });
    expect(client.getQueryState(learningCourseQueryKey('go'))?.isInvalidated).toBe(refresh);
    expect(client.getQueryState(['learn', 'review', 'course-1', 'user-1'])?.isInvalidated).toBe(
      refresh,
    );
    // Another course is not touched.
    expect(client.getQueryState(learningCourseQueryKey('other'))?.isInvalidated).toBe(false);
  },
);

test('C3b a continue-redirect token is taken once', () => {
  const tokens = createContinueReads();
  const token = tokens.issue();
  expect(tokens.take(undefined)).toBe(false);
  expect(tokens.take('not-issued')).toBe(false);
  expect(tokens.take(token)).toBe(true);
  // A Back to the same history entry brings the same token again: it no longer counts.
  expect(tokens.take(token)).toBe(false);
  // Tokens whose lesson never loaded do not pile up.
  const first = tokens.issue();
  for (let index = 0; index < 8; index += 1) tokens.issue();
  expect(tokens.take(first)).toBe(false);
});

test('C3b clear retires every untaken continue token', () => {
  const tokens = createContinueReads();
  const token = tokens.issue();
  tokens.clear();
  expect(tokens.take(token)).toBe(false);
});
