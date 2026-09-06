import type {
  LessonVideo,
  VideoError,
  VideoResult,
  VideoUploadTicket,
} from '../../domain/media/video';
import type { LessonVideoRepository } from '../../ports/LessonVideoRepository';
import { listAdminCourseVideos, mintAdminVideoTicket, retryAdminVideo } from './generated/courses';
import { errorSchema } from './course-schema';
import {
  courseVideosSchema,
  lessonVideoSchema,
  videoUploadTicketSchema,
} from './lesson-video-schema';
import { createApiTransport, unnamedStatusCode } from './transport';
import type { TransportOptions, WireResponse } from './transport';
import { createTusClient } from './tus-client';

const invalid = {
  ok: false,
  error: { code: 'invalid_response' },
} satisfies VideoResult<never>;

function wireError(response: WireResponse): VideoError {
  const parsed = errorSchema.safeParse(response.data);
  if (parsed.success) {
    switch (parsed.data.error.code) {
      case 'invalid_request':
      case 'unsupported_media':
      case 'upload_too_large':
        if (response.status === 400) return { code: parsed.data.error.code };
        break;
      case 'authentication_required':
        if (response.status === 401) return { code: 'authentication_required' };
        break;
      case 'origin_forbidden':
      case 'forbidden':
      case 'course_forbidden':
        if (response.status === 403) return { code: parsed.data.error.code };
        break;
      case 'lesson_not_found':
      case 'video_not_found':
        if (response.status === 404) return { code: parsed.data.error.code };
        break;
      case 'upload_in_progress':
      case 'video_not_retryable':
        if (response.status === 409) return { code: parsed.data.error.code };
        break;
      case 'dependency_unavailable':
        if (response.status === 503) return { code: 'dependency_unavailable' };
        break;
      case 'internal_error':
        if (response.status === 500) return { code: 'dependency_unavailable' };
        break;
    }
  }
  return { code: unnamedStatusCode(response.status) };
}

function video(raw: unknown, expectedId?: string): VideoResult<LessonVideo> {
  const parsed = lessonVideoSchema.safeParse(raw);
  if (!parsed.success || (expectedId !== undefined && parsed.data.videoId !== expectedId))
    return invalid;
  const { failureReason, ...fields } = parsed.data;
  return {
    ok: true,
    value: failureReason === undefined ? fields : { ...fields, failureReason },
  };
}

/** Generated JSON calls and the hand-written tus transport meet at this adapter boundary. */
export function createApiLessonVideoRepository(
  options: TransportOptions = {},
): LessonVideoRepository {
  const send = createApiTransport(options);
  const tus = createTusClient(options);
  async function request<T>(
    operation: string,
    execute: (init: RequestInit) => Promise<WireResponse>,
    expected: number,
    map: (data: unknown) => VideoResult<T>,
    caller?: AbortSignal,
  ): Promise<VideoResult<T>> {
    const outcome = await send(operation, execute, expected, caller);
    if (!outcome.ok) return { ok: false, error: { code: outcome.code } };
    if (outcome.status === expected) return map(outcome.data);
    return { ok: false, error: wireError(outcome) };
  }

  return {
    mint(lessonId, declaredBytes, originalName, signal) {
      return request<VideoUploadTicket>(
        'admin.videos.mint',
        (init) => mintAdminVideoTicket(lessonId, { declaredBytes, originalName }, init),
        201,
        (raw) => {
          const parsed = videoUploadTicketSchema.safeParse(raw);
          return parsed.success ? { ok: true, value: parsed.data } : invalid;
        },
        signal,
      );
    },
    listByCourse(courseId, signal) {
      return request<readonly LessonVideo[]>(
        'admin.videos.list',
        (init) => listAdminCourseVideos(courseId, init),
        200,
        (raw) => {
          const parsed = courseVideosSchema.safeParse(raw);
          if (!parsed.success) return invalid;
          const items: LessonVideo[] = [];
          for (const item of parsed.data.items) {
            const mapped = video(item);
            if (!mapped.ok) return invalid;
            items.push(mapped.value);
          }
          return { ok: true, value: items };
        },
        signal,
      );
    },
    retry(videoId, signal) {
      return request<LessonVideo>(
        'admin.videos.retry',
        (init) => retryAdminVideo(videoId, init),
        200,
        (raw) => video(raw, videoId),
        signal,
      );
    },
    transfer(target, file, onProgress, signal) {
      return tus.transfer(target, file, onProgress, signal);
    },
    cancelUpload(videoId, signal) {
      return tus.cancel(videoId, signal);
    },
  };
}
