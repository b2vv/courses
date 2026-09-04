import { context, SpanStatusCode, trace } from '@opentelemetry/api';
import { apiRequestHeaders } from './request-headers';

export type TransportOptions = {
  readonly cookie?: string;
  readonly requestSignal?: AbortSignal;
  readonly timeoutMs?: number;
};
export type WireResponse = { readonly data: unknown; readonly status: number };

/**
 * How a call to our own API can fail before any domain meaning exists. Every
 * domain error union in this application includes these four codes, so an
 * adapter can widen an outcome into its own error without translating.
 */
export type TransportFailureCode =
  | 'request_cancelled'
  | 'request_timeout'
  | 'invalid_response'
  | 'dependency_unavailable';

export type TransportOutcome =
  | { readonly ok: true; readonly status: number; readonly data: unknown }
  | { readonly ok: false; readonly code: TransportFailureCode };

const defaultTimeoutMs = 5000;

/**
 * What a status the contract does not name means. A server fault or a refusal
 * by the edge is a dependency problem the caller may retry; anything else is an
 * answer we cannot interpret, and both stay distinct from a named business
 * error that an adapter recognises itself.
 */
export function unnamedStatusCode(status: number): TransportFailureCode {
  return status >= 500 || status === 429 ? 'dependency_unavailable' : 'invalid_response';
}

/**
 * One request-scoped transport shared by every API adapter: the cookie the SSR
 * request carried, one deadline, the caller's cancellation, and the single
 * place that decides whether a thrown value was a cancellation, a timeout, a
 * malformed body or an unreachable dependency. Adapters keep their own status
 * and schema rules, which is where they legitimately differ.
 */
export function createApiTransport(options: TransportOptions = {}) {
  const tracer = trace.getTracer('courses.frontend.api');
  return function send(
    operation: string,
    execute: (init: RequestInit) => Promise<WireResponse>,
    expected: number,
    caller?: AbortSignal,
  ): Promise<TransportOutcome> {
    const deadline = AbortSignal.timeout(options.timeoutMs ?? defaultTimeoutMs);
    const controls = [caller, options.requestSignal].filter(
      (signal): signal is AbortSignal => signal !== undefined,
    );
    const signal = AbortSignal.any([...controls, deadline]);
    return tracer.startActiveSpan(
      operation,
      {},
      context.active(),
      async (span): Promise<TransportOutcome> => {
        try {
          const response = await execute({
            headers: apiRequestHeaders(options.cookie),
            signal,
            credentials: 'same-origin',
            redirect: 'error',
            cache: 'no-store',
          });
          if (response.status !== expected) span.setStatus({ code: SpanStatusCode.ERROR });
          return { ok: true, status: response.status, data: response.data };
        } catch (error) {
          span.setStatus({ code: SpanStatusCode.ERROR });
          // Control signals remain distinct from malformed JSON and network errors.
          if (controls.some((control) => control.aborted))
            return { ok: false, code: 'request_cancelled' };
          if (deadline.aborted) return { ok: false, code: 'request_timeout' };
          if (error instanceof SyntaxError) return { ok: false, code: 'invalid_response' };
          return { ok: false, code: 'dependency_unavailable' };
        } finally {
          span.end();
        }
      },
    );
  };
}
