export type MembershipState = 'active' | 'blocked' | 'revoked';

// State is authoritative. There is deliberately no `member` boolean: the backend sets member=true
// whenever a row exists (including revoked), so branching on it is a footgun (it caused the L7a
// revoked-as-member bug). Membership is `active`|`blocked`|`revoked`|null; access follows state.
export type Membership = {
  readonly state: MembershipState | null;
};

export type MembershipCard = {
  readonly courseId: string;
  readonly slug: string;
  readonly title: string;
  readonly state: Extract<MembershipState, 'active' | 'blocked'>;
  readonly progressPercent: number;
  readonly enrolledAt: string;
};

export type RosterEntry = {
  readonly userId: string;
  readonly displayName: string;
  readonly state: Extract<MembershipState, 'active' | 'blocked'>;
  readonly enrolledAt: string;
  readonly progressPercent: number;
};

export type RosterPage = {
  readonly entries: readonly RosterEntry[];
  readonly nextCursor: string | null;
};

export type ParticipantReportKind = 'on_lecturer' | 'on_student';
export type ParticipantReportState = 'open' | 'resolved';

// The queue can be narrowed to one state or span every state ('all'). 'open' is the
// default the admin screen loads; the backend reads an absent filter as 'open' too.
export type ParticipantReportStateFilter = ParticipantReportState | 'all';

export type ParticipantReportQueueItem = {
  readonly id: string;
  readonly courseId: string;
  readonly reporterId: string;
  readonly targetId: string;
  readonly kind: ParticipantReportKind;
  readonly reason: string;
  readonly state: ParticipantReportState;
  readonly createdAt: string;
  readonly reporterName: string | null;
  readonly targetName: string | null;
  readonly courseTitle: string;
};

export type ParticipantReportQueuePage = {
  readonly items: readonly ParticipantReportQueueItem[];
  readonly nextCursor: string | null;
};

// 'unknown' is a row flagged before refund_reason existed (backend sends ""); the screen renders it
// as such rather than hiding the discrepancy.
export type RefundReason = 'settlement_rejected' | 'grant_conflict' | 'unknown';

// The administrator's recorded decision on a refund-queue order. Decoupled from the membership
// revoke: an order can be refunded or declined with the grant left untouched, and access can be
// revoked without settling the refund.
export type RefundOutcome = 'refunded' | 'declined';

// Narrows the refund queue. 'pending' is the default (not yet settled); 'all' clears the filter.
export type RefundStateFilter = 'pending' | 'refunded' | 'declined' | 'all';

export type RefundQueueItem = {
  readonly orderId: string;
  readonly courseId: string;
  readonly userId: string;
  readonly orderAmountMinor: number;
  readonly chargedAmountMinor: number | null;
  readonly receivedAt: string | null;
  readonly reason: RefundReason;
  // Current (user, course) grant state, null when no grant exists; revoke is offered only for a
  // live grant (active/blocked) — A14.
  readonly membershipState: MembershipState | null;
  // The settled decision, or null while the order is still pending in the queue.
  readonly outcome: RefundOutcome | null;
};

export type RefundQueuePage = {
  readonly items: readonly RefundQueueItem[];
};

export type EnrollmentErrorCode =
  | 'request_cancelled'
  | 'request_timeout'
  | 'invalid_response'
  | 'dependency_unavailable'
  | 'internal_error'
  | 'invalid_request'
  | 'authentication_required'
  | 'enrollment_required'
  | 'enrollment_blocked'
  | 'membership_conflict'
  | 'course_not_enrollable'
  | 'enrollment_forbidden'
  | 'course_already_owned'
  | 'email_not_verified'
  | 'report_own_person'
  | 'report_invalid'
  | 'report_forbidden'
  | 'report_already_open'
  | 'comment_cursor_invalid'
  | 'course_not_found'
  | 'grant_forbidden'
  | 'grant_transition_invalid'
  | 'order_not_found';

export type EnrollmentResult<T> =
  | { readonly ok: true; readonly value: T }
  | { readonly ok: false; readonly error: { readonly code: EnrollmentErrorCode } };
