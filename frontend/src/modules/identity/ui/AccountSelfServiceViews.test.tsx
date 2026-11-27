import { expect, test } from '@rstest/core';
import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { createI18n, I18nProvider } from '../../../app/providers/i18n';
import { PasswordForgotView } from './PasswordForgotView';
import { PasswordResetView } from './PasswordResetView';
import { RegisterView } from './RegisterView';
import { VerifyEmailView } from './VerifyEmailView';
import { AccountSettingsView } from './AccountSettingsView';
import { EmailVerificationBanner } from './EmailVerificationBanner';
import { ConfirmEmailChangeView } from './ConfirmEmailChangeView';
import { ResendResult } from './ResendResult';

function view(element: React.ReactNode) {
  return render(<I18nProvider instance={createI18n('uk')}>{element}</I18nProvider>);
}

const settingsProps = {
  currentEmail: 'student@example.com',
  passwordPending: false,
  emailPending: false,
  passwordError: undefined,
  emailError: undefined,
  passwordSavedAt: undefined,
  emailRequestedAt: undefined,
  twoFactorEnabled: false,
  twoFactorPending: false,
  twoFactorError: undefined,
  twoFactorStateUnknown: false,
  enrollment: undefined,
  recovery: undefined,
  onChangePassword: () => undefined,
  onChangeEmail: () => undefined,
  onEnrollTwoFactor: () => undefined,
  onConfirmTwoFactor: () => undefined,
  onCancelTwoFactor: () => undefined,
  onDisableTwoFactor: () => undefined,
  onRegenerateRecoveryCodes: () => undefined,
  onDismissRecoveryCodes: () => undefined,
};
type SettingsProps = React.ComponentProps<typeof AccountSettingsView>;

function settings(overrides: Partial<SettingsProps> = {}) {
  const result = view(<AccountSettingsView {...settingsProps} {...overrides} />);
  return {
    ...result,
    update(next: Partial<SettingsProps>) {
      result.rerender(
        <I18nProvider instance={createI18n('uk')}>
          <AccountSettingsView {...settingsProps} {...next} />
        </I18nProvider>,
      );
    },
  };
}

const verifyProps = {
  resendPending: false,
  resendOutcome: undefined,
  retryPending: false,
  onResend: () => undefined,
  onRetry: () => undefined,
};
const tenCodes = {
  recoveryCodes: [
    'first',
    'second',
    'third',
    'fourth',
    'fifth',
    'sixth',
    'seventh',
    'eighth',
    'ninth',
    'tenth',
  ],
};
const enrollment = {
  secret: 'A'.repeat(16),
  otpauthUri: 'otpauth://totp/CourseForge:test?secret=AAAAAAAAAAAAAAAA',
};

test('S4 asks a guest for the address and resends to it', async () => {
  const user = userEvent.setup();
  const resent: unknown[] = [];
  view(
    <VerifyEmailView
      {...verifyProps}
      state={{ kind: 'link-failed', action: 'guest', offline: false }}
      onResend={(email) => resent.push(email)}
    />,
  );
  expect(screen.getByRole('heading', { level: 1, name: 'Посилання не працює' })).toBeTruthy();
  expect(screen.getByText('Посилання недійсне або вже використане.')).toBeTruthy();
  await user.click(screen.getByRole('button', { name: 'Надіслати лист знову' }));
  expect(resent).toEqual([]);
  expect(screen.getByText('Заповніть це поле.')).toBeTruthy();
  await user.type(screen.getByLabelText('Електронна пошта'), 'student@example.com');
  await user.click(screen.getByRole('button', { name: 'Надіслати лист знову' }));
  expect(resent).toEqual(['student@example.com']);
});

test('S4 resends for a signed-in viewer without asking for the address', async () => {
  const user = userEvent.setup();
  const resent: unknown[] = [];
  view(
    <VerifyEmailView
      {...verifyProps}
      state={{ kind: 'link-failed', action: 'signed-in', offline: false }}
      onResend={(email) => resent.push(email)}
    />,
  );
  expect(screen.queryByLabelText('Електронна пошта')).toBe(null);
  await user.click(screen.getByRole('button', { name: 'Надіслати лист знову' }));
  expect(resent).toEqual([undefined]);
});

test('S4 offers a reload when the session could not be read', () => {
  view(
    <VerifyEmailView
      {...verifyProps}
      state={{ kind: 'link-failed', action: 'reload', offline: false }}
    />,
  );
  expect(screen.getByRole('button', { name: 'Оновити сторінку' })).toBeTruthy();
  expect(screen.queryByRole('button', { name: 'Надіслати лист знову' })).toBe(null);
});

