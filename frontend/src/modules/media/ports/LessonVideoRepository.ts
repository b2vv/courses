import type {
  LessonVideo,
  VideoResult,
  VideoTransferTarget,
  VideoTransferProgress,
  VideoUploadTicket,
} from '../domain/video';

export interface LessonVideoRepository {
  mint(
    lessonId: string,
    declaredBytes: number,
    originalName: string,
    signal?: AbortSignal,
  ): Promise<VideoResult<VideoUploadTicket>>;
  listByCourse(
    courseId: string,
    signal?: AbortSignal,
  ): Promise<VideoResult<readonly LessonVideo[]>>;
  retry(videoId: string, signal?: AbortSignal): Promise<VideoResult<LessonVideo>>;
  transfer(
    target: VideoTransferTarget,
    file: File,
    onProgress: (progress: VideoTransferProgress) => void,
    signal?: AbortSignal,
  ): Promise<VideoResult<null>>;
  cancelUpload(videoId: string, signal?: AbortSignal): Promise<VideoResult<null>>;
}
