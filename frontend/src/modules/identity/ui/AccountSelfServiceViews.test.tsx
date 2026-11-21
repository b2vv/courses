import { expect, test } from '@rstest/core';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { createI18n, I18nProvider } from '../../../app/providers/i18n';
import { PasswordForgotView } from './PasswordForgotView';
import { PasswordResetView } from './PasswordResetView';
import { RegisterView } from './RegisterView';
import { VerifyEmailView } from './VerifyEmailView';
import { AccountSettingsView } from './AccountSettingsView';
import { EmailVerificationBanner } from './EmailVerificationBanner';

function view(element: React.ReactNode) {
  return render(<I18nProvider instance={createI18n('uk')}>{element}</I18nProvider>);
}

const twoFactorProps = {
  twoFactorEnabled: false,
  twoFactorPending: false,
  twoFactorError: undefined,
  enrollment: undefined,
  recoveryCodes: undefined,
  onEnrollTwoFactor: () => undefined,
  onConfirmTwoFactor: () => undefined,
  onDisableTwoFactor: () => undefined,
  onRegenerateRecoveryCodes: () => undefined,
};

test('registers the typed account and offers recovery for an address already taken', async () => {
  const user = userEvent.setup();
  const submitted: unknown[] = [];
  view(
    <RegisterView
      isPending={false}
      errorCode="email_taken"
      onSubmit={(input) => submitted.push(input)}
    />,
  );
  expect(screen.getByRole('link', { name: 'Увійти' })).toBeTruthy();
  expect(screen.getByRole('link', { name: 'Скинути пароль' })).toBeTruthy();
  await user.type(screen.getByLabelText('Ім’я'), 'Нова людина');
  await user.type(screen.getByLabelText('Електронна пошта'), 'new@example.com');
  await user.type(screen.getByLabelText('Пароль'), 'secret');
  await user.click(screen.getByRole('button', { name: 'Зареєструватися' }));
  expect(submitted).toEqual([
    { displayName: 'Нова людина', email: 'new@example.com', password: 'secret' },
  ]);
});

test('offers another verification email after an expired link', async () => {
  const user = userEvent.setup();
  const resent: number[] = [];
  view(
    <VerifyEmailView
      isPending={false}
      completed={false}
      errorCode="token_used"
      onResend={() => resent.push(1)}
    />,
  );
  expect(screen.getByText('Посилання недійсне або вже використане.')).toBeTruthy();
  await user.click(screen.getByRole('button', { name: 'Надіслати лист знову' }));
  expect(resent).toEqual([1]);
});

test('keeps forgot-password confirmation neutral', async () => {
  view(
    <PasswordForgotView isPending={false} sent errorCode={undefined} onSubmit={() => undefined} />,
  );
  expect(screen.getByText('Якщо акаунт існує, лист для скидання пароля надіслано.')).toBeTruthy();
  expect(screen.queryByText(/існує для цієї пошти/)).toBe(null);
});

test('submits a new password', async () => {
  const user = userEvent.setup();
  const passwords: string[] = [];
  view(
    <PasswordResetView
      isPending={false}
      errorCode={undefined}
      onSubmit={(password) => passwords.push(password)}
    />,
  );
  await user.type(screen.getByLabelText('Новий пароль'), 'new secret');
  await user.click(screen.getByRole('button', { name: 'Зберегти пароль' }));
  expect(passwords).toEqual(['new secret']);
});

test('changes a password and reports a wrong old password', async () => {
  const user = userEvent.setup();
  const submitted: unknown[] = [];
  const { rerender } = view(
    <AccountSettingsView
      {...twoFactorProps}
      passwordPending={false}
      emailPending={false}
      passwordError="invalid_credentials"
      emailError={undefined}
      passwordSaved={false}
      emailRequested={false}
      onChangePassword={(input) => submitted.push(input)}
      onChangeEmail={() => undefined}
    />,
  );
  expect(screen.getByRole('alert').textContent).toContain('Поточний пароль неправильний');
  await user.type(screen.getByLabelText('Поточний пароль'), 'old');
  await user.type(screen.getByLabelText('Новий пароль'), 'new');
  await user.click(screen.getByRole('button', { name: 'Змінити пароль' }));
  expect(submitted).toEqual([{ oldPassword: 'old', newPassword: 'new' }]);
  rerender(
    <I18nProvider instance={createI18n('uk')}>
      <AccountSettingsView
        {...twoFactorProps}
        passwordPending={false}
        emailPending={false}
        passwordError={undefined}
        emailError={undefined}
        passwordSaved
        emailRequested={false}
        onChangePassword={() => undefined}
        onChangeEmail={() => undefined}
      />
    </I18nProvider>,
  );
  expect(screen.getByRole('status').textContent).toContain('Інші сеанси завершено');
});

test('requests an email change and shows its confirmation notice', async () => {
  const user = userEvent.setup();
  const submitted: unknown[] = [];
  const { rerender } = view(
    <AccountSettingsView
      {...twoFactorProps}
      passwordPending={false}
      emailPending={false}
      passwordError={undefined}
      emailError={undefined}
      passwordSaved={false}
      emailRequested={false}
      onChangePassword={() => undefined}
      onChangeEmail={(input) => submitted.push(input)}
    />,
  );
  await user.type(screen.getByLabelText('Нова електронна пошта'), 'next@example.com');
  await user.click(screen.getByRole('button', { name: 'Змінити пошту' }));
  expect(submitted).toEqual([{ newEmail: 'next@example.com' }]);
  rerender(
    <I18nProvider instance={createI18n('uk')}>
      <AccountSettingsView
        {...twoFactorProps}
        passwordPending={false}
        emailPending={false}
        passwordError={undefined}
        emailError={undefined}
        passwordSaved={false}
        emailRequested
        onChangePassword={() => undefined}
        onChangeEmail={() => undefined}
      />
    </I18nProvider>,
  );
  expect(screen.getByRole('status').textContent).toContain('Підтвердіть зміну');
});

