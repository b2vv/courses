import { expect, test } from '@rstest/core';
import {
  defaultPreferences,
  parsePreferences,
  serializePreferences,
  type PlayerPreferences,
} from './preferences';

test('null (nothing stored) yields the defaults', () => {
  expect(parsePreferences(null)).toEqual(defaultPreferences);
});

test('a complete, valid blob round-trips through serialize/parse', () => {
  const prefs: PlayerPreferences = { volume: 0.4, muted: true, rate: 1.5, captionsOn: true };
  expect(parsePreferences(serializePreferences(prefs))).toEqual(prefs);
});

test('invalid JSON degrades to the defaults rather than throwing', () => {
  expect(parsePreferences('{not json')).toEqual(defaultPreferences);
  expect(parsePreferences('42')).toEqual(defaultPreferences);
  expect(parsePreferences('null')).toEqual(defaultPreferences);
});

test('each field falls back independently when its value is wrong', () => {
  // Volume out of range, muted not a boolean, rate present and valid.
  expect(parsePreferences(JSON.stringify({ volume: 5, muted: 'yes', rate: 1.25 }))).toEqual({
    volume: 1,
    muted: false,
    rate: 1.25,
    captionsOn: false,
  });
});

test('a rate the player does not offer is dropped, not clamped', () => {
  // 3× is finite but not a control the student could ever undo, so it must not survive a reload.
  expect(parsePreferences(JSON.stringify({ volume: 0.5, muted: false, rate: 3 }))).toEqual({
    volume: 0.5,
    muted: false,
    rate: 1,
    captionsOn: false,
  });
});

test('captionsOn is remembered and defaults off when absent', () => {
  expect(parsePreferences(JSON.stringify({ captionsOn: true })).captionsOn).toBe(true);
  expect(parsePreferences(JSON.stringify({ volume: 0.5 })).captionsOn).toBe(false);
});

test('a negative or non-finite volume falls back to full volume', () => {
  expect(parsePreferences(JSON.stringify({ volume: -0.2 })).volume).toBe(1);
  expect(parsePreferences(JSON.stringify({ volume: Number.NaN })).volume).toBe(1);
});