test('S4 keeps the resend button focused and pending while the request runs', () => {
  view(
    <VerifyEmailView
      {...verifyProps}
      resendPending
      state={{ kind: 'link-failed', action: 'signed-in', offline: false }}
    />,
  );
  const button = screen.getByRole('button', { name: 'Надсилаємо…' });
  expect(button.hasAttribute('disabled')).toBe(false);
  expect(button.getAttribute('aria-disabled')).toBe('true');
});

test.each([
  ['accepted', 'Запит прийнято.'],
  ['rate_limited', 'Забагато спроб.'],
  ['failed', 'Не вдалося надіслати лист.'],
] as const)('V1 shows the %s resend outcome', (outcome, text) => {
  view(
    <VerifyEmailView
      {...verifyProps}
      resendOutcome={outcome}
      state={{ kind: 'link-failed', action: 'signed-in', offline: false }}
    />,
  );
  expect(screen.getByText(new RegExp(text.replace('.', '\\.'))).textContent).toContain(text);
});

test('V1 a resend after the session ended offers a full-load sign-in link', () => {
  view(<ResendResult outcome="session_ended" />);
  expect(screen.getByRole('alert').textContent).toContain('Сеанс завершився');
  expect(screen.getByRole('link', { name: 'Увійти' }).getAttribute('href')).toBe('/login');
});

test('S0 and S1 announce a busy check without actions', () => {
  const { unmount } = view(<VerifyEmailView {...verifyProps} state={{ kind: 'waiting' }} />);
  expect(screen.getByText("Чекаємо на з'єднання…")).toBeTruthy();
  expect(screen.queryByRole('button')).toBe(null);
  unmount();
  view(<VerifyEmailView {...verifyProps} state={{ kind: 'checking' }} />);
  expect(screen.getByText('Перевіряємо посилання…')).toBeTruthy();
  expect(
    screen.getByRole('heading', { level: 1 }).closest('section')?.getAttribute('aria-busy'),
  ).toBe('true');
});

test('S2 leads to learning when signed in, and to a full-load sign-in otherwise', () => {
  const { unmount } = view(
    <VerifyEmailView
      {...verifyProps}
      state={{ kind: 'verified', next: 'learn', offline: false }}
    />,
  );
  expect(screen.getByRole('heading', { level: 1, name: 'Пошту підтверджено' })).toBeTruthy();
  expect(screen.getByRole('link', { name: 'До навчання' }).getAttribute('href')).toBe('/learn');
  unmount();
  view(
    <VerifyEmailView
      {...verifyProps}
      state={{ kind: 'verified', next: 'login', offline: false }}
    />,
  );
  expect(screen.getByRole('link', { name: 'Увійти' }).getAttribute('href')).toBe('/login');
  expect(screen.queryByRole('link', { name: 'До навчання' })).toBe(null);
});

test('S3 tells a signed-in viewer that the address is already verified', () => {
  view(<VerifyEmailView {...verifyProps} state={{ kind: 'already-verified' }} />);
  expect(screen.getByRole('heading', { level: 1, name: 'Пошту вже підтверджено' })).toBeTruthy();
  expect(screen.queryByRole('button', { name: 'Надіслати лист знову' })).toBe(null);
});

test('S5 retries an uncertain failure and says the address may be verified', async () => {
  const user = userEvent.setup();
  const retried: number[] = [];
  view(
    <VerifyEmailView
      {...verifyProps}
      state={{ kind: 'failed', code: 'request_timeout' }}
      onRetry={() => retried.push(1)}
    />,
  );
  expect(screen.getByRole('alert').textContent).toContain('Можливо, пошту вже підтверджено.');
  await user.click(screen.getByRole('button', { name: 'Спробувати ще раз' }));
  expect(retried).toEqual([1]);
});

test('S5 has no retry for a refused origin', () => {
  view(<VerifyEmailView {...verifyProps} state={{ kind: 'failed', code: 'origin_forbidden' }} />);
  expect(screen.queryByRole('button', { name: 'Спробувати ще раз' })).toBe(null);
  expect(screen.getByRole('link', { name: 'На головну' }).getAttribute('href')).toBe('/');
});

