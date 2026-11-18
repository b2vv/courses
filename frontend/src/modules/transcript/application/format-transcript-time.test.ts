import { expect, test } from '@rstest/core';
import { formatTranscriptTime } from './format-transcript-time';

test('Ux07 readout rounds integer milliseconds to deciseconds across minute/hour boundaries', () => {
  for (const [ms, expected] of [
    [0, '0:00.0'],
    [49, '0:00.0'],
    [50, '0:00.1'],
    [999, '0:01.0'],
    [1000, '0:01.0'],
    [59_949, '0:59.9'],
    [59_999, '1:00.0'],
    [60_000, '1:00.0'],
    [3_599_949, '59:59.9'],
    [3_599_950, '1:00:00.0'],
    [3_600_000, '1:00:00.0'],
    [36_123_456, '10:02:03.5'],
  ] as const) {
    expect(formatTranscriptTime(ms), `readout for ${ms} ms`).toBe(expected);
  }
  for (const ms of [-1, 1.5, NaN, Infinity, -Infinity]) {
    expect(formatTranscriptTime(ms)).toBeUndefined();
  }
});
