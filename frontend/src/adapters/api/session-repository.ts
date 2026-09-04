import { context, SpanStatusCode, trace } from '@opentelemetry/api';
import type {
  AuthenticatedUser,
  SessionError,
  SessionResult,
} from '../../domain/identity/identity';
import type { SessionRepository } from '../../ports/SessionRepository';
import { getCurrentUser, login, logout } from './generated/courses';
import { apiRequestHeaders } from './request-headers';
import { errorSchema } from './course-schema';
import { authenticatedUserSchema } from './session-schema';

type Options = {
  readonly cookie?: string;
  readonly requestSignal?: AbortSignal;
  readonly timeoutMs?: number;
};
type WireResponse = { readonly data: unknown; readonly status: number };
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
  return {
    code:
      response.status >= 500 || response.status === 429
        ? 'dependency_unavailable'
        : 'invalid_response',
  };
}

/** One repository per SSR request; no cookie-bearing singleton or shared session cache. */
export function createApiSessionRepository(options: Options = {}): SessionRepository {
  const tracer = trace.getTracer('courses.frontend.api');
  async function request<T>(
    operation: string,
    execute: (init: RequestInit) => Promise<WireResponse>,
    expected: number,
    map: (data: unknown) => SessionResult<T>,
    caller?: AbortSignal,
    recover?: (error: SessionError) => SessionResult<T> | undefined,
  ): Promise<SessionResult<T>> {
    const deadline = AbortSignal.timeout(options.timeoutMs ?? 5000);
    const controls = [caller, options.requestSignal].filter(
      (signal): signal is AbortSignal => signal !== undefined,
    );
    const signal = AbortSignal.any([...controls, deadline]);
    const headers = apiRequestHeaders(options.cookie);
    return tracer.startActiveSpan(
      operation,
      {},
      context.active(),
      async (span): Promise<SessionResult<T>> => {
        try {
          const response = await execute({
            headers,
            signal,
            credentials: 'same-origin',
            redirect: 'error',
            cache: 'no-store',
          });
          if (response.status !== expected) {
            const error = wireError(response);
            const recovered = recover?.(error);
            if (recovered) return recovered;
            span.setStatus({ code: SpanStatusCode.ERROR });
            return { ok: false, error };
          }
          return map(response.data);
        } catch (error) {
          span.setStatus({ code: SpanStatusCode.ERROR });
          // Control signals remain distinct from malformed JSON and network errors.
          if (controls.some((control) => control.aborted))
            return { ok: false, error: { code: 'request_cancelled' } };
          if (deadline.aborted) return { ok: false, error: { code: 'request_timeout' } };
          if (error instanceof SyntaxError) return invalid;
          return { ok: false, error: { code: 'dependency_unavailable' } };
        } finally {
          span.end();
        }
      },
    );
  }
  function user(raw: unknown): SessionResult<AuthenticatedUser> {
    const parsed = authenticatedUserSchema.safeParse(raw);
    return parsed.success ? { ok: true, value: parsed.data } : invalid;
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
    currentUser(signal) {
      return request<AuthenticatedUser | null>(
        'identity.currentUser',
        (init) => getCurrentUser(init),
        200,
        user,
        signal,
        // Anonymous is the ordinary state of a public page, not a failure.
        (error) =>
          error.code === 'authentication_required' ? { ok: true, value: null } : undefined,
      );
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