test.each([
  ['changed', 'Пошту змінено', 'Нову адресу підтверджено.'],
  ['taken', 'Пошту не змінено', 'Ця адреса вже зайнята.'],
  ['link-failed', 'Посилання не працює', 'Посилання недійсне'],
] as const)('E2–E4 confirm-email state %s leads to settings', (kind, heading, text) => {
  view(<ConfirmEmailChangeView state={{ kind }} retryPending={false} onRetry={() => undefined} />);
  expect(screen.getByRole('heading', { level: 1, name: heading })).toBeTruthy();
  expect(screen.getByText(new RegExp(text.replace('.', '\\.')))).toBeTruthy();
  expect(screen.getByRole('link', { name: /налаштувань|Налаштування/ }).getAttribute('href')).toBe(
    '/settings/account',
  );
  expect(screen.queryByRole('button')).toBe(null);
});

test('E5 retries an uncertain failure but not a refused origin', async () => {
  const user = userEvent.setup();
  const retried: number[] = [];
  const { unmount } = view(
    <ConfirmEmailChangeView
      state={{ kind: 'failed', code: 'dependency_unavailable' }}
      retryPending={false}
      onRetry={() => retried.push(1)}
    />,
  );
  expect(screen.getByRole('alert').textContent).toContain('Можливо, пошту вже змінено.');
  await user.click(screen.getByRole('button', { name: 'Спробувати ще раз' }));
  expect(retried).toEqual([1]);
  unmount();
  view(
    <ConfirmEmailChangeView
      state={{ kind: 'failed', code: 'origin_forbidden' }}
      retryPending={false}
      onRetry={() => undefined}
    />,
  );
  expect(screen.queryByRole('button', { name: 'Спробувати ще раз' })).toBe(null);
  expect(screen.getByRole('link', { name: 'Перейти до налаштувань акаунта' })).toBeTruthy();
});

test('changes a password, reports a wrong old one, and clears the form on success', async () => {
  const user = userEvent.setup();
  const submitted: unknown[] = [];
  const screenView = settings({
    passwordError: 'invalid_credentials',
    onChangePassword: (input) => submitted.push(input),
  });
  expect(screen.getByRole('alert').textContent).toContain('Поточний пароль неправильний');
  await user.type(screen.getByLabelText('Поточний пароль'), 'old');
  await user.type(screen.getByLabelText('Новий пароль'), 'new');
  await user.click(screen.getByRole('button', { name: 'Змінити пароль' }));
  expect(submitted).toEqual([{ oldPassword: 'old', newPassword: 'new' }]);
  screenView.update({ passwordSavedAt: 1 });
  expect(screen.getByText('Пароль змінено. Інші сеанси завершено.').tagName).toBe('OUTPUT');
  expect((screen.getByLabelText('Поточний пароль') as HTMLInputElement).value).toBe('');
  expect((screen.getByLabelText('Новий пароль') as HTMLInputElement).value).toBe('');
});

test('V10 keeps password fields read-only while the change is in flight', () => {
  settings({ passwordPending: true });
  expect(screen.getByLabelText('Поточний пароль').hasAttribute('readonly')).toBe(true);
  expect(screen.getByRole('button', { name: 'Зберігаємо…' }).hasAttribute('disabled')).toBe(false);
});

test('V10 adds a recovery hint to an uncertain password change', () => {
  settings({ passwordError: 'request_timeout' });
  expect(screen.getByRole('alert').textContent).toContain('Можливо, пароль уже змінено.');
});

test('requests an email change, shows the current address, and clears on success', async () => {
  const user = userEvent.setup();
  const submitted: unknown[] = [];
  const screenView = settings({ onChangeEmail: (input) => submitted.push(input) });
  expect(screen.getByText('student@example.com').closest('p')?.textContent).toBe(
    'Поточна пошта: student@example.com',
  );
  await user.type(screen.getByLabelText('Нова електронна пошта'), 'next@example.com');
  await user.click(screen.getByRole('button', { name: 'Змінити пошту' }));
  expect(submitted).toEqual([{ newEmail: 'next@example.com' }]);
  screenView.update({ emailRequestedAt: 1 });
  expect(screen.getByText(/Підтвердіть зміну/).tagName).toBe('OUTPUT');
  expect((screen.getByLabelText('Нова електронна пошта') as HTMLInputElement).value).toBe('');
});

