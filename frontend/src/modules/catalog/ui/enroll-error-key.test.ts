import { expect, test } from '@rstest/core';
import { createI18n } from '../../../app/providers/i18n';
import { enrollErrorKey } from './CourseView';

// An administrator who tries to enrol is told why, not shown the generic failure.
test('names the admin_not_learner enrol refusal', () => {
  const key = `enrollment.errors.${enrollErrorKey('admin_not_learner')}`;
  expect(createI18n('en').t(key)).toBe(
    'Administrators are not learners and cannot enroll in courses.',
  );
  expect(createI18n('uk').t(key)).toBe(
    'Адміністратори не є слухачами й не можуть записуватися на курси.',
  );
});
