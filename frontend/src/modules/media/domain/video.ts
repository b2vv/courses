export const videoStates = [
  'created',
  'uploading',
  'queued',
  'processing',
  'ready',
  'failed',
] as const;
export type VideoState = (typeof videoStates)[number];

export const videoFailureReasons = [
  'unsupported_source',
  'probe_failed',
  'render_failed',
  'publish_failed',
  'storage_unavailable',
] as const;
export type VideoFailureReason = (typeof videoFailureReasons)[number];

export type LessonVideo = {
  readonly lessonId: string;
  readonly videoId: string;
  readonly state: VideoState;
  readonly generation: number;
  readonly originalName: string;
  readonly receivedBytes: number;
  readonly declaredBytes: number;
  readonly updatedAt: string;
  readonly failureReason?: VideoFailureReason;
};

export type VideoUploadTicket = {
  readonly videoId: string;
  readonly uploadUrl: string;
  readonly chunkBytes: number;
  readonly maxBytes: number;
  readonly expiresAt: string;
};

export type VideoTransferTarget =
  | { readonly kind: 'new'; readonly ticket: VideoUploadTicket }
  | { readonly kind: 'resume'; readonly videoId: string; readonly declaredBytes: number };

export type VideoTransferProgress = {
  readonly receivedBytes: number;
  readonly declaredBytes: number;
};

export type VideoErrorCode =
  | 'invalid_request'
  | 'authentication_required'
  | 'origin_forbidden'
  | 'forbidden'
  | 'course_forbidden'
  | 'lesson_not_found'
  | 'upload_in_progress'
  | 'unsupported_media'
  | 'upload_too_large'
  | 'video_not_found'
  | 'video_not_retryable'
  | 'upload_resume_unverified'
  | 'dependency_unavailable'
  | 'invalid_response'
  | 'request_timeout'
  | 'request_cancelled';
export type VideoError = { readonly code: VideoErrorCode };
export type VideoResult<T> =
  | { readonly ok: true; readonly value: T }
  | { readonly ok: false; readonly error: VideoError };

export const maximumVideoBytes = 2 * 1024 * 1024 * 1024;
export const videoChunkBytes = 8 * 1024 * 1024;
export const acceptedVideoFileTypes = ['.mp4', '.mov', '.mkv'] as const;

const uuidPattern = /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/;

export function validMediaId(value: string): boolean {
  return uuidPattern.test(value) && value !== '00000000-0000-0000-0000-000000000000';
}

export function validVideoTicketInput(declaredBytes: number, originalName: string): boolean {
  return (
    Number.isSafeInteger(declaredBytes) &&
    declaredBytes >= 1 &&
    declaredBytes <= maximumVideoBytes &&
    [...originalName].length >= 1 &&
    [...originalName].length <= 255 &&
    !originalName.includes('\u0000')
  );
}
