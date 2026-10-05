import { expect, test } from '@rstest/core';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { I18nProvider, createI18n } from '../../app/providers/i18n';
import { LoginView } from './LoginView';
import type { AuthenticatedUser, CredentialsInput } from '../../domain/identity/identity';

const signedIn: AuthenticatedUser = {
  id: '40000000-0000-4000-8000-000000000001',
  email: 'student@courseforge.local',
  displayName: 'Студентка Курсу',
  emailVerified: true,
  twoFactorEnabled: false,
  role: 'student',
};

function view(props: Partial<Parameters<typeof LoginView>[0]> = {}, locale: 'uk' | 'en' = 'uk') {
  const submitted: CredentialsInput[] = [];
  const continued: number[] = [];
  render(
    <I18nProvider instance={createI18n(locale)}>
      <LoginView
        user={null}
        isPending={false}
        errorCode={undefined}
        notice={undefined}
        onSubmit={(input) => submitted.push(input)}
        onContinue={() => continued.push(1)}
        {...props}
      />
    </I18nProvider>,
  );
  return { submitted, continued };
}

test('submits what the visitor typed', async () => {
  const user = userEvent.setup();
  const { submitted } = view();
  await user.type(screen.getByLabelText('Електронна пошта'), 'Student@CourseForge.Local');
  await user.type(screen.getByLabelText('Пароль'), 'secret');
  await user.click(screen.getByRole('button', { name: 'Увійти' }));
  expect(submitted).toEqual([
    { email: 'Student@CourseForge.Local', password: 'secret', rememberMe: false },
  ]);
});

test('submits from the keyboard alone', async () => {
  const user = userEvent.setup();
  const { submitted } = view();
  await user.click(screen.getByLabelText('Електронна пошта'));
  await user.keyboard('a@b.co');
  await user.tab();
  await user.keyboard('secret');
  await user.tab();
  expect(document.activeElement).toBe(screen.getByRole('checkbox', { name: 'Запам’ятати мене' }));
  await user.tab();
  expect(document.activeElement).toBe(screen.getByRole('button', { name: 'Увійти' }));
  await user.keyboard('{Enter}');
  expect(submitted).toEqual([{ email: 'a@b.co', password: 'secret', rememberMe: false }]);
});

test('includes remember-me when the visitor chooses it', async () => {
  const user = userEvent.setup();
  const { submitted } = view();
  await user.type(screen.getByLabelText('Електронна пошта'), 'a@b.co');
  await user.type(screen.getByLabelText('Пароль'), 'secret');
  await user.click(screen.getByRole('checkbox', { name: 'Запам’ятати мене' }));
  await user.click(screen.getByRole('button', { name: 'Увійти' }));
  expect(submitted).toEqual([{ email: 'a@b.co', password: 'secret', rememberMe: true }]);
});

test('announces every refusal in the same place with the same wording', () => {
  view({ errorCode: 'invalid_credentials' });
  expect(screen.getByRole('alert').textContent).toBe('Невірна пошта або пароль.');
});

test.each([
  ['dependency_unavailable', 'Сервіс тимчасово недоступний. Спробуйте пізніше.'],
  ['request_timeout', 'Перевищено час очікування. Спробуйте ще раз.'],
  ['origin_forbidden', 'Запит відхилено з міркувань безпеки. Оновіть сторінку.'],
] as const)('localizes the %s refusal', (code, message) => {
  view({ errorCode: code });
  expect(screen.getByRole('alert').textContent).toBe(message);
});

test('renders the English copy for an English request', () => {
  view({ errorCode: 'invalid_credentials' }, 'en');
  expect(screen.getByRole('alert').textContent).toBe('That email or password is not correct.');
  expect(screen.getByRole('button', { name: 'Log in' })).toBeTruthy();
});

test('blocks a second submit while one is in flight', async () => {
  const user = userEvent.setup();
  const { submitted } = view({ isPending: true });
  const button = screen.getByRole('button', { name: 'Входимо…' });
  await user.click(button);
  expect(submitted).toEqual([]);
  expect(button.hasAttribute('disabled')).toBe(true);
  expect(button.closest('form')?.getAttribute('aria-busy')).toBe('true');
});

test('offers to continue instead of a form when a session already exists', async () => {
  const user = userEvent.setup();
  const { continued } = view({ user: signedIn });
  expect(screen.queryByLabelText('Пароль')).toBe(null);
  expect(screen.getByRole('heading').textContent).toBe('Ви вже увійшли');
  expect(screen.getByText('Студентка Курсу — сеанс активний.')).toBeTruthy();
  await user.click(screen.getByRole('button', { name: 'Продовжити' }));
  expect(continued).toEqual([1]);
});

test('never renders the password as readable text', async () => {
  const user = userEvent.setup();
  view();
  const password = screen.getByLabelText('Пароль');
  await user.type(password, 'secret');
  expect(password.getAttribute('type')).toBe('password');
});

test('submits authenticator and recovery codes on the second login step', async () => {
  const user = userEvent.setup();
  const codes: string[] = [];
  view({ pendingId: 'pending-login', onComplete: (code) => codes.push(code) });
  await user.type(screen.getByLabelText('Шестизначний код'), '123456');
  await user.click(screen.getByRole('button', { name: 'Підтвердити вхід' }));
  await user.click(screen.getByRole('button', { name: 'Ввести резервний код' }));
  await user.type(screen.getByLabelText('Резервний код'), 'recovery-code');
  await user.click(screen.getByRole('button', { name: 'Підтвердити вхід' }));
  expect(codes).toEqual(['123456', 'recovery-code']);
});

test('shows an invalid second-factor refusal', () => {
  view({ pendingId: 'pending-login', errorCode: 'invalid_totp' });
  expect(screen.getByRole('alert').textContent).toBe('Код автентифікатора неправильний.');
});
