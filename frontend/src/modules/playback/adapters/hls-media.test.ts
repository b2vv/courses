import { expect, test } from '@rstest/core';
import { exactAudioTrackIndex } from './hls-media';

const track = (lang: string | undefined, name = 'Track', isDefault = false) => ({
  lang,
  name,
  default: isDefault,
});

// hls.js languagesMatch accepts a prefix in either direction; both pairs below can coexist on one
// asset (migration 00028 falls back to und-x-original when und is taken), so selection is exact.
test.each([
  [['und', 'und-x-original'], 'und-x-original', 1],
  [['und-x-original', 'und'], 'und', 1],
  [['und', 'und-x-original'], 'und', 0],
  [['en', 'en-us'], 'en-us', 1],
  [['en-us', 'en'], 'en', 1],
] as const)('%j selects exactly %s', (languages, wanted, index) => {
  expect(
    exactAudioTrackIndex(
      languages.map((lang) => track(lang)),
      wanted,
    ),
  ).toBe(index);
});

test('the case is part of the language, as the backend keeps it', () => {
  expect(exactAudioTrackIndex([track('en-us')], 'en-US')).toBe(-1);
  expect(exactAudioTrackIndex([track('en-us'), track('en-US')], 'en-US')).toBe(1);
});

test('an absent or unset language selects nothing, so the manifest default plays', () => {
  expect(exactAudioTrackIndex([track('und', 'Original', true), track('en')], 'uk')).toBe(-1);
  expect(exactAudioTrackIndex([track('und', 'Original', true)], null)).toBe(-1);
  expect(exactAudioTrackIndex([track(undefined)], 'und')).toBe(-1);
});

test('shared names do not matter: the language decides', () => {
  expect(exactAudioTrackIndex([track('uk', 'Dub'), track('en', 'Dub')], 'en')).toBe(1);
});
