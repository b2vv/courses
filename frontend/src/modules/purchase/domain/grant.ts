export type CourseGrant = {
  readonly id: string;
  readonly userId: string;
  readonly courseId: string;
  readonly grantedAt: string;
};

export type GrantErrorCode =
  | 'request_cancelled'
  | 'request_timeout'
  | 'invalid_response'
  | 'dependency_unavailable'
  | 'invalid_request'
  | 'authentication_required'
  | 'origin_forbidden'
  | 'grant_forbidden'
  | 'grant_self_forbidden'
  | 'admin_not_learner'
  | 'course_already_owned';

export type GrantResult<T> =
  | { readonly ok: true; readonly value: T }
  | { readonly ok: false; readonly error: { readonly code: GrantErrorCode } };

export const grantReasonMaximum = 500;

/** The audit record needs a reason a person can read later; the server trims it the same way. */
export function grantReason(text: string): string | null {
  const reason = text.trim();
  return reason.length >= 1 && [...reason].length <= grantReasonMaximum ? reason : null;
}
