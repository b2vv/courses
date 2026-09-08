import { createIsomorphicFn } from '@tanstack/react-start';
import { getRequest, getRequestHeader } from '@tanstack/react-start/server';
import { createApiPlaybackRepository } from '../../adapters/api/playback-repository';
import { createPlaybackMedia } from '../../adapters/playback/hls-media';
import { createLearnQueries } from './learn-queries';

const requestOptions = createIsomorphicFn()
  .server(() => ({ cookie: getRequestHeader('cookie') ?? '', requestSignal: getRequest().signal }))
  .client(() => ({}));

export function createLearnDependencies() {
  return {
    ...createLearnQueries(createApiPlaybackRepository(requestOptions())),
    createMedia: createPlaybackMedia,
  };
}

export type LearnDependencies = ReturnType<typeof createLearnDependencies>;