test('enrolls, moves focus through setup and codes, and drops the codes on «Готово»', async () => {
  const user = userEvent.setup();
  const enrolled: number[] = [];
  const confirmed: string[] = [];
  const dismissed: number[] = [];
  const screenView = settings({
    onEnrollTwoFactor: () => enrolled.push(1),
    onConfirmTwoFactor: (code) => confirmed.push(code),
  });
  expect(screen.getByText('Захист 2FA ще не ввімкнено.')).toBeTruthy();
  await user.click(screen.getByRole('button', { name: 'Увімкнути 2FA' }));
  expect(enrolled).toEqual([1]);
  screenView.update({ enrollment, onConfirmTwoFactor: (code) => confirmed.push(code) });
  expect(document.activeElement?.id).toBe('two-factor-setup');
  expect(screen.getByDisplayValue('A'.repeat(16)).hasAttribute('readonly')).toBe(true);
  await user.type(screen.getByLabelText('Код автентифікатора або резервний код'), '123456');
  await user.click(screen.getByRole('button', { name: 'Підтвердити 2FA' }));
  expect(confirmed).toEqual(['123456']);
  screenView.update({
    twoFactorEnabled: true,
    recovery: { codes: tenCodes, kind: 'enabled' },
    onDismissRecoveryCodes: () => dismissed.push(1),
  });
  expect(document.activeElement?.id).toBe('two-factor-status');
  expect(screen.getByText('Збережіть ці резервні коди: вдруге їх не буде показано.')).toBeTruthy();
  expect(screen.getAllByRole('listitem')).toHaveLength(10);
  await user.click(screen.getByRole('button', { name: 'Готово' }));
  expect(dismissed).toEqual([1]);
  screenView.update({ twoFactorEnabled: true });
  expect(screen.queryByText('first')).toBe(null);
  expect(document.activeElement?.id).toBe('two-factor-status');
});

test('V5 cancels the setup and returns focus to «Увімкнути 2FA»', async () => {
  const user = userEvent.setup();
  const cancelled: number[] = [];
  const screenView = settings({ enrollment, onCancelTwoFactor: () => cancelled.push(1) });
  await user.click(screen.getByRole('button', { name: 'Скасувати' }));
  expect(cancelled).toEqual([1]);
  screenView.update({});
  expect(document.activeElement?.id).toBe('two-factor-enable');
});

test('V5a hides «Скасувати» and asks for a reload once the 2FA state is unknown', () => {
  settings({
    enrollment,
    twoFactorStateUnknown: true,
    twoFactorError: { action: 'confirm', code: 'request_timeout' },
  });
  expect(screen.queryByRole('button', { name: 'Скасувати' })).toBe(null);
  expect(screen.getByText(/Стан 2FA міг змінитися/)).toBeTruthy();
  expect(screen.getByRole('button', { name: 'Оновити сторінку' })).toBeTruthy();
});

test('V5 a failed confirm keeps focus in the form and shows the error under it', async () => {
  const user = userEvent.setup();
  const screenView = settings({ enrollment });
  const field = screen.getByLabelText('Код автентифікатора або резервний код');
  await user.type(field, '000000');
  await user.click(screen.getByRole('button', { name: 'Підтвердити 2FA' }));
  screenView.update({
    enrollment,
    twoFactorError: { action: 'confirm', code: 'invalid_credentials' },
  });
  expect(screen.getByRole('alert').textContent).toBe('Поточний пароль неправильний.');
  expect(document.activeElement?.id).not.toBe('two-factor-status');
});

test('renders enabled 2FA as two named forms and submits their proofs', async () => {
  const user = userEvent.setup();
  const disabled: unknown[] = [];
  const regenerated: string[] = [];
  settings({
    twoFactorEnabled: true,
    onDisableTwoFactor: (input) => disabled.push(input),
    onRegenerateRecoveryCodes: (code) => regenerated.push(code),
  });
  expect(screen.getByText('2FA увімкнено.')).toBeTruthy();
  const forms = screen.getAllByRole('form');
  expect(forms.map((form) => form.getAttribute('aria-labelledby'))).toEqual([
    'two-factor-codes-title',
    'two-factor-disable-title',
  ]);
  const disableForm = screen.getByRole('form', { name: 'Вимкнути 2FA' });
  await user.type(within(disableForm).getByLabelText('Поточний пароль'), 'secret');
  await user.type(
    within(disableForm).getByLabelText('Код автентифікатора або резервний код'),
    '123456',
  );
  await user.click(within(disableForm).getByRole('button', { name: 'Вимкнути 2FA' }));
  expect(disabled).toEqual([{ password: 'secret', code: '123456' }]);
  const codesForm = screen.getByRole('form', { name: 'Резервні коди' });
  await user.type(
    within(codesForm).getByLabelText('Код автентифікатора або резервний код'),
    '654321',
  );
  await user.click(within(codesForm).getByRole('button', { name: 'Отримати нові резервні коди' }));
  expect(regenerated).toEqual(['654321']);
});

