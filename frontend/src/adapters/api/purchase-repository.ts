import { z } from 'zod';
import type {
  PurchaseErrorCode,
  PurchaseOrder,
  PurchaseResult,
} from '../../domain/purchase/purchase';
import type { PurchaseRepository } from '../../ports/PurchaseRepository';
import { getPurchaseStatus, startCoursePurchase } from './generated/courses';
import { priceSchema } from './price-schema';
import { createApiTransport, unnamedStatusCode } from './transport';
import type { TransportOptions, WireResponse } from './transport';

const orderSchema = z
  .object({
    id: z.uuid(),
    courseId: z.uuid(),
    amount: priceSchema,
    status: z.enum(['pending', 'paid', 'failed']),
    createdAt: z.iso.datetime({ offset: true }),
    paidAt: z.iso.datetime({ offset: true }).nullable(),
  })
  .strict();

const errorSchema = z.object({ error: z.object({ code: z.string() }).loose() }).loose();

const startErrors = {
  400: ['invalid_request'],
  401: ['authentication_required'],
  403: ['origin_forbidden'],
  409: ['course_already_owned'],
} as const satisfies Readonly<Record<number, readonly PurchaseErrorCode[]>>;

const statusErrors = {
  400: ['invalid_request'],
  401: ['authentication_required'],
  404: ['order_not_found'],
} as const satisfies Readonly<Record<number, readonly PurchaseErrorCode[]>>;

function failure(code: PurchaseErrorCode): PurchaseResult<never> {
  return { ok: false, error: { code } };
}

function namedError(
  response: WireResponse,
  allowed: Readonly<Record<number, readonly PurchaseErrorCode[]>>,
): PurchaseErrorCode {
  const parsed = errorSchema.safeParse(response.data);
  const code = parsed.success ? parsed.data.error.code : undefined;
  const named = allowed[response.status]?.find((candidate) => candidate === code);
  return named ?? (unnamedStatusCode(response.status) as PurchaseErrorCode);
}

function order(data: unknown, expectedId?: string): PurchaseResult<PurchaseOrder> {
  const parsed = orderSchema.safeParse(data);
  if (!parsed.success || (expectedId !== undefined && parsed.data.id !== expectedId)) {
    return failure('invalid_response');
  }
  return { ok: true, value: parsed.data };
}

/** One repository per router request; generated transport types never leave this adapter. */
export function createApiPurchaseRepository(options: TransportOptions = {}): PurchaseRepository {
  const send = createApiTransport(options);
  return {
    async start(courseId, idempotencyKey, signal) {
      const outcome = await send(
        'learn.purchase.start',
        (init) => {
          // Merge, never replace: the transport's own headers (Accept, Cookie on the server)
          // travel next to the replay key.
          const headers = new Headers(init.headers);
          headers.set('Idempotency-Key', idempotencyKey);
          return startCoursePurchase(courseId, { ...init, headers });
        },
        201,
        signal,
      );
      if (!outcome.ok) return failure(outcome.code as PurchaseErrorCode);
      if (outcome.status !== 201) return failure(namedError(outcome, startErrors));
      return order(outcome.data);
    },
    async status(orderId, signal) {
      const outcome = await send(
        'learn.purchase.status',
        (init) => getPurchaseStatus(orderId, init),
        200,
        signal,
      );
      if (!outcome.ok) return failure(outcome.code as PurchaseErrorCode);
      if (outcome.status !== 200) return failure(namedError(outcome, statusErrors));
      return order(outcome.data, orderId);
    },
  };
}
