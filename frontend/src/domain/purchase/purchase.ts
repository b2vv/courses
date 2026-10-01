export type PurchaseStatus = 'pending' | 'paid' | 'failed';

export type PurchasePrice = { readonly minor: number; readonly currency: 'UAH' };

/** A course purchase as the buyer sees it. The amount is the price snapshot taken at start. */
export type PurchaseOrder = {
  readonly id: string;
  readonly courseId: string;
  readonly amount: PurchasePrice;
  readonly status: PurchaseStatus;
  readonly createdAt: string;
  readonly paidAt: string | null;
};

export type PurchaseErrorCode =
  | 'request_cancelled'
  | 'request_timeout'
  | 'invalid_response'
  | 'dependency_unavailable'
  | 'invalid_request'
  | 'authentication_required'
  | 'origin_forbidden'
  | 'course_already_owned'
  | 'order_not_found';

export type PurchaseResult<T> =
  | { readonly ok: true; readonly value: T }
  | { readonly ok: false; readonly error: { readonly code: PurchaseErrorCode } };

/** The order no longer changes on its own: stop polling. */
export function isSettled(order: PurchaseOrder): boolean {
  return order.status !== 'pending';
}
