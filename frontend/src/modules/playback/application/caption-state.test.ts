import { expect, test } from '@rstest/core';
import { captionQuery, reduceCaptionSelection, type CaptionSelectionEvent } from './caption-state';

const mount = Symbol('A');
const other = Symbol('B');
const input = {
  locked: false,
  emailGated: false,
  ready: true,
  mount,
  record: null,
  predictedLanguage: 'EN',
};
const event = (
  generation: number,
  selection: CaptionSelectionEvent['selection'],
  identity = mount,
): CaptionSelectionEvent => ({ mount: identity, generation, selection });

test.each([
  ['locked', { locked: true }, { mode: 'off', panel: 'hidden' }],
  ['email gate', { emailGated: true }, { mode: 'off', panel: 'hidden' }],
  ['non-ready', { ready: false }, { mode: 'primary' }],
  ['before Play', {}, { mode: 'exact', language: 'en' }],
  [
    'old mount is prediction',
    { record: event(2, { state: 'confirmed', language: 'fr' }, other) },
    { mode: 'exact', language: 'en' },
  ],
  [
    'confirmed',
    { record: event(1, { state: 'confirmed', language: 'DE' }) },
    { mode: 'exact', language: 'de' },
  ],
  [
    'pending after load',
    { record: event(1, { state: 'pending' }) },
    { mode: 'off', panel: 'loading' },
  ],
  [
    'unavailable',
    { record: event(1, { state: 'unavailable' }) },
    { mode: 'off', panel: 'none-for-track' },
  ],
  ['und prediction', { predictedLanguage: 'UND' }, { mode: 'off', panel: 'none-for-track' }],
  [
    'und family prediction',
    { predictedLanguage: 'uNd-x-original' },
    { mode: 'off', panel: 'none-for-track' },
  ],
  [
    'und confirmation',
    { record: event(1, { state: 'confirmed', language: 'UND-US' }) },
    { mode: 'off', panel: 'none-for-track' },
  ],
  ['no predicted track', { predictedLanguage: null }, { mode: 'off', panel: 'loading' }],
] as const)('M4/M5 %s', (_name, overrides, expected) => {
  expect(captionQuery({ ...input, ...overrides })).toEqual(expected);
});

test('drops foreign mounts, accepts A gen 2 → B gen 1, drops older generations within B', () => {
  const a = event(2, { state: 'confirmed', language: 'en' });
  const b = event(1, { state: 'pending' }, other);
  expect(reduceCaptionSelection(mount, a, b)).toBe(a);
  expect(reduceCaptionSelection(other, a, b)).toBe(b);
  const confirmed = event(2, { state: 'confirmed', language: 'de' }, other);
  expect(reduceCaptionSelection(other, b, confirmed)).toBe(confirmed);
  expect(reduceCaptionSelection(other, confirmed, b)).toBe(confirmed);
  expect(reduceCaptionSelection(mount, null, b)).toBeNull();
  expect(reduceCaptionSelection(mount, null, a)).toBe(a);
});

test('once loaded, retry pending and unavailable never return to prediction', () => {
  let record: CaptionSelectionEvent | null = null;
  for (const next of [
    event(1, { state: 'pending' }),
    event(1, { state: 'confirmed', language: 'de' }),
    event(2, { state: 'pending' }),
    event(2, { state: 'unavailable' }),
  ]) {
    record = reduceCaptionSelection(mount, record, next);
    const query = captionQuery({ ...input, record });
    expect(query).not.toEqual({ mode: 'exact', language: 'en' });
  }
  // A→B→A has a fresh symbol, so the old A record cannot suppress prediction.
  expect(captionQuery({ ...input, mount: Symbol('A'), record })).toEqual({
    mode: 'exact',
    language: 'en',
  });
});
