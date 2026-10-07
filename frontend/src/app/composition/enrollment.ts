import { createIsomorphicFn } from '@tanstack/react-start';
import { getRequest, getRequestHeader } from '@tanstack/react-start/server';
import { createApiEnrollmentRepository } from '../../adapters/api/enrollment-repository';
import { createEnrollmentQueries } from './enrollment-queries';

const requestOptions = createIsomorphicFn()
  .server(() => ({ cookie: getRequestHeader('cookie') ?? '', requestSignal: getRequest().signal }))
  .client(() => ({}));

export function createEnrollmentDependencies() {
  return createEnrollmentQueries(createApiEnrollmentRepository(requestOptions()));
}
