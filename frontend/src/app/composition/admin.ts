import { createIsomorphicFn } from '@tanstack/react-start';
import { getRequest, getRequestHeader } from '@tanstack/react-start/server';
import { createApiAdminCourseRepository } from '../../adapters/api/admin-course-repository';
import { createApiAdminUserRepository } from '../../adapters/api/admin-user-repository';
import { createAuthoring } from '../../application/admin/authoring';
import { createUserAdministration } from '../../application/admin/users';
import { createAdminQueries } from './admin-queries';
import { createAdminUserQueries } from './admin-user-queries';

const requestOptions = createIsomorphicFn()
  .server(() => ({ cookie: getRequestHeader('cookie') ?? '', requestSignal: getRequest().signal }))
  .client(() => ({}));

export function createAdminDependencies() {
  const options = requestOptions();
  const courses = createAdminQueries(createAuthoring(createApiAdminCourseRepository(options)));
  const users = createAdminUserQueries(
    createUserAdministration(createApiAdminUserRepository(options)),
  );
  return { ...courses, users };
}

export type AdminDependencies = ReturnType<typeof createAdminDependencies>;
