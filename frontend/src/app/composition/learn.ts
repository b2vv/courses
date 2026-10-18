import type { QueryClient } from '@tanstack/react-query';
import { createIsomorphicFn } from '@tanstack/react-start';
import { getRequest, getRequestHeader } from '@tanstack/react-start/server';
import { createApiPlaybackRepository } from '../../modules/playback';
import { createApiDiscussionRepository } from '../../modules/discussion';
import { createApiTranscriptRepository } from '../../modules/transcript';
import { createPlaybackMedia } from '../../modules/playback';
import { primeAv1Probe, takeAv1ProbeResult } from './av1-probe';
import { createLearnQueries } from './learn-queries';
import { createTranscriptQueries } from './transcript-queries';
import { createApiDocumentRepository, createDocumentQueries } from '../../modules/document';
import { createDiscussionQueries } from './discussion-queries';
import { createApiPurchaseRepository } from '../../modules/purchase';
import { createPurchaseQueries } from './purchase-queries';
import { createApiCourseRepository } from '../../modules/catalog';

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
