import { expect, test } from '@rstest/core';
import {
  heartbeatFailureDirective,
  mediaFailureDirective,
  mintFailureDirective,
  nextHeartbeatDelayMs,
} from './recovery';

test('maps ownership outcomes without allowing a terminal refusal to remint', () => {
  expect(heartbeatFailureDirective('playback_session_expired', 1_000)).toEqual({ kind: 'remint' });
  expect(heartbeatFailureDirective('playback_taken_over', 1_000)).toEqual({
    kind: 'taken-over',
  });
  expect(heartbeatFailureDirective('playback_forbidden', 1_000)).toEqual({
    kind: 'denied',
    refresh: true,
  });
  expect(heartbeatFailureDirective('dependency_unavailable', 1_000)).toEqual({
    kind: 'retry',
    delayMs: 2_000,
  });
  expect(nextHeartbeatDelayMs(30, 1_000)).toBe(30_000);
  expect(nextHeartbeatDelayMs(30, 4_000)).toBe(4_000);
});

test('maps exact media statuses and bounds consecutive retries', () => {
  // A refused segment no longer decides on its own. The edge answers 403 the same way to an
  // expired session, an evicted one and a revoked course (A7, pinned by R29), so the player asks
  // the heartbeat, which is the only place those three differ.
  expect(mediaFailureDirective(403, false, 0, 'initial', false)).toEqual({
    kind: 'await-verdict',
  });
  expect(mediaFailureDirective(403, false, 0, 'expired', true)).toEqual({ kind: 'ignore' });
  // A 403 from a prefix the player has already replaced is ignored whatever the mint is doing.
  // Keying this on "a mint is in flight" was not enough: that flag clears the moment the mint
  // resolves, while a request already on the wire from the old prefix lands afterwards - and the
  // student, whose access is intact, was shown that the course is no longer available.
  const oldPrefix = '/media/p/aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa/hls/v/master.m3u8';
  const currentPrefix = '/media/p/bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb/hls/v/master.m3u8';
  expect(mediaFailureDirective(403, false, 0, 'initial', false, oldPrefix, currentPrefix)).toEqual({
    kind: 'ignore',
  });
  // The same 403 on the prefix in use is a genuine denial and must still stop playback.
  expect(
    mediaFailureDirective(403, false, 0, 'initial', false, currentPrefix, currentPrefix),
  ).toEqual({ kind: 'await-verdict' });
  expect(mediaFailureDirective(404, false, 0, 'initial', false)).toEqual({
    kind: 'reload-course',
    nextRetryCount: 1,
  });
  expect(mediaFailureDirective(404, false, 1, 'missing', false)).toEqual({
    kind: 'unavailable',
    refresh: true,
  });
  expect(mediaFailureDirective(503, true, 2, 'initial', false)).toEqual({
    kind: 'retry-media',
    delayMs: 2_000,
    nextRetryCount: 3,
  });
  expect(mediaFailureDirective(503, true, 3, 'initial', false)).toEqual({
    kind: 'unavailable',
    refresh: true,
  });
  expect(mediaFailureDirective(undefined, false, 0, 'initial', false)).toEqual({
    kind: 'unavailable',
    refresh: false,
  });
});

test('retries mint dependency failures with a bounded exponential delay', () => {
  expect(mintFailureDirective('dependency_unavailable', 0)).toEqual({
    kind: 'retry',
    delayMs: 500,
    nextRetryCount: 1,
  });
  expect(mintFailureDirective('request_timeout', 2)).toEqual({
    kind: 'retry',
    delayMs: 2_000,
    nextRetryCount: 3,
  });
  expect(mintFailureDirective('dependency_unavailable', 3)).toEqual({
    kind: 'unavailable',
    refresh: false,
  });
  for (const code of [
    'playback_forbidden',
    'course_purchase_required',
    'email_not_verified',
  ] as const) {
    expect(mintFailureDirective(code, 0)).toEqual({ kind: 'denied', refresh: true });
    expect(heartbeatFailureDirective(code, 1_000)).toEqual({ kind: 'denied', refresh: true });
  }
});
