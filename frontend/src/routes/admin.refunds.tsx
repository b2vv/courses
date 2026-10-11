import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { createFileRoute } from '@tanstack/react-router';
import { RefundQueue } from '../presentation/admin/RefundQueue';

export const Route = createFileRoute('/admin/refunds')({
  ssr: false,
  loader: ({ context }) => context.queryClient.prefetchQuery(context.enrollment.refundQueue()),
  component: RefundsRoute,
});

function RefundsRoute() {
  const { enrollment } = Route.useRouteContext();
  const queryClient = useQueryClient();
  const refunds = useQuery(enrollment.refundQueue());
  const resolve = useMutation(enrollment.resolveRefund(queryClient));
  const revoke = useMutation(enrollment.revokeRefund(queryClient));
  return (
    <RefundQueue
      items={refunds.data?.items ?? []}
      isLoading={refunds.isLoading}
      error={refunds.error}
      resolvePending={resolve.isPending}
      revokePending={revoke.isPending}
      onRetry={() => void refunds.refetch()}
      onResolve={(orderId) => resolve.mutateAsync(orderId)}
      onRevoke={(orderId) => revoke.mutateAsync(orderId)}
    />
  );
}
