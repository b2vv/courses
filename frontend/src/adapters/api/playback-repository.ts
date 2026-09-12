import type {
  Av1ProbeResult,
  LearningCourse,
  PlaybackErrorCode,
  PlaybackHeartbeat,
  PlaybackResult,
  PlaybackSession,
} from '../../domain/playback/playback';
import type { PlaybackRepository } from '../../ports/PlaybackRepository';
import { createPlaybackSession, getLearningCourse, heartbeatPlayback } from './generated/courses';
import {
  learningCourseSchema,
  playbackErrorSchema,
  playbackSessionSchema,
} from './playback-schema';
import { createApiTransport, unnamedStatusCode } from './transport';
import type { TransportOptions, WireResponse } from './transport';

function failure(code: PlaybackErrorCode): PlaybackResult<never> {
  return { ok: false, error: { code } };
}

function namedError(
  response: WireResponse,
  allowed: Readonly<Record<number, readonly PlaybackErrorCode[]>>,
): PlaybackErrorCode {
  const parsed = playbackErrorSchema.safeParse(response.data);
  if (parsed.success) {
    const code = allowed[response.status]?.find(
      (candidate) => candidate === parsed.data.error.code,
    );
    if (code !== undefined) return code;
  }
  return unnamedStatusCode(response.status);
}

/** One repository per browser/SSR request scope; generated transport never escapes this adapter. */
export function createApiPlaybackRepository(
  options: TransportOptions = {},
  takeAv1ProbeResult: () => Av1ProbeResult | undefined = () => undefined,
): PlaybackRepository {
  const send = createApiTransport(options);
  return {
    async getCourse(slug, signal) {
      const outcome = await send(
        'learn.course',
        (init) => getLearningCourse(encodeURIComponent(slug), init),
        200,
        signal,
      );
      if (!outcome.ok) return failure(outcome.code);
      if (outcome.status !== 200)
        return failure(
          namedError(outcome, {
            400: ['invalid_slug', 'invalid_request'],
            401: ['authentication_required'],
            403: ['course_forbidden'],
            404: ['course_not_found'],
          }),
        );
      const parsed = learningCourseSchema.safeParse(outcome.data);
      if (!parsed.success || parsed.data.slug !== slug) return failure('invalid_response');
      const course: LearningCourse = {
        id: parsed.data.id,
        slug: parsed.data.slug,
        title: parsed.data.title,
        summary: parsed.data.summary,
        description: parsed.data.description,
        status: parsed.data.status,
        modules: parsed.data.modules.map((module) => ({
          id: module.id,
          title: module.title,
          lessons: module.lessons.map((lesson) => ({
            id: lesson.id,
            title: lesson.title,
            videoState: lesson.videoState,
            position: lesson.position,
            completed: lesson.completed,
            clientSeq: lesson.clientSeq,
          })),
        })),
        continueLessonId: parsed.data.continueLessonId,
      };
      return { ok: true, value: course };
    },
    async play(lessonId, signal) {
      const av1Support = takeAv1ProbeResult();
      const outcome = await send(
        'playback.play',
        (init) =>
          createPlaybackSession(
            av1Support === undefined ? { lessonId } : { lessonId, av1Support },
            init,
          ),
        201,
        signal,
      );
      if (!outcome.ok) return failure(outcome.code);
      if (outcome.status !== 201)
        return failure(
          namedError(outcome, {
            400: ['invalid_request'],
            401: ['authentication_required'],
            403: ['playback_forbidden'],
            404: ['lesson_not_found', 'video_not_ready'],
          }),
        );
      const parsed = playbackSessionSchema.safeParse(outcome.data);
      if (!parsed.success) return failure('invalid_response');
      const session: PlaybackSession = {
        playbackId: parsed.data.playbackId,
        mediaPath: parsed.data.mediaPath,
        clientSeq: parsed.data.clientSeq,
        heartbeatSeconds: parsed.data.heartbeatSeconds,
        ownershipTtlSeconds: parsed.data.ownershipTtlSeconds,
      };
      return { ok: true, value: session };
    },
    async heartbeat(heartbeat: PlaybackHeartbeat, signal) {
      const outcome = await send(
        'playback.heartbeat',
        (init) => heartbeatPlayback(heartbeat, init),
        204,
        signal,
      );
      if (!outcome.ok) return failure(outcome.code);
      if (outcome.status === 204) return { ok: true, value: undefined };
      return failure(
        namedError(outcome, {
          400: ['invalid_request'],
          401: ['authentication_required'],
          403: ['playback_forbidden'],
          404: ['lesson_not_found'],
          409: ['playback_taken_over'],
          410: ['playback_session_expired'],
        }),
      );
    },
  };
}
