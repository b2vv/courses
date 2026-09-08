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
  expect(mediaFailureDirective(403, false, 0, 'initial', false)).toEqual({
    kind: 'denied',
    refresh: true,
  });
  expect(mediaFailureDirective(403, false, 0, 'expired', true)).toEqual({ kind: 'ignore' });
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
  expect(mintFailureDirective('playback_forbidden', 0)).toEqual({
    kind: 'denied',
    refresh: true,
  });
});
