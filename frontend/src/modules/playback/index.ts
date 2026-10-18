// Public surface of the `playback` feature module (HLS delivery, playback sessions, heartbeats and
// the lesson player). Everything another module, the composition roots or the routes may use is
// re-exported here; nothing outside this folder imports the module's internals directly (enforced by
// scripts/check-architecture.mjs — no deep cross-module import).

// Domain: learning-course tree, playback session/heartbeat types, codes and pure helpers.
export * from './domain/playback';

// Application: heartbeat/mint/media failure directives and backoff maths.
export * from './application/recovery';

// Ports: the I/O contracts this feature depends on.
export type {
  CreatePlaybackMedia,
  PlaybackMedia,
  PlaybackMediaCallbacks,
  PlaybackMediaError,
  PlaybackQuality,
} from './ports/PlaybackMedia';
export type { PlaybackRepository } from './ports/PlaybackRepository';

// Adapters: the HTTP / HLS implementations wired by the composition roots.
export { createApiPlaybackRepository } from './adapters/playback-repository';
export { createPlaybackMedia } from './adapters/hls-media';

// UI.
export { LessonPlayer } from './ui/LessonPlayer';
export type { LessonPlayerActions } from './ui/LessonPlayer';
