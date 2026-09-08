import { QueryClient } from '@tanstack/react-query';
import { expect, test } from '@rstest/core';
import type { LearningCourse, PlaybackResult } from '../../domain/playback/playback';
import type { PlaybackRepository } from '../../ports/PlaybackRepository';
import { createLearnQueries, learningCourseQueryKey } from './learn-queries';

const course: LearningCourse = {
  id: '10000000-0000-4000-8000-000000000001',
  slug: 'go-basics',
  title: 'Go basics',
  summary: 'Practice',
  description: 'Learn Go.',
  status: 'published',
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
