import type {
  AuthenticatedUser,
  SessionError,
  SessionResult,
} from '../../domain/identity/identity';
import type { SessionRepository } from '../../ports/SessionRepository';
import {
  changePassword,
  confirmEmailChange,
  getCurrentUser,
  login,
  logout,
  register,
  requestEmailChange,
  requestPasswordReset,
  resendVerification,
  resetPassword,
  verifyEmail,
} from './generated/courses';
import { errorSchema } from './course-schema';
import { authenticatedUserSchema, emptyResponseSchema } from './session-schema';
import { createApiTransport, unnamedStatusCode } from './transport';
import type { TransportOptions, WireResponse } from './transport';

const invalid = { ok: false, error: { code: 'invalid_response' } } satisfies SessionResult<never>;

/**
 * The server answers a rejected login and an unusable session with distinct
 * codes; anything it does not name collapses to a transport fault rather than
 * being reported as a credential decision.
 */
function wireError(response: WireResponse): SessionError {
  const parsed = errorSchema.safeParse(response.data);
  if (parsed.success) {
    switch (parsed.data.error.code) {
      case 'invalid_credentials':
        if (response.status === 401) return { code: 'invalid_credentials' };
        break;
      case 'email_taken':
        if (response.status === 409) return { code: 'email_taken' };
        break;
      case 'email_not_verified':
        return { code: 'email_not_verified' };
      case 'weak_password':
        if (response.status === 422) return { code: 'weak_password' };
        break;
      case 'password_unchanged':
        if (response.status === 422) return { code: 'password_unchanged' };
        break;
      case 'invalid_account_token':
        if (response.status === 400) return { code: 'invalid_account_token' };
        break;
      case 'token_used':
        if (response.status === 410) return { code: 'token_used' };
        break;
      case 'rate_limited':
        if (response.status === 429) return { code: 'rate_limited' };
        break;
      case 'authentication_required':
        if (response.status === 401) return { code: 'authentication_required' };
        break;
      case 'invalid_request':
        if (response.status === 400) return { code: 'invalid_request' };
        break;
      case 'origin_forbidden':
        if (response.status === 403) return { code: 'origin_forbidden' };
        break;
    }
  }
  return { code: unnamedStatusCode(response.status) };
}

function user(raw: unknown): SessionResult<AuthenticatedUser> {
  const parsed = authenticatedUserSchema.safeParse(raw);
  return parsed.success ? { ok: true, value: parsed.data } : invalid;
}

function empty(raw: unknown): SessionResult<null> {
  return emptyResponseSchema.safeParse(raw).success ? { ok: true, value: null } : invalid;
}

/** One repository per SSR request; no cookie-bearing singleton or shared session cache. */
export function createApiSessionRepository(options: TransportOptions = {}): SessionRepository {
  const send = createApiTransport(options);
  async function request<T>(
    operation: string,
    execute: (init: RequestInit) => Promise<WireResponse>,
    expected: number,
    map: (data: unknown) => SessionResult<T>,
    caller?: AbortSignal,
  ): Promise<SessionResult<T>> {
    const outcome = await send(operation, execute, expected, caller);
    if (!outcome.ok) return { ok: false, error: { code: outcome.code } };
    if (outcome.status === expected) return map(outcome.data);
    return { ok: false, error: wireError(outcome) };
  }
  return {
    logIn(credentials, signal) {
      return request<AuthenticatedUser>(
        'identity.login',
        (init) => login(credentials, init),
        200,
        user,
        signal,
      );
    },
    register(input, signal) {
      return request<AuthenticatedUser>(
        'identity.register',
        (init) => register(input, init),
        201,
        user,
        signal,
      );
    },
    verifyEmail(input, signal) {
      return request<null>(
        'identity.verifyEmail',
        (init) => verifyEmail(input, init),
        204,
        empty,
        signal,
      );
    },
    resendVerification(input, signal) {
      return request<null>(
        'identity.resendVerification',
        (init) => resendVerification(input, init),
        204,
        empty,
        signal,
      );
    },
    forgotPassword(input, signal) {
      return request<null>(
        'identity.forgotPassword',
        (init) => requestPasswordReset(input, init),
        200,
        empty,
        signal,
      );
    },
    resetPassword(input, signal) {
      return request<null>(
        'identity.resetPassword',
        (init) => resetPassword(input, init),
        204,
        empty,
        signal,
      );
    },
    changePassword(input, signal) {
      return request<null>(
        'identity.changePassword',
        (init) => changePassword(input, init),
        204,
        empty,
        signal,
      );
    },
    requestEmailChange(input, signal) {
      return request<null>(
        'identity.requestEmailChange',
        (init) => requestEmailChange(input, init),
        204,
        empty,
        signal,
      );
    },
    confirmEmailChange(input, signal) {
      return request<null>(
        'identity.confirmEmailChange',
        (init) => confirmEmailChange(input, init),
        204,
        empty,
        signal,
      );
    },
    async currentUser(signal) {
      const result = await request<AuthenticatedUser | null>(
        'identity.currentUser',
        (init) => getCurrentUser(init),
        200,
        user,
        signal,
      );
      // Anonymous is the ordinary state of a public page, not a failure.
      return !result.ok && result.error.code === 'authentication_required'
        ? { ok: true, value: null }
        : result;
    },
    logOut(signal) {
      return request<null>(
        'identity.logout',
        (init) => logout(init),
        204,
        () => ({ ok: true, value: null }),
        signal,
      );
    },
  };
}
