import { createIsomorphicFn } from '@tanstack/react-start';
import { getRequest, getRequestHeader } from '@tanstack/react-start/server';
import { createApiAdminCourseRepository } from '../../adapters/api/admin-course-repository';
import { createApiAdminUserRepository } from '../../adapters/api/admin-user-repository';
import { createApiLessonVideoRepository } from '../../modules/media';
import { createAuthoring } from '../../application/admin/authoring';
import { createLessonVideos } from '../../modules/media';
import { createUserAdministration } from '../../application/admin/users';
import { createAdminQueries } from './admin-queries';
import { createAdminUserQueries } from './admin-user-queries';
import { createAdminVideoQueries } from './admin-video-queries';
import { createApiDraftRepository } from '../../modules/transcript';
import { createApiTranscriptRepository } from '../../modules/transcript';
import { createDraftQueries } from './admin-draft-queries';
import { createTranscriptQueries } from './transcript-queries';
import { createApiCourseGrantRepository } from '../../modules/purchase';
import { createAdminGrantQueries } from './admin-grant-queries';
import { createApiAdminStatsRepository } from '../../adapters/api/admin-stats-repository';
import { createAdminStatsQueries } from './admin-stats-queries';
import {
  createAdminDocumentQueries,
  createApiDocumentAuthoringRepository,
} from '../../modules/document';

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
  const drafts = createDraftQueries(createApiDraftRepository(options));
  const grants = createAdminGrantQueries(createApiCourseGrantRepository(options));
  const stats = createAdminStatsQueries(createApiAdminStatsRepository(options));
  const documents = createAdminDocumentQueries(createApiDocumentAuthoringRepository(options));
  return { ...courses, users, videos, transcripts, drafts, grants, stats, documents };
}

export type AdminDependencies = ReturnType<typeof createAdminDependencies>;
