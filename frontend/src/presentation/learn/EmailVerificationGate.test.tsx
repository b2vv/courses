import { expect, test } from '@rstest/core';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { createI18n, I18nProvider } from '../../app/providers/i18n';
import { EmailVerificationGate, shouldGateLessonPlayback } from './EmailVerificationGate';

test('offers resend instead of a player for an unverified learner', async () => {
  const user = userEvent.setup();
  const resent: number[] = [];
  render(
    <I18nProvider instance={createI18n('uk')}>
      <EmailVerificationGate isPending={false} onResend={() => resent.push(1)} />
    </I18nProvider>,
  );
  expect(screen.getByRole('heading', { name: 'Підтвердіть пошту, щоб дивитися' })).toBeTruthy();
  expect(screen.getByRole('link', { name: 'Налаштування акаунта' }).getAttribute('href')).toBe(
    '/settings/account',
  );
  await user.click(screen.getByRole('button', { name: 'Надіслати лист знову' }));
  expect(resent).toEqual([1]);
});

test('gates only an unverified signed-in learner', () => {
  const user = {
    id: 'id',
    email: 'user@example.com',
    displayName: 'User',
    role: 'student' as const,
  };
  expect(
    shouldGateLessonPlayback({ status: 'signed-in', user: { ...user, emailVerified: false } }),
  ).toBe(true);
  expect(
    shouldGateLessonPlayback({ status: 'signed-in', user: { ...user, emailVerified: true } }),
  ).toBe(false);
  expect(shouldGateLessonPlayback({ status: 'anonymous' })).toBe(false);
});
