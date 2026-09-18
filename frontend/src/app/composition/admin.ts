import { createIsomorphicFn } from '@tanstack/react-start';
import { getRequest, getRequestHeader } from '@tanstack/react-start/server';
import { createApiAdminCourseRepository } from '../../adapters/api/admin-course-repository';
import { createApiAdminUserRepository } from '../../adapters/api/admin-user-repository';
import { createApiLessonVideoRepository } from '../../adapters/api/lesson-video-repository';
import { createAuthoring } from '../../application/admin/authoring';
import { createLessonVideos } from '../../application/admin/lesson-video';
import { createUserAdministration } from '../../application/admin/users';
import { createAdminQueries } from './admin-queries';
import { createAdminUserQueries } from './admin-user-queries';
import { createAdminVideoQueries } from './admin-video-queries';
import { createApiTranscriptRepository } from '../../adapters/api/transcript-repository';
import { createTranscriptQueries } from './transcript-queries';

const requestOptions = createIsomorphicFn()
  .server(() => ({ cookie: getRequestHeader('cookie') ?? '', requestSignal: getRequest().signal }))
  .client(() => ({}));

export function createAdminDependencies() {
  const options = requestOptions();
  const courses = createAdminQueries(createAuthoring(createApiAdminCourseRepository(options)));
  const users = createAdminUserQueries(
    createUserAdministration(createApiAdminUserRepository(options)),
  );
  const videos = createAdminVideoQueries(
    createLessonVideos(createApiLessonVideoRepository(options)),
  );
  const transcripts = createTranscriptQueries(createApiTranscriptRepository());
  return { ...courses, users, videos, transcripts };
}

export type AdminDependencies = ReturnType<typeof createAdminDependencies>;
