import {
  maximumVideoBytes,
  validMediaId,
  validVideoTicketInput,
  videoChunkBytes,
} from '../../domain/media/video';
import type {
  VideoResult,
  VideoTransferProgress,
  VideoTransferTarget,
} from '../../domain/media/video';
import type { LessonVideoRepository } from '../../ports/LessonVideoRepository';

const invalid = { ok: false, error: { code: 'invalid_request' } } as const;

function validTransferTarget(target: VideoTransferTarget): boolean {
  if (target.kind === 'resume') {
    return (
      validMediaId(target.videoId) &&
      Number.isSafeInteger(target.declaredBytes) &&
      target.declaredBytes >= 1 &&
      target.declaredBytes <= maximumVideoBytes
    );
  }
  return (
    validMediaId(target.ticket.videoId) &&
    target.ticket.uploadUrl === `/uploads/${target.ticket.videoId}` &&
    target.ticket.chunkBytes === videoChunkBytes &&
    target.ticket.maxBytes === maximumVideoBytes
  );
}

export function createLessonVideos(repository: LessonVideoRepository) {
  return {
    mint(lessonId: string, declaredBytes: number, originalName: string, signal?: AbortSignal) {
      return validMediaId(lessonId) && validVideoTicketInput(declaredBytes, originalName)
        ? repository.mint(lessonId, declaredBytes, originalName, signal)
        : Promise.resolve(invalid);
    },
    listByCourse(courseId: string, signal?: AbortSignal) {
      return validMediaId(courseId)
        ? repository.listByCourse(courseId, signal)
        : Promise.resolve(invalid);
    },
    retry(videoId: string, signal?: AbortSignal) {
      return validMediaId(videoId) ? repository.retry(videoId, signal) : Promise.resolve(invalid);
    },
    transfer(
      target: VideoTransferTarget,
      file: File,
      onProgress: (progress: VideoTransferProgress) => void,
      signal?: AbortSignal,
    ): Promise<VideoResult<null>> {
      return validTransferTarget(target) && file.size >= 1 && file.size <= maximumVideoBytes
        ? repository.transfer(target, file, onProgress, signal)
        : Promise.resolve(invalid);
    },
    cancelUpload(videoId: string, signal?: AbortSignal) {
      return validMediaId(videoId)
        ? repository.cancelUpload(videoId, signal)
        : Promise.resolve(invalid);
    },
  };
}
