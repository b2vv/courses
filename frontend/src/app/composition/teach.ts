import { createIsomorphicFn } from '@tanstack/react-start';
import { getRequest, getRequestHeader } from '@tanstack/react-start/server';
import { createApiTeachRepository } from '../../adapters/api/teach-repository';
import { createTeachQueries } from './teach-queries';

const requestOptions = createIsomorphicFn()
  .server(() => ({ cookie: getRequestHeader('cookie') ?? '', requestSignal: getRequest().signal }))
  .client(() => ({}));

export function createTeachDependencies() {
  return createTeachQueries(createApiTeachRepository(requestOptions()));
}

export type TeachDependencies = ReturnType<typeof createTeachDependencies>;
