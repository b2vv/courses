import { expect, test } from '@rstest/core';
import { createI18n } from '../../../app/providers/i18n';
import { purchaseErrorKey } from './LockedLessonPanel';

// A named refusal must reach its own sentence in both languages, not the generic "unavailable".
test.each([
  [
    'email_not_verified',
    'Confirm your email to buy the course.',
    'Підтвердіть пошту, щоб купити курс.',
  ],
  [
    'admin_not_learner',
    'Administrators cannot buy courses.',
    'Адміністратори не можуть купувати курси.',
  ],
] as const)('names the %s purchase refusal', (code, english, ukrainian) => {
  const key = `learn.purchase.errors.${purchaseErrorKey(code)}`;
  expect(createI18n('en').t(key)).toBe(english);
  expect(createI18n('uk').t(key)).toBe(ukrainian);
});
