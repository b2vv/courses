import { expect, test } from '@rstest/core';
import { QueryClient } from '@tanstack/react-query';
import type { PurchaseOrder, PurchaseResult } from '../../domain/purchase/purchase';
import type { CheckoutForm } from '../../domain/purchase/purchase';
import type { PurchaseRepository } from '../../ports/PurchaseRepository';
import { createPurchaseQueries, purchaseQueryKey } from './purchase-queries';

const pending: PurchaseOrder = {
  id: '60000000-0000-4000-8000-000000000001',
  courseId: '10000000-0000-4000-8000-000000000001',
  amount: { minor: 189000, currency: 'UAH' },
  status: 'pending',
  createdAt: '2026-10-01T10:00:00Z',
  paidAt: null,
};

function repository(keys: string[], answer: PurchaseResult<PurchaseOrder>): PurchaseRepository {
  return {
    start: async (_courseId, key) => {
      keys.push(key);
      return answer;
    },
    status: async () => answer,
    checkout: async () => ({
      ok: true,
      value: { action: 'https://secure.wayforpay.com/pay', fields: [] } satisfies CheckoutForm,
    }),
  };
}

test('checkout unwraps the form and keeps the started order id', async () => {
  const calls: string[] = [];
  const queries = createPurchaseQueries({
    ...repository([], { ok: true, value: pending }),
    checkout: async (orderId) => {
      calls.push(orderId);
      return { ok: true, value: { action: 'https://secure.wayforpay.com/pay', fields: [] } };
    },
  });
  const client = new QueryClient();
  const start = queries.start(client).mutationFn;
  const checkout = queries.checkout().mutationFn;
  if (start === undefined || checkout === undefined) throw new Error('mutation function missing');
  const started = await start({ courseId: pending.courseId }, {} as never);
  const form = await checkout({ orderId: started.id, courseSlug: 'go-basics' }, {} as never);
  expect(calls).toEqual([pending.id]);
  expect(form.action).toBe('https://secure.wayforpay.com/pay');
});

test('every start is a new attempt with a new idempotency key', async () => {
  const keys: string[] = [];
  let next = 0;
  const queries = createPurchaseQueries(
    repository(keys, { ok: true, value: pending }),
    () => `key-${++next}`,
  );
  const client = new QueryClient();
  const start = queries.start(client).mutationFn;
  if (start === undefined) throw new Error('mutation function missing');

  await start({ courseId: pending.courseId }, {} as never);
  await start({ courseId: pending.courseId }, {} as never);

  // A retry after a failed payment must not replay the failed order under the old key.
  expect(keys).toEqual(['key-1', 'key-2']);
});

test('a started order seeds its status query', async () => {
  const queries = createPurchaseQueries(repository([], { ok: true, value: pending }), () => 'k');
  const client = new QueryClient();
  const options = queries.start(client);
  await options.onSuccess?.(pending, { courseId: pending.courseId }, undefined, {} as never);
  expect(client.getQueryData(purchaseQueryKey(pending.id))).toEqual(pending);
});

test('polls while pending and stops once the order is settled', () => {
  const queries = createPurchaseQueries(repository([], { ok: true, value: pending }), () => 'k');
  const interval = queries.status(pending.id).refetchInterval;
  if (typeof interval !== 'function') throw new Error('refetchInterval must be a function');
  const at = (order: PurchaseOrder | undefined) =>
    interval({ state: { data: order } } as Parameters<typeof interval>[0]);

  expect(at(undefined)).toBeGreaterThan(0);
  expect(at(pending)).toBeGreaterThan(0);
  expect(at({ ...pending, status: 'paid', paidAt: '2026-10-01T10:01:00Z' })).toBe(false);
  expect(at({ ...pending, status: 'failed' })).toBe(false);
});
