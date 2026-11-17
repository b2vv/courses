import { expect, test } from '@rstest/core';
import { draftLoadable, isUndetermined } from './draft';

test.each([
  ['en', false, 'EN', false, false, true],
  ['en', true, 'en', true, false, true],
  ['en', false, 'uk', false, false, false],
  ['en', false, 'en', false, true, false],
  ['en', true, 'en', false, false, false],
  ['en', false, 'en', true, false, false],
  ['und', false, 'uk', false, false, true],
  ['und-x-original', true, 'en', true, false, true],
  ['und', false, 'en', true, false, false],
  ['und-x-original', true, 'uk', false, false, false],
  ['und', true, 'uk', true, true, false],
  ['under', false, 'en', false, false, false],
] as const)(
  'loadable snapshot %s/%s for %s/%s stale=%s → %s',
  (trackLanguage, trackMultilingual, language, multilingual, stale, expected) => {
    expect(
      draftLoadable({ trackLanguage, trackMultilingual, stale }, { language, multilingual }),
    ).toBe(expected);
  },
);

test.each([
  ['und', true],
  ['und-x-original', true],
  ['under', false],
  ['en', false],
] as const)('und family %s → %s', (language, expected) =>
  expect(isUndetermined(language)).toBe(expected),
);
