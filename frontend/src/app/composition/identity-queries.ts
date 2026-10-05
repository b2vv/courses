import { queryOptions } from '@tanstack/react-query';
import type { QueryClient } from '@tanstack/react-query';
import type { createSession } from '../../application/identity/session';
import type { SessionError, SessionResult } from '../../domain/identity/identity';

export class SessionQueryError extends Error {
  readonly code: SessionError['code'];
  constructor(error: SessionError) {
    super(error.code);
    this.name = 'SessionQueryError';
    this.code = error.code;
  }
}
function unwrap<T>(result: SessionResult<T>): T {
  if (!result.ok) throw new SessionQueryError(result.error);
  return result.value;
}

/** One key for the viewer's identity; logging in or out replaces it wholesale. */
export const sessionQueryKey = ['session'] as const;

export function createSessionQueries(session: ReturnType<typeof createSession>) {
  return {
    session,
    current: () =>
      queryOptions({
        queryKey: sessionQueryKey,
        queryFn: async ({ signal }) => unwrap(await session.current(signal)),
        // Identity is never considered fresh, so every mount revalidates it and a
        // sign-in or sign-out elsewhere cannot leave a stale viewer on screen.
        // Collection stays at the default: dropping the entry immediately would
        // race SSR dehydration.
        staleTime: 0,
      }),
    /** Identity-changing commands call this after success; active readers refetch `/me`. */
    refresh: (client: QueryClient) =>
      client.invalidateQueries({ queryKey: sessionQueryKey, exact: true }),
    unwrap,
  };
}
