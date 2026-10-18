import { useMemo } from 'react';
import { useInfiniteQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { createFileRoute } from '@tanstack/react-router';
import type { ParticipantReportStateFilter } from '../modules/enrollment';
import { ParticipantReportQueue } from '../modules/admin';

// 'open' is the default view, so it is never written to the URL; only an explicit
// 'resolved' or 'all' narrows/widens it. An unknown value falls back to the default.
type ReportSearch = { state?: Exclude<ParticipantReportStateFilter, 'open'> };

function parseReportSearch(search: Record<string, unknown>): ReportSearch {
  return search.state === 'resolved' || search.state === 'all' ? { state: search.state } : {};
}

export const Route = createFileRoute('/admin/participant-reports')({
  ssr: false,
  validateSearch: parseReportSearch,
  loaderDeps: ({ search }) =>
    ({ state: search.state ?? 'open' }) as { state: ParticipantReportStateFilter },
  loader: ({ context, deps }) =>
    context.queryClient.prefetchInfiniteQuery(context.enrollment.participantReports(deps.state)),
  component: ParticipantReportsRoute,
});

function ParticipantReportsRoute() {
  const { state = 'open' } = Route.useSearch();
  const { enrollment } = Route.useRouteContext();
  const navigate = Route.useNavigate();
  const queryClient = useQueryClient();
  const reports = useInfiniteQuery(enrollment.participantReports(state));
  const resolve = useMutation(enrollment.resolveReport(queryClient));
  const revoke = useMutation(enrollment.revoke(queryClient));
  const items = useMemo(
    () => reports.data?.pages.flatMap((page) => page.items) ?? [],
    [reports.data],
  );
  return (
    <ParticipantReportQueue
      items={items}
      state={state}
      isLoading={reports.isLoading}
      error={reports.error}
      hasNextPage={reports.hasNextPage}
      isFetchingNextPage={reports.isFetchingNextPage}
      resolvePending={resolve.isPending}
      revokePending={revoke.isPending}
      onFilterChange={(next) => void navigate({ search: next === 'open' ? {} : { state: next } })}
      onRetry={() => void reports.refetch()}
      onLoadMore={() => void reports.fetchNextPage()}
      onResolve={(reportId) => resolve.mutateAsync(reportId)}
      onRevoke={(courseId, userId) => revoke.mutateAsync({ courseId, userId })}
    />
  );
}
