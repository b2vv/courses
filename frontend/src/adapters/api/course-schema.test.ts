import { expect, test } from '@rstest/core';
import { courseSchema, publicReviewPageSchema } from './course-schema';

const course = {
  id: '10000000-0000-4000-8000-000000000001',
  slug: 'go',
  title: 'Go',
  summary: 'Learn Go',
  description: 'Learn Go well',
  status: 'published',
  createdAt: '2026-01-01T00:00:00Z',
  updatedAt: '2026-01-01T00:00:00Z',
  publishedAt: '2026-01-01T00:00:00Z',
  price: null,
  paid: false,
  rating: { average: 4.7, count: 3 },
  lessonCount: 1,
};

test('validates catalog rating and lesson count', () => {
  expect(courseSchema.safeParse(course).success).toBe(true);
  expect(courseSchema.safeParse({ ...course, rating: { average: 6, count: 3 } }).success).toBe(
    false,
  );
  expect(courseSchema.safeParse({ ...course, rating: { average: 4, count: 2 } }).success).toBe(
    false,
  );
  expect(courseSchema.safeParse({ ...course, lessonCount: -1 }).success).toBe(false);
});

test('validates public review items', () => {
  expect(
    publicReviewPageSchema.safeParse({
      rating: null,
      nextCursor: null,
      items: [
        {
          id: '20000000-0000-4000-8000-000000000001',
          authorDisplayName: 'Ada',
          stars: 4,
          body: 'Helpful course review',
          completedLessons: 1,
          totalLessons: 3,
          createdAt: '2026-01-01T00:00:00Z',
          editedAt: null,
        },
      ],
    }).success,
  ).toBe(true);
});
