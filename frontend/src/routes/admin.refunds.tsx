import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { createFileRoute } from '@tanstack/react-router';
import type { RefundStateFilter } from '../modules/enrollment';
import { RefundQueue } from '../presentation/admin/RefundQueue';

// 'pending' is the default view, so it is never written to the URL; only an explicit
// 'refunded', 'declined' or 'all' narrows/widens it. An unknown value falls back to the default.
type RefundSearch = { state?: Exclude<RefundStateFilter, 'pending'> };

function parseRefundSearch(search: Record<string, unknown>): RefundSearch {
  return search.state === 'refunded' || search.state === 'declined' || search.state === 'all'
    ? { state: search.state }
    : {};
}

export const Route = createFileRoute('/admin/refunds')({
  ssr: false,
  validateSearch: parseRefundSearch,
  loaderDeps: ({ search }) =>
    ({ state: search.state ?? 'pending' }) as { state: RefundStateFilter },
  loader: ({ context, deps }) =>
    context.queryClient.prefetchQuery(context.enrollment.refundQueue(deps.state)),
  component: RefundsRoute,
});

function RefundsRoute() {
  const { state = 'pending' } = Route.useSearch();
  const { enrollment } = Route.useRouteContext();
  const navigate = Route.useNavigate();
  const queryClient = useQueryClient();
  const refunds = useQuery(enrollment.refundQueue(state));
  const refund = useMutation(enrollment.refundOrder(queryClient));
  const decline = useMutation(enrollment.declineRefund(queryClient));
  const revoke = useMutation(enrollment.revokeRefund(queryClient));
  return (
    <RefundQueue
      items={refunds.data?.items ?? []}
      state={state}
      isLoading={refunds.isLoading}
      error={refunds.error}
      refundPending={refund.isPending}
      declinePending={decline.isPending}
      revokePending={revoke.isPending}
      onFilterChange={(next) =>
        void navigate({ search: next === 'pending' ? {} : { state: next } })
      }
      onRetry={() => void refunds.refetch()}
      onRefund={(orderId) => refund.mutateAsync(orderId)}
      onDecline={(orderId) => decline.mutateAsync(orderId)}
      onRevoke={(orderId) => revoke.mutateAsync(orderId)}
    />
  );
}
