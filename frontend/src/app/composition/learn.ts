import type { QueryClient } from '@tanstack/react-query';
import { createIsomorphicFn } from '@tanstack/react-start';
import { getRequest, getRequestHeader } from '@tanstack/react-start/server';
import { createApiPlaybackRepository } from '../../adapters/api/playback-repository';
import { createApiDiscussionRepository } from '../../adapters/api/discussion-repository';
import { createApiTranscriptRepository } from '../../adapters/api/transcript-repository';
import { createApiDocumentRepository } from '../../adapters/api/document-repository';
import { createPlaybackMedia } from '../../adapters/playback/hls-media';
import { primeAv1Probe, takeAv1ProbeResult } from './av1-probe';
import { createLearnQueries } from './learn-queries';
import { createTranscriptQueries } from './transcript-queries';
import { createDocumentQueries } from './document-queries';
import { createDiscussionQueries } from './discussion-queries';
import { createApiPurchaseRepository } from '../../adapters/api/purchase-repository';
import { createPurchaseQueries } from './purchase-queries';
import { createApiCourseRepository } from '../../adapters/api/course-repository';

const requestOptions = createIsomorphicFn()
  .server(() => ({ cookie: getRequestHeader('cookie') ?? '', requestSignal: getRequest().signal }))
  .client(() => ({}));

export function createLearnDependencies(queryClient: QueryClient) {
  return {
    ...createLearnQueries(
      createApiPlaybackRepository(requestOptions(), () => takeAv1ProbeResult(queryClient)),
      createApiCourseRepository(requestOptions()),
    ),
    ...createTranscriptQueries(createApiTranscriptRepository(requestOptions())),
    ...createDocumentQueries(createApiDocumentRepository(requestOptions())),
    discussion: createDiscussionQueries(createApiDiscussionRepository(requestOptions())),
    purchase: createPurchaseQueries(createApiPurchaseRepository(requestOptions())),
    createMedia: createPlaybackMedia,
    primeAv1Probe: () => primeAv1Probe(queryClient),
  };
}

export type LearnDependencies = ReturnType<typeof createLearnDependencies>;
