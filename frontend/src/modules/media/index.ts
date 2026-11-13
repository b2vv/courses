// Public surface of the `media` feature module (lesson-video upload and transcode status).
// Everything another module, the composition roots or the routes may use is re-exported here;
// nothing outside this folder imports the module's internals directly (enforced by
// scripts/check-architecture.mjs — no deep cross-module import). Playback-side media (HLS delivery,
// PlaybackMedia) is a separate concern and stays with the playback module.

// Domain: video states, upload-ticket / transfer types, limits and pure helpers.
export * from './domain/video';
// Domain: audio tracks of a video asset (multi-audio).
export * from './domain/audio-track';

// Application: the lesson-video query/mutation factory.
export { createLessonVideos } from './application/lesson-video';
export { createAudioTracks } from './application/audio-tracks';

// Ports: the I/O contract this feature depends on.
export type { LessonVideoRepository } from './ports/LessonVideoRepository';
export type { AudioTrackRepository } from './ports/AudioTrackRepository';

// Adapters: the HTTP implementation wired by the composition roots.
export { createApiLessonVideoRepository } from './adapters/lesson-video-repository';
export { createApiAudioTrackRepository } from './adapters/audio-track-repository';

// UI.
export { LessonVideoField } from './ui/LessonVideoField';
export type {
  CourseEditorVideo,
  LessonVideoFieldActionResult,
  LessonVideoFieldActions,
} from './ui/LessonVideoField';
