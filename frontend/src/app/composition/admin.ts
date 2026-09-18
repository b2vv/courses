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
import { createTranscriptHistory } from './transcript-history';
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
  // Created here, never at module scope: SSR renders many requests in one process, and a shared
  // store would carry one author's unsaved lines into another author's page.
  const transcriptHistory = createTranscriptHistory();
  return { ...courses, users, videos, transcripts, transcriptHistory };
}

export type AdminDependencies = ReturnType<typeof createAdminDependencies>;
