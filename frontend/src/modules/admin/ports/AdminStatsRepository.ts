import type { Snapshot, StatsResult, TimeseriesPoint, TimeseriesRange } from '../domain/stats';

export interface AdminStatsRepository {
  snapshot(signal?: AbortSignal): Promise<StatsResult<Snapshot>>;
  timeseries(
    range: TimeseriesRange,
    signal?: AbortSignal,
  ): Promise<StatsResult<readonly TimeseriesPoint[]>>;
}
