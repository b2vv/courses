import { expect, test } from '@rstest/core';
import { createI18n } from '../../../app/providers/i18n';

// A field the server refused must be named, not reported as an unexpected response.
test.each([
  ['invalid_email', 'Enter a valid email address.', 'Введіть коректну адресу пошти.'],
  ['invalid_display_name', 'Enter a valid display name.', 'Введіть коректне ім’я.'],
] as const)('names the %s registration refusal', (code, english, ukrainian) => {
  expect(createI18n('en').t(`register.error.${code}`)).toBe(english);
  expect(createI18n('uk').t(`register.error.${code}`)).toBe(ukrainian);
});
