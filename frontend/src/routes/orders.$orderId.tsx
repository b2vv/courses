import { useEffect } from 'react';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { createFileRoute, notFound, useNavigate } from '@tanstack/react-router';
import { learningCourseQueryKey } from '../app/composition/learn-queries';
import { PurchaseQueryError } from '../app/composition/purchase-queries';
import { requireIdentity, useRedirectWhenSessionEnds } from '../app/composition/require-identity';
import { parseOrderId, parseOrderSearch } from '../app/router/learn-route';
import type { PurchaseErrorCode } from '../domain/purchase/purchase';
import { ProtectedRouteFailure, ProtectedRoutePending } from '../modules/identity';
import { OrderStatusView, submitCheckoutForm } from '../presentation/learn/OrderStatusView';

export const Route = createFileRoute('/orders/$orderId')({
  ssr: false,
  validateSearch: parseOrderSearch,
  beforeLoad: async ({ context, location, params }) => {
    await requireIdentity(context, location);
    if (parseOrderId(params.orderId) === null) throw notFound();
  },
  component: OrderRoute,
  pendingComponent: ProtectedRoutePending,
  errorComponent: ProtectedRouteFailure,
});

function errorCode(error: unknown): PurchaseErrorCode | undefined {
  if (error === null || error === undefined) return undefined;
  return error instanceof PurchaseQueryError ? error.code : 'dependency_unavailable';
}

function OrderRoute() {
  const { learn, identityDependencies } = Route.useRouteContext();
  useRedirectWhenSessionEnds(identityDependencies);
  const { orderId } = Route.useParams();
  const { course: slug } = Route.useSearch();
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const order = useQuery(learn.purchase.status(orderId));
  const retry = useMutation(learn.purchase.start(queryClient));
  const checkout = useMutation(learn.purchase.checkout());
  const paid = order.data?.status === 'paid';

  // A paid order changes the course outline: its lessons must unlock without a reload.
  useEffect(() => {
    if (!paid || slug === undefined) return;
    void queryClient.invalidateQueries({ queryKey: learningCourseQueryKey(slug), exact: true });
  }, [paid, queryClient, slug]);

  const courseHref = slug === undefined ? '/courses' : `/learn/${slug}`;
  const loadError = errorCode(order.error);
  const retryError = errorCode(retry.error);
  const checkoutError = errorCode(checkout.error);
  return (
    <OrderStatusView
      order={order.data}
      {...(loadError === undefined ? {} : { error: loadError })}
      courseHref={courseHref}
      isRetrying={retry.isPending}
      {...(retryError === undefined ? {} : { retryError })}
      onRetry={() => {
        if (order.data === undefined) return;
        retry.mutate(
          { courseId: order.data.courseId },
          {
            onSuccess: (next) =>
              void navigate({
                to: '/orders/$orderId',
                params: { orderId: next.id },
                search: slug === undefined ? {} : { course: slug },
              }),
          },
        );
      }}
      isCheckingOut={checkout.isPending}
      {...(checkoutError === undefined ? {} : { checkoutError })}
      onCheckout={() => {
        if (order.data === undefined || order.data.status !== 'pending') return;
        checkout.mutate(
          { orderId: order.data.id, courseSlug: slug },
          {
            onSuccess: (form) => submitCheckoutForm(form),
            onError: (error) => {
              if (error.code !== 'checkout_already_opened') return;
              retry.mutate(
                { courseId: order.data.courseId },
                {
                  onSuccess: (next) => {
                    checkout.mutate(
                      { orderId: next.id, courseSlug: slug },
                      {
                        onSuccess: (form) => {
                          void navigate({
                            to: '/orders/$orderId',
                            params: { orderId: next.id },
                            search: slug === undefined ? {} : { course: slug },
                          });
                          submitCheckoutForm(form);
                        },
                      },
                    );
                  },
                  onError: (error) => {
                    if (
                      error instanceof PurchaseQueryError &&
                      error.code === 'course_already_owned'
                    ) {
                      void queryClient.invalidateQueries({
                        queryKey: learn.purchase.status(order.data.id).queryKey,
                        exact: true,
                      });
                    }
                  },
                },
              );
            },
          },
        );
      }}
    />
  );
}
