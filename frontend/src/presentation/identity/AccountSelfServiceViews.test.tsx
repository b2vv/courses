import { expect, test } from '@rstest/core';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { createI18n, I18nProvider } from '../../app/providers/i18n';
import { PasswordForgotView } from './PasswordForgotView';
import { PasswordResetView } from './PasswordResetView';
import { RegisterView } from './RegisterView';
import { VerifyEmailView } from './VerifyEmailView';
import { AccountSettingsView } from './AccountSettingsView';
import { EmailVerificationBanner } from './EmailVerificationBanner';

function view(element: React.ReactNode) {
  return render(<I18nProvider instance={createI18n('uk')}>{element}</I18nProvider>);
}

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
