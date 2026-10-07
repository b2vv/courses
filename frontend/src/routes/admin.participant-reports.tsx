import { useMemo } from 'react';
import { useInfiniteQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { createFileRoute } from '@tanstack/react-router';
import { ParticipantReportQueue } from '../presentation/admin/ParticipantReportQueue';

export const Route = createFileRoute('/admin/participant-reports')({
  ssr: false,
  loader: ({ context }) =>
    context.queryClient.prefetchInfiniteQuery(context.enrollment.participantReports()),
  component: ParticipantReportsRoute,
});

function ParticipantReportsRoute() {
  const { enrollment } = Route.useRouteContext();
  const queryClient = useQueryClient();
  const reports = useInfiniteQuery(enrollment.participantReports());
  const resolve = useMutation(enrollment.resolveReport(queryClient));
  const revoke = useMutation(enrollment.revoke(queryClient));
  const items = useMemo(
    () => reports.data?.pages.flatMap((page) => page.items) ?? [],
    [reports.data],
  );
  return (
    <ParticipantReportQueue
      items={items}
      isLoading={reports.isLoading}
      error={reports.error}
      hasNextPage={reports.hasNextPage}
      isFetchingNextPage={reports.isFetchingNextPage}
      resolvePending={resolve.isPending}
      revokePending={revoke.isPending}
      onRetry={() => void reports.refetch()}
      onLoadMore={() => void reports.fetchNextPage()}
      onResolve={(reportId) => resolve.mutateAsync(reportId)}
      onRevoke={(courseId, userId) => revoke.mutateAsync({ courseId, userId })}
    />
  );
}
