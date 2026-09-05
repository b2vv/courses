import { createIsomorphicFn } from '@tanstack/react-start';
import { getCookie, getRequest, getRequestHeader } from '@tanstack/react-start/server';
import { createSession } from '../../application/identity/session';
import { createApiSessionRepository } from '../../adapters/api/session-repository';
import { createSessionQueries } from './identity-queries';

const sessionCookieName = 'courseforge_session';
const identitySsrTimeoutMs = 1_000;

const requestOptions = createIsomorphicFn()
  .server(() => ({
    shouldLoadCurrentUser: getCookie(sessionCookieName) !== undefined,
    transport: {
      cookie: getRequestHeader('cookie') ?? '',
      requestSignal: getRequest().signal,
      timeoutMs: identitySsrTimeoutMs,
    },
  }))
  .client(() => ({ shouldLoadCurrentUser: true, transport: {} }));

export function createIdentityDependencies() {
  const request = requestOptions();
  return {
    ...createSessionQueries(createSession(createApiSessionRepository(request.transport))),
    shouldLoadCurrentUser: request.shouldLoadCurrentUser,
  };
}

export type IdentityDependencies = ReturnType<typeof createIdentityDependencies>;
