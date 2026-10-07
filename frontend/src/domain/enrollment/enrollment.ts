export type MembershipState = 'active' | 'blocked' | 'revoked';

export type Membership = {
  readonly member: boolean;
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

export type EnrollmentErrorCode =
  | 'request_cancelled'
  | 'request_timeout'
  | 'invalid_response'
  | 'dependency_unavailable'
  | 'invalid_request'
  | 'authentication_required'
  | 'enrollment_required'
  | 'enrollment_blocked'
  | 'membership_conflict'
  | 'course_not_enrollable'
  | 'enrollment_forbidden'
  | 'course_already_owned'
  | 'email_not_verified';

export type EnrollmentResult<T> =
  | { readonly ok: true; readonly value: T }
  | { readonly ok: false; readonly error: { readonly code: EnrollmentErrorCode } };
