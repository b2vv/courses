import { afterAll, afterEach, beforeAll, describe, expect, test } from '@rstest/core';
import { http, HttpResponse } from 'msw';
import { setupServer } from 'msw/node';
import type { StatsErrorCode } from '../../domain/admin/stats';
import { apiOrigin } from './api-origin';
import { createApiAdminStatsRepository } from './admin-stats-repository';

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
    avgDurationSeconds: null,
    withDurationCount: 0,
  },
  engagement: { activeLearners30d: 4, avgProgress: 0.42, completions: 2 },
  instructors: [],
  platform: { singleInstructor: true },
};

const envelope = (code: string) => ({ error: { code, message: 'x', requestId: 'a'.repeat(32) } });
const statsUrl = `${apiOrigin()}/api/v1/admin/stats`;
const timeseriesUrl = `${statsUrl}/timeseries`;

const server = setupServer();
beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('admin stats repository success', () => {
  test('parses the snapshot', async () => {
    server.use(http.get(statsUrl, () => HttpResponse.json(snapshot)));
    const result = await createApiAdminStatsRepository().snapshot();
    expect(result.ok).toBe(true);
    if (result.ok) expect(result.value.platform.singleInstructor).toBe(true);
  });

  test('parses the timeseries', async () => {
    server.use(
      http.get(timeseriesUrl, () =>
        HttpResponse.json([{ bucketStart: '2026-03-01T00:00:00Z', value: 4 }]),
      ),
    );
    const result = await createApiAdminStatsRepository().timeseries({
      metric: 'signups',
      bucket: 'day',
    });
    expect(result.ok).toBe(true);
    if (result.ok) expect(result.value).toHaveLength(1);
  });

  test('rejects a malformed snapshot body as invalid_response', async () => {
    server.use(http.get(statsUrl, () => HttpResponse.json({ users: {} })));
    const result = await createApiAdminStatsRepository().snapshot();
    expect(result.ok).toBe(false);
    if (!result.ok) expect(result.error.code).toBe('invalid_response');
  });
});

describe('admin stats repository error mapping', () => {
  const snapshotFailures: readonly { status: number; code: string; expected: StatsErrorCode }[] = [
    { status: 401, code: 'authentication_required', expected: 'authentication_required' },
    { status: 403, code: 'grant_forbidden', expected: 'forbidden' },
    { status: 503, code: 'dependency_unavailable', expected: 'dependency_unavailable' },
    { status: 500, code: 'internal_error', expected: 'dependency_unavailable' },
    { status: 404, code: 'not_found', expected: 'invalid_response' },
  ];
  test.each(snapshotFailures)('snapshot $status $code -> $expected', async (failure) => {
    server.use(
      http.get(statsUrl, () =>
        HttpResponse.json(envelope(failure.code), { status: failure.status }),
      ),
    );
    const result = await createApiAdminStatsRepository().snapshot();
    expect(result.ok).toBe(false);
    if (!result.ok) expect(result.error.code).toBe(failure.expected);
  });

  const timeseriesFailures: readonly { code: string; expected: StatsErrorCode }[] = [
    { code: 'invalid_range', expected: 'invalid_range' },
    { code: 'range_too_wide', expected: 'range_too_wide' },
    { code: 'unknown_metric', expected: 'unknown_metric' },
    { code: 'unknown_bucket', expected: 'unknown_bucket' },
  ];
  test.each(timeseriesFailures)('timeseries 400 $code -> $expected', async (failure) => {
    server.use(
      http.get(timeseriesUrl, () => HttpResponse.json(envelope(failure.code), { status: 400 })),
    );
    const result = await createApiAdminStatsRepository().timeseries({
      metric: 'signups',
      bucket: 'day',
    });
    expect(result.ok).toBe(false);
    if (!result.ok) expect(result.error.code).toBe(failure.expected);
  });
});
