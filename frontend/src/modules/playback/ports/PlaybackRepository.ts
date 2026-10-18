import type {
  LearningCourse,
  PlaybackHeartbeat,
  PlaybackResult,
  PlaybackSession,
} from '../domain/playback';

export interface PlaybackRepository {
  getCourse(slug: string, signal?: AbortSignal): Promise<PlaybackResult<LearningCourse>>;
  play(lessonId: string, signal?: AbortSignal): Promise<PlaybackResult<PlaybackSession>>;
  heartbeat(heartbeat: PlaybackHeartbeat, signal?: AbortSignal): Promise<PlaybackResult<void>>;
}
