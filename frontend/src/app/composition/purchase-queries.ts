import { mutationOptions, queryOptions } from '@tanstack/react-query';
import type { QueryClient } from '@tanstack/react-query';
import {
  isSettled,
  type PurchaseErrorCode,
  type CheckoutForm,
  type PurchaseOrder,
  type PurchaseResult,
} from '../../modules/purchase';
import type { PurchaseRepository } from '../../modules/purchase';
import {
  enrollmentMembershipQueryKey,
  enrollmentMyMembershipsQueryKey,
} from './enrollment-queries';

/** How often a pending order is re-read while the buyer waits for the provider's answer. */
export const purchasePollMs = 2_000;

export function purchaseQueryKey(orderId: string) {
  return ['learn', 'purchase', orderId] as const;
}

export class PurchaseQueryError extends Error {
  readonly code: PurchaseErrorCode;
  constructor(code: PurchaseErrorCode) {
    super(code);
    this.name = 'PurchaseQueryError';
    this.code = code;
  }
}

function unwrap<T>(result: PurchaseResult<T>): T {
  if (!result.ok) throw new PurchaseQueryError(result.error.code);
  return result.value;
}

/**
 * Every start is a new attempt with a new idempotency key: the backend replays whatever
 * order a key already holds — pending, paid or failed — so a retry after a failed payment
 * must not reuse the old key. Double submission is prevented by the button, not the key.
 */
export function createPurchaseQueries(
  purchase: PurchaseRepository,
  newIdempotencyKey: () => string = () => crypto.randomUUID(),
) {
  return {
    purchase,
    start: (client: QueryClient) =>
      mutationOptions({
        mutationFn: async ({ courseId }: { courseId: string }) =>
          unwrap(await purchase.start(courseId, newIdempotencyKey())),
        onSuccess: (order: PurchaseOrder) => {
          client.setQueryData(purchaseQueryKey(order.id), order);
        },
      }),
    status: (orderId: string) =>
      queryOptions({
        queryKey: purchaseQueryKey(orderId),
        queryFn: async ({ signal, client }) => {
          const order = unwrap(await purchase.status(orderId, signal));
          const before = client.getQueryData<PurchaseOrder>(purchaseQueryKey(orderId));
          // A paid order makes the buyer a member. Moving back to the course inside the app keeps
          // the cache, so the cached "not a member" would offer «Купити курс» for up to a minute
          // (C3b R6). Once per move into `paid`: polling stops there (`refetchInterval`). The
          // order's course key covers every viewer's entry for that course.
          if (order.status === 'paid' && before?.status !== 'paid') {
            await Promise.all([
              client.invalidateQueries({ queryKey: enrollmentMembershipQueryKey(order.courseId) }),
              client.invalidateQueries({ queryKey: enrollmentMyMembershipsQueryKey }),
            ]);
          }
          return order;
        },
        refetchInterval: (query) => {
          const order = query.state.data;
          return order !== undefined && isSettled(order) ? false : purchasePollMs;
        },
      }),
    checkout: () =>
      mutationOptions({
        mutationFn: async ({
          orderId,
          courseSlug,
        }: {
          orderId: string;
          courseSlug?: string | undefined;
        }) => unwrap(await purchase.checkout(orderId, courseSlug)),
      }) as ReturnType<
        typeof mutationOptions<
          CheckoutForm,
          PurchaseQueryError,
          { orderId: string; courseSlug?: string | undefined }
        >
      >,
  };
}
