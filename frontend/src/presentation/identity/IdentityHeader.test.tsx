import { expect, test } from '@rstest/core';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { I18nProvider, createI18n } from '../../app/providers/i18n';
import type { AuthenticatedUser, Role } from '../../domain/identity/identity';
import { IdentityHeader } from './IdentityHeader';

function signedIn(role: Role): AuthenticatedUser {
  return {
    id: `40000000-0000-4000-8000-00000000000${role === 'student' ? '1' : role === 'instructor' ? '2' : '3'}`,
    email: `${role}@courseforge.local`,
    displayName: role,
    role,
  };
}

test.each([
  { state: 'anonymous', role: null, hasName: false, hasSignOut: false, hasAdministration: false },
  { state: 'student', role: 'student', hasName: true, hasSignOut: true, hasAdministration: false },
  {
    state: 'instructor',
    role: 'instructor',
    hasName: true,
    hasSignOut: true,
    hasAdministration: true,
  },
  { state: 'admin', role: 'admin', hasName: true, hasSignOut: true, hasAdministration: true },
] satisfies ReadonlyArray<{
  state: string;
  role: Role | null;
  hasName: boolean;
  hasSignOut: boolean;
  hasAdministration: boolean;
}>)('G20 renders the $state header state', ({ role, hasName, hasSignOut, hasAdministration }) => {
  render(
    <I18nProvider instance={createI18n('uk')}>
      <IdentityHeader
        user={role === null ? null : signedIn(role)}
        isSigningOut={false}
        onSignOut={() => {}}
      />
    </I18nProvider>,
  );

  expect(screen.queryByText(role ?? 'student') !== null).toBe(hasName);
  expect(screen.queryByRole('button', { name: 'Вийти' }) !== null).toBe(hasSignOut);
  expect(screen.queryByRole('link', { name: 'Адміністрування' }) !== null).toBe(hasAdministration);
  expect(screen.queryByRole('link', { name: 'Користувачі' }) !== null).toBe(role === 'admin');
  expect(screen.queryByRole('link', { name: 'Увійти' }) !== null).toBe(role === null);
});

test.each(['Ю', 'Ю'.repeat(120)])('G19 renders the complete display name %s', (displayName) => {
  render(
    <I18nProvider instance={createI18n('uk')}>
      <IdentityHeader
        user={{ ...signedIn('student'), displayName }}
        isSigningOut={false}
        onSignOut={() => {}}
      />
    </I18nProvider>,
  );

  expect(screen.getByText(displayName).textContent).toBe(displayName);
});

test('G29 reaches and activates sign-out from the keyboard without focusing the name', async () => {
  const user = userEvent.setup();
  let signOuts = 0;
  render(
    <I18nProvider instance={createI18n('uk')}>
      <IdentityHeader
        user={signedIn('student')}
        isSigningOut={false}
        onSignOut={() => {
          signOuts += 1;
        }}
      />
    </I18nProvider>,
  );

  expect(screen.getByText('student').getAttribute('tabindex')).toBeNull();
  await user.tab();
  expect(document.activeElement).toBe(screen.getByRole('button', { name: 'Вийти' }));
  await user.keyboard('{Enter}');
  expect(signOuts).toBe(1);
});

test('G31 renders the English anonymous action', () => {
  render(
    <I18nProvider instance={createI18n('en')}>
      <IdentityHeader user={null} isSigningOut={false} onSignOut={() => {}} />
    </I18nProvider>,
  );

  expect(screen.getByRole('link', { name: 'Log in' })).toBeTruthy();
});

test('G31 renders the English signed-in actions', () => {
  render(
    <I18nProvider instance={createI18n('en')}>
      <IdentityHeader user={signedIn('admin')} isSigningOut onSignOut={() => {}} />
    </I18nProvider>,
  );

  expect(screen.getByRole('link', { name: 'Administration' })).toBeTruthy();
  expect(screen.getByRole('link', { name: 'Users' }).getAttribute('href')).toBe('/admin/users');
  expect(screen.getByRole('button', { name: 'Signing out…' })).toBeTruthy();
});
