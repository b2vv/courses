import { expect, test } from '@rstest/core';
import { parseLessonRoute } from './learn-route';

const lessonId = '30000000-0000-4000-8000-000000000001';

test('parses exactly a course slug and lesson UUID from the protected learn splat', () => {
  expect(parseLessonRoute(`go-basics/${lessonId}`)).toEqual({ slug: 'go-basics', lessonId });
});

test.each([
  undefined,
  '',
  'go-basics',
  `go-basics/${lessonId}/extra`,
  `Go-Basics/${lessonId}`,
  'go-basics/not-a-uuid',
])('rejects a non-lesson learn path: %s', (path) => {
  expect(parseLessonRoute(path)).toBeNull();
});
