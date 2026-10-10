import { queryOptions } from '@tanstack/react-query';
import type { StatsError, StatsResult, TimeseriesRange } from '../../domain/admin/stats';
import type { AdminStatsRepository } from '../../ports/AdminStatsRepository';
import { adminListStaleMs } from './admin-queries';

export const adminStatsKey = ['admin', 'stats'] as const;

export class AdminStatsQueryError extends Error {
  readonly code: StatsError['code'];
  constructor(error: StatsError) {
    super(error.code);
    this.name = 'AdminStatsQueryError';
    this.code = error.code;
  }
}

function unwrap<T>(result: StatsResult<T>): T {
  if (!result.ok) throw new AdminStatsQueryError(result.error);
  return result.value;
}

export function timeseriesQueryKey(range: TimeseriesRange) {
  return [
    ...adminStatsKey,
    'timeseries',
    range.metric,
    range.bucket,
    range.from ?? '',
    range.to ?? '',
  ] as const;
}

export function createAdminStatsQueries(stats: AdminStatsRepository) {
  return {
    snapshot: () =>
      queryOptions({
        queryKey: [...adminStatsKey, 'snapshot'] as const,
        queryFn: async ({ signal }) => unwrap(await stats.snapshot(signal)),
        staleTime: adminListStaleMs,
      }),
    timeseries: (range: TimeseriesRange) =>
      queryOptions({
        queryKey: timeseriesQueryKey(range),
        queryFn: async ({ signal }) => unwrap(await stats.timeseries(range, signal)),
        staleTime: adminListStaleMs,
      }),
    unwrap,
  };
}
