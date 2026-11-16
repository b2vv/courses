import { mutationOptions, queryOptions } from '@tanstack/react-query';
import type { QueryClient } from '@tanstack/react-query';
import { terminalAudioSignature } from '../../modules/media';
import type {
  AudioTrack,
  LessonVideo,
  AudioTrackError,
  AudioTrackResult,
  AudioTransferProgress,
  AudioTransferTarget,
  createAudioTracks,
} from '../../modules/media';
import { invalidateLessonDrafts } from './admin-draft-queries';
import {
  adminCourseVideosQueryKey,
  invalidateAfterVideoWrite,
  videoPollMs,
} from './admin-video-queries';

export const adminAudioTrackListKey = ['admin', 'audio-tracks', 'list'] as const;
const terminalSignatureKey = (videoId: string) =>
  [...adminAudioTrackListKey, videoId, 'terminal-signature'] as const;

export class AudioTrackQueryError extends Error {
  readonly code: AudioTrackError['code'];
  constructor(error: AudioTrackError) {
    super(error.code);
    this.name = 'AudioTrackQueryError';
    this.code = error.code;
  }
}

function unwrap<T>(result: AudioTrackResult<T>): T {
  if (!result.ok) throw new AudioTrackQueryError(result.error);
  return result.value;
}

export function adminAudioTracksQueryKey(videoId: string) {
  return [...adminAudioTrackListKey, videoId] as const;
}

export function audioTrackRefetchInterval(
  tracks: readonly AudioTrack[] | undefined,
  transferInFlightInThisTab: boolean,
): typeof videoPollMs | false {
  return transferInFlightInThisTab ||
    tracks?.some((track) => track.state === 'queued' || track.state === 'processing') === true
    ? videoPollMs
    : false;
}

async function invalidateAfterAudioWrite(client: QueryClient, videoId: string) {
  await client.invalidateQueries({ queryKey: adminAudioTracksQueryKey(videoId), exact: true });
}

async function invalidateAudioLessonDrafts(client: QueryClient, courseId: string, videoId: string) {
  const videos = client.getQueryData<readonly LessonVideo[]>(adminCourseVideosQueryKey(courseId));
  const lessonId = videos?.find((video) => video.videoId === videoId)?.lessonId;
  // A detached caller may not have the course video list cached; invalidate all in that case.
  await invalidateLessonDrafts(client, lessonId);
}

export type AudioTrackVariables = {
  readonly courseId: string;
  readonly videoId: string;
  readonly trackId: string;
};
export type MintAudioTrackVariables = {
  readonly courseId: string;
  readonly videoId: string;
  readonly declaredBytes: number;
  readonly originalName: string;
  readonly language: string;
  readonly label: string;
  readonly multilingual: boolean;
};
export type RenameAudioTrackVariables = AudioTrackVariables & {
  readonly language: string;
  readonly label: string;
  readonly multilingual: boolean;
};
export type TransferAudioTrackVariables = {
  readonly courseId: string;
  readonly videoId: string;
  readonly target: AudioTransferTarget;
  readonly file: File;
  readonly onProgress: (progress: AudioTransferProgress) => void;
  readonly signal?: AbortSignal;
};

export function createAdminAudioQueries(tracks: ReturnType<typeof createAudioTracks>) {
  return {
    tracks,
    /**
     * The course id is the course whose video list this asset appears in: the worker changes the
     * asset state together with a track's terminal transition, so when the terminal signature
     * differs from the last one this cache saw — including on the first read, with no snapshot —
     * that list is re-read (D-L2-3). Identical responses keep the signature and do nothing.
     */
    list: (courseId: string, videoId: string, transferInFlightInThisTab = false) =>
      queryOptions({
        queryKey: adminAudioTracksQueryKey(videoId),
        queryFn: async ({ client, signal }) => {
          const next = unwrap(await tracks.list(videoId, signal));
          const signature = terminalAudioSignature(next.items);
          const key = terminalSignatureKey(videoId);
          const previous = client.getQueryData<string>(key);
          client.setQueryData(key, signature);
          if (signature !== (previous ?? '')) {
            await client.invalidateQueries({
              queryKey: adminCourseVideosQueryKey(courseId),
              exact: true,
            });
            // A newly Ready track can now have its own draft.
            await invalidateAudioLessonDrafts(client, courseId, videoId);
          }
          return next;
        },
        refetchInterval: (query) =>
          audioTrackRefetchInterval(query.state.data?.items, transferInFlightInThisTab),
      }),
    mint: (client: QueryClient) =>
      mutationOptions({
        mutationFn: async ({ videoId, ...input }: MintAudioTrackVariables) =>
          unwrap(
            await tracks.mint(videoId, {
              declaredBytes: input.declaredBytes,
              originalName: input.originalName,
              language: input.language,
              label: input.label,
              multilingual: input.multilingual,
            }),
          ),
        onSuccess: (_ticket, { videoId }) => invalidateAfterAudioWrite(client, videoId),
      }),
    transfer: (client: QueryClient) =>
      mutationOptions({
        mutationFn: async ({ target, file, onProgress, signal }: TransferAudioTrackVariables) =>
          unwrap(await tracks.transfer(target, file, onProgress, signal)),
        onSettled: (_value, _error, { videoId }) => invalidateAfterAudioWrite(client, videoId),
      }),
    cancelUpload: (client: QueryClient) =>
      mutationOptions({
        mutationFn: async ({ trackId }: AudioTrackVariables) =>
          unwrap(await tracks.cancelUpload(trackId)),
        onSuccess: (_value, { videoId }) => invalidateAfterAudioWrite(client, videoId),
      }),
    rename: (client: QueryClient) =>
      mutationOptions({
        mutationFn: async ({
          videoId,
          trackId,
          language,
          label,
          multilingual,
        }: RenameAudioTrackVariables) =>
          unwrap(await tracks.rename(videoId, trackId, language, label, multilingual)),
        onSuccess: async (_track, { courseId, videoId }) => {
          await invalidateAfterAudioWrite(client, videoId);
          await invalidateAudioLessonDrafts(client, courseId, videoId);
        },
      }),
    setDefault: (client: QueryClient) =>
      mutationOptions({
        mutationFn: async ({ videoId, trackId }: AudioTrackVariables) =>
          unwrap(await tracks.setDefault(videoId, trackId)),
        onSuccess: async (_value, { courseId, videoId }) => {
          await invalidateAfterAudioWrite(client, videoId);
          await invalidateAfterVideoWrite(client, courseId);
          // Server order is default first; refresh the draft order.
          await invalidateAudioLessonDrafts(client, courseId, videoId);
        },
      }),
    remove: (client: QueryClient) =>
      mutationOptions({
        mutationFn: async ({ videoId, trackId }: AudioTrackVariables) =>
          unwrap(await tracks.remove(videoId, trackId)),
        onSuccess: async (_value, { courseId, videoId }) => {
          await invalidateAfterAudioWrite(client, videoId);
          await invalidateAfterVideoWrite(client, courseId);
          // Deleted tracks and their drafts disappear; a new default can be elected.
          await invalidateAudioLessonDrafts(client, courseId, videoId);
        },
      }),
    unwrap,
  };
}
