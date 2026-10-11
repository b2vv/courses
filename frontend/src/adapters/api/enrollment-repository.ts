import {
  blockCourseParticipant,
  enrollInCourse,
  getCourseRoster,
  getMyCourseMembership,
  getMyCourseMemberships,
  listParticipantReports,
  listRefundQueue,
  reportCourseLecturer,
  reportCourseParticipant,
  resolveParticipantReport,
  resolveRefund,
  revokeCourseParticipant,
  revokeRefundGrant,
  unblockCourseParticipant,
} from './generated/courses';
import { createApiTransport, unnamedStatusCode } from './transport';
import type { TransportOptions, WireResponse } from './transport';
import type { EnrollmentRepository } from '../../ports/EnrollmentRepository';
import type {
  EnrollmentErrorCode,
  EnrollmentResult,
  Membership,
  MembershipCard,
  MembershipState,
  ParticipantReportQueueItem,
  ParticipantReportQueuePage,
  RefundQueueItem,
  RefundQueuePage,
  RefundReason,
  RosterEntry,
  RosterPage,
} from '../../domain/enrollment/enrollment';

function record(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null;
}
function uuid(value: unknown): value is string {
  return (
    typeof value === 'string' &&
    /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i.test(value)
  );
}
function date(value: unknown): value is string {
  return typeof value === 'string' && Number.isFinite(Date.parse(value));
}
function membershipState(value: unknown): value is MembershipState {
  return value === 'active' || value === 'blocked' || value === 'revoked';
}
function membership(value: unknown): value is Membership {
  return record(value) && (value.state === null || membershipState(value.state));
}
function card(value: unknown): value is MembershipCard {
  return (
    record(value) &&
    uuid(value.courseId) &&
    typeof value.slug === 'string' &&
    /^[a-z0-9]+(?:-[a-z0-9]+)*$/.test(value.slug) &&
    typeof value.title === 'string' &&
    value.title.length > 0 &&
    value.title.length <= 200 &&
    (value.state === 'active' || value.state === 'blocked') &&
    typeof value.progressPercent === 'number' &&
    Number.isInteger(value.progressPercent) &&
    value.progressPercent >= 0 &&
    value.progressPercent <= 100 &&
    date(value.enrolledAt)
  );
}
function membershipsPage(value: unknown): value is { readonly items: readonly MembershipCard[] } {
  return record(value) && Array.isArray(value.items) && value.items.every(card);
}
function rosterEntry(value: unknown): value is RosterEntry {
  return (
    record(value) &&
    uuid(value.userId) &&
    typeof value.displayName === 'string' &&
    value.displayName.length <= 120 &&
    (value.state === 'active' || value.state === 'blocked') &&
    date(value.enrolledAt) &&
    typeof value.progressPercent === 'number' &&
    Number.isInteger(value.progressPercent) &&
    value.progressPercent >= 0 &&
    value.progressPercent <= 100
  );
}
function rosterPage(value: unknown): value is RosterPage {
  return (
    record(value) &&
    Array.isArray(value.entries) &&
    value.entries.length <= 100 &&
    value.entries.every(rosterEntry) &&
    (value.nextCursor === null || typeof value.nextCursor === 'string')
  );
}
function participantReportItem(value: unknown): value is ParticipantReportQueueItem {
  return (
    record(value) &&
    uuid(value.id) &&
    uuid(value.courseId) &&
    uuid(value.reporterId) &&
    uuid(value.targetId) &&
    (value.kind === 'on_lecturer' || value.kind === 'on_student') &&
    typeof value.reason === 'string' &&
    value.reason.length >= 1 &&
    value.reason.length <= 500 &&
    (value.state === 'open' || value.state === 'resolved') &&
    date(value.createdAt) &&
    (value.reporterName === null || typeof value.reporterName === 'string') &&
    (value.targetName === null || typeof value.targetName === 'string') &&
    typeof value.courseTitle === 'string'
  );
}
function participantReportPage(value: unknown): value is ParticipantReportQueuePage {
  return (
    record(value) &&
    Array.isArray(value.items) &&
    value.items.length <= 20 &&
    value.items.every(participantReportItem) &&
    (value.nextCursor === null ||
      (typeof value.nextCursor === 'string' && value.nextCursor.length > 0))
  );
}
function refundReason(value: unknown): RefundReason | null {
  if (value === 'settlement_rejected' || value === 'grant_conflict') return value;
  if (value === '') return 'unknown';
  return null;
}
function refundItem(value: unknown): RefundQueueItem | null {
  if (
    !record(value) ||
    !uuid(value.orderId) ||
    !uuid(value.courseId) ||
    !uuid(value.userId) ||
    typeof value.orderAmountMinor !== 'number' ||
    !Number.isInteger(value.orderAmountMinor) ||
    value.orderAmountMinor < 1 ||
    !(
      value.chargedAmountMinor === null ||
      (typeof value.chargedAmountMinor === 'number' && Number.isInteger(value.chargedAmountMinor))
    ) ||
    !(value.receivedAt === null || date(value.receivedAt))
  ) {
    return null;
  }
  const reason = refundReason(value.reason);
  if (reason === null) return null;
  if (!(value.membershipState === null || membershipState(value.membershipState))) return null;
  return {
    orderId: value.orderId,
    courseId: value.courseId,
    userId: value.userId,
    orderAmountMinor: value.orderAmountMinor,
    chargedAmountMinor: value.chargedAmountMinor,
    receivedAt: value.receivedAt,
    reason,
    membershipState: value.membershipState,
  };
}
function refundPage(value: unknown): RefundQueuePage | null {
  if (!record(value) || !Array.isArray(value.items)) return null;
  const items: RefundQueueItem[] = [];
  for (const raw of value.items) {
    const item = refundItem(raw);
    if (item === null) return null;
    items.push(item);
  }
  return { items };
}
function membershipGrant(value: unknown): boolean {
  return (
    record(value) &&
    uuid(value.id) &&
    uuid(value.userId) &&
    uuid(value.courseId) &&
    typeof value.source === 'string' &&
    membershipState(value.state) &&
    date(value.grantedAt) &&
    (value.grantedBy === null || uuid(value.grantedBy))
  );
}
function enrolled(value: unknown): Membership | null {
  return record(value) && uuid(value.courseId) && value.state === 'active'
    ? { state: 'active' }
    : null;
}
function failure(code: EnrollmentErrorCode): EnrollmentResult<never> {
  return { ok: false, error: { code } };
}
function namedError(
  response: WireResponse,
  allowed: Readonly<Record<number, readonly EnrollmentErrorCode[]>>,
): EnrollmentErrorCode {
  const error =
    record(response.data) && record(response.data.error) ? response.data.error.code : undefined;
  const named = allowed[response.status]?.find((candidate) => candidate === error);
  return named ?? (unnamedStatusCode(response.status) as EnrollmentErrorCode);
}
const readErrors = {
  400: ['invalid_request'],
  401: ['authentication_required'],
  403: ['enrollment_forbidden'],
} as const satisfies Readonly<Record<number, readonly EnrollmentErrorCode[]>>;
const enrollErrors = {
  400: ['invalid_request'],
  401: ['authentication_required'],
  403: ['enrollment_forbidden', 'email_not_verified'],
  409: ['membership_conflict', 'course_not_enrollable', 'course_already_owned'],
} as const satisfies Readonly<Record<number, readonly EnrollmentErrorCode[]>>;
const rosterErrors = {
  400: ['invalid_request'],
  401: ['authentication_required'],
  403: ['enrollment_forbidden'],
} as const satisfies Readonly<Record<number, readonly EnrollmentErrorCode[]>>;
const membershipMutationErrors = {
  400: ['invalid_request'],
  401: ['authentication_required'],
  403: ['enrollment_forbidden'],
  409: ['membership_conflict'],
} as const satisfies Readonly<Record<number, readonly EnrollmentErrorCode[]>>;
const reportErrors = {
  400: ['report_invalid', 'report_own_person'],
  401: ['authentication_required'],
  403: ['report_forbidden', 'email_not_verified'],
  404: ['course_not_found'],
  409: ['report_already_open'],
} as const satisfies Readonly<Record<number, readonly EnrollmentErrorCode[]>>;
const participantReportListErrors = {
  400: ['comment_cursor_invalid'],
  401: ['authentication_required'],
  403: ['report_forbidden'],
  503: ['dependency_unavailable'],
} as const satisfies Readonly<Record<number, readonly EnrollmentErrorCode[]>>;
const participantReportResolveErrors = {
  401: ['authentication_required'],
  403: ['report_forbidden'],
  503: ['dependency_unavailable'],
} as const satisfies Readonly<Record<number, readonly EnrollmentErrorCode[]>>;
const revokeErrors = {
  400: ['invalid_request'],
  401: ['authentication_required'],
  403: ['enrollment_forbidden'],
  409: ['membership_conflict'],
  500: ['internal_error'],
  503: ['dependency_unavailable'],
} as const satisfies Readonly<Record<number, readonly EnrollmentErrorCode[]>>;
const refundListErrors = {
  400: ['invalid_request'],
  401: ['authentication_required'],
  403: ['grant_forbidden'],
  500: ['internal_error'],
  503: ['dependency_unavailable'],
} as const satisfies Readonly<Record<number, readonly EnrollmentErrorCode[]>>;
const refundResolveErrors = {
  400: ['invalid_request'],
  401: ['authentication_required'],
  403: ['grant_forbidden'],
  500: ['internal_error'],
  503: ['dependency_unavailable'],
} as const satisfies Readonly<Record<number, readonly EnrollmentErrorCode[]>>;
const refundRevokeErrors = {
  400: ['invalid_request'],
  401: ['authentication_required'],
  403: ['grant_forbidden'],
  404: ['order_not_found'],
  409: ['membership_conflict', 'grant_transition_invalid'],
  500: ['internal_error'],
  503: ['dependency_unavailable'],
} as const satisfies Readonly<Record<number, readonly EnrollmentErrorCode[]>>;

