import type { PlaybackErrorCode } from '../../domain/playback/playback';

export const initialHeartbeatBackoffMs = 1_000;
export const maximumPlaybackRetries = 3;

export type PlaybackReloadMode = 'initial' | 'expired' | 'missing';

export type HeartbeatFailureDirective =
  | { readonly kind: 'remint' }
  | { readonly kind: 'taken-over' }
  | { readonly kind: 'denied'; readonly refresh: true }
  | { readonly kind: 'unavailable'; readonly refresh: boolean }
  | { readonly kind: 'retry'; readonly delayMs: number };

export function heartbeatFailureDirective(
  code: PlaybackErrorCode,
  currentBackoffMs: number,
): HeartbeatFailureDirective {
  switch (code) {
    case 'playback_session_expired':
      return { kind: 'remint' };
    case 'playback_taken_over':
      return { kind: 'taken-over' };
    case 'playback_forbidden':
      return { kind: 'denied', refresh: true };
    case 'lesson_not_found':
      return { kind: 'unavailable', refresh: true };
    case 'dependency_unavailable':
    case 'request_timeout':
      return { kind: 'retry', delayMs: Math.min(currentBackoffMs * 2, 8_000) };
    default:
      return { kind: 'unavailable', refresh: false };
  }
}

export function nextHeartbeatDelayMs(heartbeatSeconds: number, backoffMs: number): number {
  return backoffMs === initialHeartbeatBackoffMs ? heartbeatSeconds * 1_000 : backoffMs;
}

export type MintFailureDirective =
  | { readonly kind: 'denied'; readonly refresh: true }
  | { readonly kind: 'retry'; readonly delayMs: number; readonly nextRetryCount: number }
  | { readonly kind: 'unavailable'; readonly refresh: boolean };

export function mintFailureDirective(
  code: PlaybackErrorCode,
  retryCount: number,
): MintFailureDirective {
  if (code === 'playback_forbidden') return { kind: 'denied', refresh: true };
  if (code === 'dependency_unavailable' || code === 'request_timeout') {
    return retryCount >= maximumPlaybackRetries
      ? { kind: 'unavailable', refresh: false }
      : {
          kind: 'retry',
          delayMs: 500 * 2 ** retryCount,
          nextRetryCount: retryCount + 1,
        };
  }
  return {
    kind: 'unavailable',
    refresh: code === 'lesson_not_found' || code === 'video_not_ready',
  };
}

export type MediaFailureDirective =
  | { readonly kind: 'ignore' }
  | { readonly kind: 'denied'; readonly refresh: true }
  | { readonly kind: 'reload-course'; readonly nextRetryCount: number }
  | { readonly kind: 'retry-media'; readonly delayMs: number; readonly nextRetryCount: number }
  | { readonly kind: 'unavailable'; readonly refresh: boolean };

export function mediaFailureDirective(
  status: number | undefined,
  retryable: boolean,
  retryCount: number,
  reloadMode: PlaybackReloadMode,
  mintInFlight: boolean,
): MediaFailureDirective {
  if (status === 403) {
    return reloadMode === 'expired' && mintInFlight
      ? { kind: 'ignore' }
      : { kind: 'denied', refresh: true };
  }
  if (status === 404) {
    return retryCount > 0
      ? { kind: 'unavailable', refresh: true }
      : { kind: 'reload-course', nextRetryCount: 1 };
  }
  if (status === 503 || (status === undefined && retryable)) {
    return retryCount >= maximumPlaybackRetries
      ? { kind: 'unavailable', refresh: true }
      : {
          kind: 'retry-media',
          delayMs: 500 * 2 ** retryCount,
          nextRetryCount: retryCount + 1,
        };
  }
  return { kind: 'unavailable', refresh: false };
}
