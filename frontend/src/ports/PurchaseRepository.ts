import type { PurchaseOrder, PurchaseResult } from '../domain/purchase/purchase';

export interface PurchaseRepository {
  /** Starts (or replays, for the same key) a purchase of a course. */
  start(
    courseId: string,
    idempotencyKey: string,
    signal?: AbortSignal,
  ): Promise<PurchaseResult<PurchaseOrder>>;
  status(orderId: string, signal?: AbortSignal): Promise<PurchaseResult<PurchaseOrder>>;
}
