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

// `/media/p/<32 hex>/...` - the playback id the bytes were requested under. Undefined for a URL
// that carries none, including the native branch, which never learns which request failed.
function playbackPrefix(url: string | undefined): string | undefined {
  if (url === undefined) return undefined;
  return /\/media\/p\/([0-9a-f]{32})\//.exec(url)?.[1];
}

export function mediaFailureDirective(
  status: number | undefined,
  retryable: boolean,
  retryCount: number,
  reloadMode: PlaybackReloadMode,
  mintInFlight: boolean,
  failedUrl?: string,
  currentMediaPath?: string,
): MediaFailureDirective {
  if (status === 403) {
    // A 403 says the course is closed to this student only when it came from the prefix the
    // player is using now. A request issued on a prefix that has since been replaced answers
    // about a session that no longer exists, and treating it as denial shows "the course is no
    // longer available" to someone whose access is intact.
    //
    // `mintInFlight` alone was the wrong key. It goes false the moment the mint resolves, while
    // a request already on the wire from the old prefix can land afterwards - which is exactly
    // what the O1 measurement caught, by holding one in flight on purpose.
    const failedPrefix = playbackPrefix(failedUrl);
    const currentPrefix = playbackPrefix(currentMediaPath);
    if (failedPrefix !== undefined && currentPrefix !== undefined && failedPrefix !== currentPrefix)
      return { kind: 'ignore' };
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