test('enrolls, confirms, and exposes recovery codes exactly once', async () => {
  const user = userEvent.setup();
  const enrolled: number[] = [];
  const confirmed: string[] = [];
  const { rerender } = view(
    <AccountSettingsView
      {...twoFactorProps}
      passwordPending={false}
      emailPending={false}
      passwordError={undefined}
      emailError={undefined}
      passwordSaved={false}
      emailRequested={false}
      onChangePassword={() => undefined}
      onChangeEmail={() => undefined}
      onEnrollTwoFactor={() => enrolled.push(1)}
      onConfirmTwoFactor={(code) => confirmed.push(code)}
    />,
  );
  expect(screen.getByText('Захист 2FA ще не ввімкнено.')).toBeTruthy();
  await user.click(screen.getByRole('button', { name: 'Увімкнути 2FA' }));
  expect(enrolled).toEqual([1]);
  rerender(
    <I18nProvider instance={createI18n('uk')}>
      <AccountSettingsView
        {...twoFactorProps}
        enrollment={{
          secret: 'A'.repeat(16),
          otpauthUri: 'otpauth://totp/CourseForge:test?secret=AAAAAAAAAAAAAAAA',
        }}
        passwordPending={false}
        emailPending={false}
        passwordError={undefined}
        emailError={undefined}
        passwordSaved={false}
        emailRequested={false}
        onChangePassword={() => undefined}
        onChangeEmail={() => undefined}
        onConfirmTwoFactor={(code) => confirmed.push(code)}
      />
    </I18nProvider>,
  );
  expect(screen.getByLabelText('Двофакторна автентифікація')).toBeTruthy();
  expect(screen.getByDisplayValue('A'.repeat(16))).toBeTruthy();
  await user.type(screen.getByLabelText('Код автентифікатора або резервний код'), '123456');
  await user.click(screen.getByRole('button', { name: 'Підтвердити 2FA' }));
  expect(confirmed).toEqual(['123456']);
  rerender(
    <I18nProvider instance={createI18n('uk')}>
      <AccountSettingsView
        {...twoFactorProps}
        twoFactorEnabled
        recoveryCodes={{
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
        }}
        passwordPending={false}
        emailPending={false}
        passwordError={undefined}
        emailError={undefined}
        passwordSaved={false}
        emailRequested={false}
        onChangePassword={() => undefined}
        onChangeEmail={() => undefined}
      />
    </I18nProvider>,
  );
  expect(screen.getByText('Збережіть ці резервні коди: вдруге їх не буде показано.')).toBeTruthy();
  expect(screen.getByText('first')).toBeTruthy();
  rerender(
    <I18nProvider instance={createI18n('uk')}>
      <AccountSettingsView
        {...twoFactorProps}
        twoFactorEnabled
        passwordPending={false}
        emailPending={false}
        passwordError={undefined}
        emailError={undefined}
        passwordSaved={false}
        emailRequested={false}
        onChangePassword={() => undefined}
        onChangeEmail={() => undefined}
      />
    </I18nProvider>,
  );
  expect(screen.queryByText('first')).toBe(null);
});

test('renders enabled controls and submits disable and regeneration proofs', async () => {
  const user = userEvent.setup();
  const disabled: unknown[] = [];
  const regenerated: string[] = [];
  view(
    <AccountSettingsView
      {...twoFactorProps}
      twoFactorEnabled
      passwordPending={false}
      emailPending={false}
      passwordError={undefined}
      emailError={undefined}
      passwordSaved={false}
      emailRequested={false}
      onChangePassword={() => undefined}
      onChangeEmail={() => undefined}
      onDisableTwoFactor={(input) => disabled.push(input)}
      onRegenerateRecoveryCodes={(code) => regenerated.push(code)}
    />,
  );
  expect(screen.getByText('2FA увімкнено.')).toBeTruthy();
  await user.type(
    screen.getByLabelText('Поточний пароль', {
      selector: '[name="two-factor-disable-password"]',
    }),
    'secret',
  );
  await user.type(
    screen.getByLabelText('Код автентифікатора або резервний код', {
      selector: '[name="two-factor-disable-code"]',
    }),
    '123456',
  );
  await user.click(screen.getByRole('button', { name: 'Вимкнути 2FA' }));
  expect(disabled).toEqual([{ password: 'secret', code: '123456' }]);
  await user.type(
    screen.getByLabelText('Код автентифікатора або резервний код', {
      selector: '[name="two-factor-regenerate-code"]',
    }),
    '654321',
  );
  await user.click(screen.getByRole('button', { name: 'Перегенерувати резервні коди' }));
  expect(regenerated).toEqual(['654321']);
});

test('shows, resends, and dismisses the verification banner until reload', async () => {
  const user = userEvent.setup();
  const resent: number[] = [];
  view(<EmailVerificationBanner isPending={false} onResend={() => resent.push(1)} />);
  expect(screen.getByLabelText('Підтвердження пошти')).toBeTruthy();
  await user.click(screen.getByRole('button', { name: 'Надіслати лист знову' }));
  expect(resent).toEqual([1]);
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
