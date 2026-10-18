import { describe, expect, test } from '@rstest/core';
import { snapshotSchema, timeseriesSchema } from './admin-stats-schema';

const snapshot = {
  users: {
    active: 5,
    blocked: 1,
    deleted: 2,
    byRole: { student: 3, instructor: 2, admin: 1 },
    verified: 4,
    twoFactor: 1,
  },
  courses: { draft: 1, published: 2, archived: 1 },
  grants: { active: 3, blocked: 1, revoked: 1 },
  orders: { pending: 1, paid: 2, failed: 1, revenueMinorUAH: 12000 },
  reports: { open: 1, resolved: 2 },
  media: {
    inProgress: 1,
    ready: 2,
    failed: 1,
    totalDurationSeconds: 247,
    avgDurationSeconds: 123.5,
    withDurationCount: 2,
  },
  engagement: { activeLearners30d: 4, avgProgress: 0.42, completions: 2 },
  instructors: [
    {
      instructorId: '10000000-0000-4000-8000-000000000009',
      displayName: 'Ada',
      publishedCourses: 2,
      activeEnrollments: 3,
      revenueMinorUAH: 12000,
    },
  ],
  platform: { singleInstructor: true },
};

describe('admin stats snapshot schema', () => {
  test('accepts a full snapshot from the contract', () => {
    const parsed = snapshotSchema.safeParse(snapshot);
    expect(parsed.success).toBe(true);
  });

  test('accepts null averages and an empty leaderboard (nothing to average)', () => {
    const parsed = snapshotSchema.safeParse({
      ...snapshot,
      media: { ...snapshot.media, avgDurationSeconds: null },
      engagement: { ...snapshot.engagement, avgProgress: null },
      instructors: [],
    });
    expect(parsed.success).toBe(true);
  });

  test.each([
    ['an unknown member', { ...snapshot, extra: 1 }],
    ['a negative count', { ...snapshot, courses: { draft: -1, published: 2, archived: 1 } }],
    [
      'a progress fraction above one',
      { ...snapshot, engagement: { ...snapshot.engagement, avgProgress: 1.5 } },
    ],
    [
      'a missing engagement field',
      { ...snapshot, engagement: { activeLearners30d: 4, completions: 2 } },
    ],
    [
      'an instructor missing revenue',
      {
        ...snapshot,
        instructors: [
          {
            instructorId: '10000000-0000-4000-8000-000000000009',
            displayName: 'Ada',
            publishedCourses: 2,
            activeEnrollments: 3,
          },
        ],
      },
    ],
  ])('refuses %s', (_name, malformed) => {
    expect(snapshotSchema.safeParse(malformed).success).toBe(false);
  });
});

describe('admin stats timeseries schema', () => {
  test('accepts an ordered series and the empty series', () => {
    expect(
      timeseriesSchema.safeParse([
        { bucketStart: '2026-03-01T00:00:00Z', value: 4 },
        { bucketStart: '2026-03-02T00:00:00Z', value: 7 },
      ]).success,
    ).toBe(true);
    expect(timeseriesSchema.safeParse([]).success).toBe(true);
  });

  test.each([
    ['a non-datetime bucket', [{ bucketStart: 'banana', value: 1 }]],
    ['an unknown member', [{ bucketStart: '2026-03-01T00:00:00Z', value: 1, extra: 1 }]],
    ['a missing value', [{ bucketStart: '2026-03-01T00:00:00Z' }]],
  ])('refuses %s', (_name, malformed) => {
    expect(timeseriesSchema.safeParse(malformed).success).toBe(false);
  });
});