export function createApiEnrollmentRepository(
  options: TransportOptions = {},
): EnrollmentRepository {
  const send = createApiTransport(options);
  return {
    async membership(courseId, signal) {
      const outcome = await send(
        'enrollment.membership',
        (init) => getMyCourseMembership(courseId, init),
        200,
        signal,
      );
      if (!outcome.ok) return failure(outcome.code);
      if (outcome.status !== 200) return failure(namedError(outcome, readErrors));
      return membership(outcome.data)
        ? { ok: true, value: outcome.data }
        : failure('invalid_response');
    },
    async enroll(courseId, signal) {
      const outcome = await send(
        'enrollment.enroll',
        (init) => enrollInCourse(courseId, init),
        201,
        signal,
      );
      if (!outcome.ok) return failure(outcome.code);
      if (outcome.status !== 201) return failure(namedError(outcome, enrollErrors));
      const value = enrolled(outcome.data);
      return value === null ? failure('invalid_response') : { ok: true, value };
    },
    async myMemberships(signal) {
      const outcome = await send(
        'enrollment.my-memberships',
        (init) => getMyCourseMemberships(init),
        200,
        signal,
      );
      if (!outcome.ok) return failure(outcome.code);
      if (outcome.status !== 200) return failure(namedError(outcome, readErrors));
      return membershipsPage(outcome.data)
        ? { ok: true, value: outcome.data.items }
        : failure('invalid_response');
    },
    async roster(courseId, cursor, signal) {
      const outcome = await send(
        'enrollment.roster',
        (init) => getCourseRoster(courseId, cursor === undefined ? undefined : { cursor }, init),
        200,
        signal,
      );
      if (!outcome.ok) return failure(outcome.code);
      if (outcome.status !== 200) return failure(namedError(outcome, rosterErrors));
      return rosterPage(outcome.data)
        ? { ok: true, value: outcome.data }
        : failure('invalid_response');
    },
    async block(courseId, userId, signal) {
      const outcome = await send(
        'enrollment.block',
        (init) => blockCourseParticipant(courseId, userId, init),
        200,
        signal,
      );
      if (!outcome.ok) return failure(outcome.code);
      if (outcome.status !== 200) return failure(namedError(outcome, membershipMutationErrors));
      return membershipGrant(outcome.data)
        ? { ok: true, value: undefined }
        : failure('invalid_response');
    },
    async unblock(courseId, userId, signal) {
      const outcome = await send(
        'enrollment.unblock',
        (init) => unblockCourseParticipant(courseId, userId, init),
        200,
        signal,
      );
      if (!outcome.ok) return failure(outcome.code);
      if (outcome.status !== 200) return failure(namedError(outcome, membershipMutationErrors));
      return membershipGrant(outcome.data)
        ? { ok: true, value: undefined }
        : failure('invalid_response');
    },
    async reportLecturer(courseId, reason, signal) {
      const outcome = await send(
        'enrollment.report-lecturer',
        (init) => reportCourseLecturer(courseId, { reason }, init),
        201,
        signal,
      );
      if (!outcome.ok) return failure(outcome.code);
      if (outcome.status !== 201) return failure(namedError(outcome, reportErrors));
      return record(outcome.data) ? { ok: true, value: undefined } : failure('invalid_response');
    },
    async reportStudent(courseId, userId, reason, signal) {
      const outcome = await send(
        'enrollment.report-student',
        (init) => reportCourseParticipant(courseId, userId, { reason }, init),
        201,
        signal,
      );
      if (!outcome.ok) return failure(outcome.code);
      if (outcome.status !== 201) return failure(namedError(outcome, reportErrors));
      return record(outcome.data) ? { ok: true, value: undefined } : failure('invalid_response');
    },
    async participantReports(cursor, signal) {
      const outcome = await send(
        'enrollment.participant-reports',
        (init) => listParticipantReports(cursor === undefined ? undefined : { cursor }, init),
        200,
        signal,
      );
      if (!outcome.ok) return failure(outcome.code);
      if (outcome.status !== 200) return failure(namedError(outcome, participantReportListErrors));
      return participantReportPage(outcome.data)
        ? { ok: true, value: outcome.data }
        : failure('invalid_response');
    },
    async resolveReport(reportId, signal) {
      const outcome = await send(
        'enrollment.resolve-participant-report',
        (init) => resolveParticipantReport(reportId, init),
        204,
        signal,
      );
      if (!outcome.ok) return failure(outcome.code);
      return outcome.status === 204
        ? { ok: true, value: undefined }
        : failure(namedError(outcome, participantReportResolveErrors));
    },
    async revoke(courseId, userId, signal) {
      const outcome = await send(
        'enrollment.revoke',
        (init) => revokeCourseParticipant(courseId, userId, init),
        200,
        signal,
      );
      if (!outcome.ok) return failure(outcome.code);
      if (outcome.status !== 200) return failure(namedError(outcome, revokeErrors));
      return membershipGrant(outcome.data)
        ? { ok: true, value: undefined }
        : failure('invalid_response');
    },
    async refundQueue(signal) {
      const outcome = await send(
        'enrollment.refund-queue',
        (init) => listRefundQueue(init),
        200,
        signal,
      );
      if (!outcome.ok) return failure(outcome.code);
      if (outcome.status !== 200) return failure(namedError(outcome, refundListErrors));
      const page = refundPage(outcome.data);
      return page ? { ok: true, value: page } : failure('invalid_response');
    },
    async resolveRefund(orderId, signal) {
      const outcome = await send(
        'enrollment.resolve-refund',
        (init) => resolveRefund(orderId, init),
        204,
        signal,
      );
      if (!outcome.ok) return failure(outcome.code);
      return outcome.status === 204
        ? { ok: true, value: undefined }
        : failure(namedError(outcome, refundResolveErrors));
    },
    async revokeRefund(orderId, signal) {
      const outcome = await send(
        'enrollment.revoke-refund',
        (init) => revokeRefundGrant(orderId, init),
        200,
        signal,
      );
      if (!outcome.ok) return failure(outcome.code);
      if (outcome.status !== 200) return failure(namedError(outcome, refundRevokeErrors));
      return membershipGrant(outcome.data)
        ? { ok: true, value: undefined }
        : failure('invalid_response');
    },
  };
}
