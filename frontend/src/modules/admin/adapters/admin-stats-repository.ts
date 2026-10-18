import type {
  Snapshot,
  StatsError,
  StatsResult,
  TimeseriesPoint,
  TimeseriesRange,
} from '../domain/stats';
import type { AdminStatsRepository } from '../ports/AdminStatsRepository';
import { getAdminStats, getAdminStatsTimeseries } from '../../../adapters/api/generated/courses';
import { errorSchema } from '../../../adapters/api/course-schema';
import { snapshotSchema, timeseriesSchema } from './admin-stats-schema';
import { createApiTransport, unnamedStatusCode } from '../../../adapters/api/transport';
import type { TransportOptions, WireResponse } from '../../../adapters/api/transport';

const invalid = { ok: false, error: { code: 'invalid_response' } } satisfies StatsResult<never>;

function wireError(response: WireResponse): StatsError {
  const parsed = errorSchema.safeParse(response.data);
  if (parsed.success) {
    switch (parsed.data.error.code) {
      case 'authentication_required':
        if (response.status === 401) return { code: 'authentication_required' };
        break;
      case 'grant_forbidden':
      case 'forbidden':
        if (response.status === 403) return { code: 'forbidden' };
        break;
      case 'invalid_range':
      case 'range_too_wide':
      case 'unknown_metric':
      case 'unknown_bucket':
        if (response.status === 400) return { code: parsed.data.error.code };
        break;
      case 'dependency_unavailable':
        if (response.status === 503) return { code: 'dependency_unavailable' };
        break;
      case 'internal_error':
        if (response.status === 500) return { code: 'dependency_unavailable' };
        break;
    }
  }
  return { code: unnamedStatusCode(response.status) };
}

/** One repository per router request; generated transport types never leave this adapter. */
export function createApiAdminStatsRepository(
  options: TransportOptions = {},
): AdminStatsRepository {
  const send = createApiTransport(options);
  async function request<T>(
    operation: string,
    execute: (init: RequestInit) => Promise<WireResponse>,
    map: (data: unknown) => StatsResult<T>,
    caller?: AbortSignal,
  ): Promise<StatsResult<T>> {
    const outcome = await send(operation, execute, 200, caller);
    if (!outcome.ok) return { ok: false, error: { code: outcome.code } };
    if (outcome.status === 200) return map(outcome.data);
    return { ok: false, error: wireError(outcome) };
  }

  return {
    snapshot(signal) {
      return request<Snapshot>(
        'admin.stats.snapshot',
        (init) => getAdminStats(init),
        (raw) => {
          const parsed = snapshotSchema.safeParse(raw);
          if (!parsed.success) return invalid;
          return { ok: true, value: parsed.data };
        },
        signal,
      );
    },
    timeseries(range: TimeseriesRange, signal) {
      const params: Parameters<typeof getAdminStatsTimeseries>[0] = {
        metric: range.metric,
        bucket: range.bucket,
        ...(range.from === undefined ? {} : { from: range.from }),
        ...(range.to === undefined ? {} : { to: range.to }),
      };
      return request<readonly TimeseriesPoint[]>(
        'admin.stats.timeseries',
        (init) => getAdminStatsTimeseries(params, init),
        (raw) => {
          const parsed = timeseriesSchema.safeParse(raw);
          if (!parsed.success) return invalid;
          return { ok: true, value: parsed.data };
        },
        signal,
      );
    },
  };
}
