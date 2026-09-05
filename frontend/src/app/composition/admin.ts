import { createIsomorphicFn } from '@tanstack/react-start';
import { getRequest, getRequestHeader } from '@tanstack/react-start/server';
import { createApiAdminCourseRepository } from '../../adapters/api/admin-course-repository';
import { createAuthoring } from '../../application/admin/authoring';
import { createAdminQueries } from './admin-queries';

const requestOptions = createIsomorphicFn()
  .server(() => ({ cookie: getRequestHeader('cookie') ?? '', requestSignal: getRequest().signal }))
  .client(() => ({}));

export function createAdminDependencies() {
  return createAdminQueries(createAuthoring(createApiAdminCourseRepository(requestOptions())));
}

export type AdminDependencies = ReturnType<typeof createAdminDependencies>;
