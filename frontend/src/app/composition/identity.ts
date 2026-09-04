import { createIsomorphicFn } from '@tanstack/react-start';
import { getRequest, getRequestHeader } from '@tanstack/react-start/server';
import { createSession } from '../../application/identity/session';
import { createApiSessionRepository } from '../../adapters/api/session-repository';
import { createSessionQueries } from './identity-queries';

const requestOptions = createIsomorphicFn()
  .server(() => ({ cookie: getRequestHeader('cookie') ?? '', requestSignal: getRequest().signal }))
  .client(() => ({}));

export function createIdentityDependencies() {
  return createSessionQueries(createSession(createApiSessionRepository(requestOptions())));
}
