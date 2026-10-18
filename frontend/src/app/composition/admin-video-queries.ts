import { mutationOptions, queryOptions } from '@tanstack/react-query';
import type { QueryClient } from '@tanstack/react-query';
import type { createLessonVideos } from '../../modules/media';
import type {
  LessonVideo,
  VideoError,
  VideoResult,
  VideoTransferTarget,
  VideoTransferProgress,
} from '../../modules/media';
import { lessonDraftQueryKey } from './admin-draft-queries';

export const videoPollMs = 3_000;
export const adminVideoListKey = ['admin', 'videos', 'list'] as const;
const readyGenerationSnapshotKey = (courseId: string) =>
  [...adminVideoListKey, courseId, 'ready-generations'] as const;

export class VideoQueryError extends Error {
  readonly code: VideoError['code'];
  constructor(error: VideoError) {
    super(error.code);
    this.name = 'VideoQueryError';
    this.code = error.code;
  }
}

function unwrap<T>(result: VideoResult<T>): T {
  if (!result.ok) throw new VideoQueryError(result.error);
  return result.value;
}

export function adminCourseVideosQueryKey(courseId: string) {
  return [...adminVideoListKey, courseId] as const;
}

export function videoRefetchInterval(
  rows: readonly LessonVideo[] | undefined,
  transferInFlightInThisTab: boolean,
): typeof videoPollMs | false {
  return transferInFlightInThisTab ||
    rows?.some((row) => row.state === 'queued' || row.state === 'processing') === true
    ? videoPollMs
    : false;
}

export async function invalidateAfterVideoWrite(client: QueryClient, courseId: string) {
  await client.invalidateQueries({ queryKey: adminCourseVideosQueryKey(courseId), exact: true });
}

export type MintVideoVariables = {
  readonly courseId: string;
  readonly lessonId: string;
  readonly declaredBytes: number;
  readonly originalName: string;
};
export type VideoWriteVariables = { readonly courseId: string; readonly videoId: string };
export type TransferVideoVariables = {
  readonly courseId: string;
  readonly target: VideoTransferTarget;
  readonly file: File;
  readonly onProgress: (progress: VideoTransferProgress) => void;
  readonly signal?: AbortSignal;
};

export function createAdminVideoQueries(videos: ReturnType<typeof createLessonVideos>) {
  return {
    videos,
    list: (courseId: string, transferInFlightInThisTab = false) =>
      queryOptions({
        queryKey: adminCourseVideosQueryKey(courseId),
        queryFn: async ({ client, signal }) => {
          const previous = client.getQueryData<readonly LessonVideo[]>(
            adminCourseVideosQueryKey(courseId),
          );
          const next = unwrap(await videos.listByCourse(courseId, signal));
          const snapshotKey = readyGenerationSnapshotKey(courseId);
          const previousReady =
            client.getQueryData<Readonly<Record<string, number>>>(snapshotKey) ??
            Object.fromEntries(
              (previous ?? [])
                .filter((row) => row.state === 'ready')
                .map((row) => [row.lessonId, row.generation]),
            );
          const nextReady = { ...previousReady };
          for (const row of next) {
            if (row.state !== 'ready') continue;
            const previousGeneration = previousReady[row.lessonId];
            if (previousGeneration !== undefined && previousGeneration !== row.generation) {
              await client.invalidateQueries({
                queryKey: lessonDraftQueryKey(row.lessonId),
                exact: true,
              });
            }
            nextReady[row.lessonId] = Math.max(nextReady[row.lessonId] ?? 0, row.generation);
          }
          client.setQueryData(snapshotKey, nextReady);
          return next;
        },
        refetchInterval: (query) =>
          videoRefetchInterval(query.state.data, transferInFlightInThisTab),
      }),
    mint: (client: QueryClient) =>
      mutationOptions({
        mutationFn: async ({ lessonId, declaredBytes, originalName }: MintVideoVariables) =>
          unwrap(await videos.mint(lessonId, declaredBytes, originalName)),
        onSuccess: (_ticket, { courseId }) => invalidateAfterVideoWrite(client, courseId),
      }),
    retry: (client: QueryClient) =>
      mutationOptions({
        mutationFn: async ({ videoId }: VideoWriteVariables) => unwrap(await videos.retry(videoId)),
        onSuccess: (_video, { courseId }) => invalidateAfterVideoWrite(client, courseId),
      }),
    transfer: (client: QueryClient) =>
      mutationOptions({
        mutationFn: async ({ target, file, onProgress, signal }: TransferVideoVariables) =>
          unwrap(await videos.transfer(target, file, onProgress, signal)),
        onSettled: (_value, _error, { courseId }) => invalidateAfterVideoWrite(client, courseId),
      }),
    cancelUpload: (client: QueryClient) =>
      mutationOptions({
        mutationFn: async ({ videoId }: VideoWriteVariables) =>
          unwrap(await videos.cancelUpload(videoId)),
        onSuccess: (_value, { courseId }) => invalidateAfterVideoWrite(client, courseId),
      }),
    unwrap,
  };
}
