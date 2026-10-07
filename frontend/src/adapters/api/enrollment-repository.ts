import {
  enrollInCourse,
  getMyCourseMembership,
  getMyCourseMemberships,
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
} from '../../domain/enrollment/enrollment';

function record(value: unknown): value is Record<string, unknown> {
  return typeof value === 'object' && value !== null;
}
function uuid(value: unknown): value is string {
  return typeof value === 'string' && /^[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$/i.test(value);
}
function date(value: unknown): value is string {
  return typeof value === 'string' && Number.isFinite(Date.parse(value));
}
function membershipState(value: unknown): value is MembershipState {
  return value === 'active' || value === 'blocked' || value === 'revoked';
}
function membership(value: unknown): value is Membership {
  return (
    record(value) &&
    typeof value.member === 'boolean' &&
    (value.state === null || membershipState(value.state)) &&
    (value.member ? value.state !== null : value.state === null)
  );
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
function enrolled(value: unknown): Membership | null {
  return record(value) && uuid(value.courseId) && value.state === 'active'
    ? { member: true, state: 'active' }
    : null;
}
function failure(code: EnrollmentErrorCode): EnrollmentResult<never> {
  return { ok: false, error: { code } };
}
function namedError(
  response: WireResponse,
  allowed: Readonly<Record<number, readonly EnrollmentErrorCode[]>>,
): EnrollmentErrorCode {
  const error = record(response.data) && record(response.data.error) ? response.data.error.code : undefined;
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
      return membership(outcome.data) ? { ok: true, value: outcome.data } : failure('invalid_response');
    },
    async enroll(courseId, signal) {
      const outcome = await send('enrollment.enroll', (init) => enrollInCourse(courseId, init), 201, signal);
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
  };
}