test('V5 shows each 2FA error only under the form that failed', () => {
  settings({
    twoFactorEnabled: true,
    twoFactorError: { action: 'regenerate', code: 'request_timeout' },
  });
  const codesForm = screen.getByRole('form', { name: 'Резервні коди' });
  const disableForm = screen.getByRole('form', { name: 'Вимкнути 2FA' });
  expect(within(codesForm).getByRole('alert').textContent).toContain(
    'Можливо, старі резервні коди вже не діють.',
  );
  expect(within(disableForm).queryByRole('alert')).toBe(null);
});

test('V14 copies the recovery codes and reports a clipboard failure', async () => {
  const user = userEvent.setup();
  const screenView = settings({
    twoFactorEnabled: true,
    recovery: { codes: tenCodes, kind: 'regenerated' },
  });
  expect(screen.getByText('Нові резервні коди')).toBeTruthy();
  await user.click(screen.getByRole('button', { name: 'Скопіювати коди' }));
  expect(await screen.findByText('Резервні коди скопійовано.')).toBeTruthy();
  expect(await navigator.clipboard.readText()).toBe(tenCodes.recoveryCodes.join('\n'));
  screenView.unmount();
  const writeText = navigator.clipboard.writeText;
  Object.defineProperty(navigator.clipboard, 'writeText', {
    configurable: true,
    value: () => Promise.reject(new Error('denied')),
  });
  try {
    settings({ twoFactorEnabled: true, recovery: { codes: tenCodes, kind: 'regenerated' } });
    await user.click(screen.getByRole('button', { name: 'Скопіювати коди' }));
    expect((await screen.findByRole('alert')).textContent).toContain('Не вдалося скопіювати');
  } finally {
    Object.defineProperty(navigator.clipboard, 'writeText', {
      configurable: true,
      value: writeText,
    });
  }
});

test('shows, resends, reports, and dismisses the verification banner until reload', async () => {
  const user = userEvent.setup();
  const resent: number[] = [];
  const { rerender } = view(
    <EmailVerificationBanner isPending={false} onResend={() => resent.push(1)} />,
  );
  expect(screen.getByLabelText('Підтвердження пошти')).toBeTruthy();
  await user.click(screen.getByRole('button', { name: 'Надіслати лист знову' }));
  expect(resent).toEqual([1]);
  rerender(
    <I18nProvider instance={createI18n('uk')}>
      <EmailVerificationBanner isPending={false} outcome="accepted" onResend={() => undefined} />
    </I18nProvider>,
  );
  expect(screen.getByText(/Запит прийнято/).tagName).toBe('OUTPUT');
  await user.click(screen.getByRole('button', { name: 'Закрити' }));
  expect(screen.queryByLabelText('Підтвердження пошти')).toBe(null);
});

test('AUTH-2 always offers the way back to log in from registration', () => {
  view(<RegisterView isPending={false} errorCode={undefined} onSubmit={() => undefined} />);
  expect(screen.getByRole('link', { name: 'Увійти' }).getAttribute('href')).toBe('/login');
  expect(screen.queryByRole('link', { name: 'Скинути пароль' })).toBe(null);
});

test('AUTH-3 leads back to log in from the forgot-password form and its confirmation', () => {
  const { unmount } = view(
    <PasswordForgotView
      isPending={false}
      sent={false}
      errorCode={undefined}
      onSubmit={() => undefined}
    />,
  );
  expect(screen.getByRole('link', { name: 'Увійти' }).getAttribute('href')).toBe('/login');
  unmount();
  view(
    <PasswordForgotView isPending={false} sent errorCode={undefined} onSubmit={() => undefined} />,
  );
  expect(screen.getByRole('link', { name: 'Повернутися до входу' }).getAttribute('href')).toBe(
    '/login',
  );
});

test.each(['invalid_account_token', 'token_used'] as const)(
  'AUTH-4 replaces the form with a request for a new link on %s',
  (errorCode) => {
    view(<PasswordResetView isPending={false} errorCode={errorCode} onSubmit={() => undefined} />);
    expect(screen.getByRole('heading', { level: 1, name: 'Посилання не працює' })).toBeTruthy();
    expect(screen.getByRole('alert').textContent).toContain('недійсне або вже використане');
    expect(
      screen.getByRole('link', { name: 'Запросити нове посилання' }).getAttribute('href'),
    ).toBe('/password/forgot');
    expect(screen.queryByLabelText('Новий пароль')).toBe(null);
  },
);

test('AUTH-4 keeps the form for an error the user can fix', () => {
  view(
    <PasswordResetView isPending={false} errorCode="weak_password" onSubmit={() => undefined} />,
  );
  expect(screen.getByRole('alert').textContent).toBe('Пароль не відповідає вимогам.');
  expect(screen.getByLabelText('Новий пароль')).toBeTruthy();
});
