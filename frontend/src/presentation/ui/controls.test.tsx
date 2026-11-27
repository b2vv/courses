import { expect, test } from '@rstest/core';
import { render, screen } from '@testing-library/react';
import { Form } from 'react-aria-components';
import userEvent from '@testing-library/user-event';
import { createI18n, I18nProvider } from '../../app/providers/i18n';
import type { Locale } from '../../app/composition/locale';
import { Button, TextArea, TextField } from './index';

function form(locale: Locale, fields: React.ReactNode) {
  return render(
    <I18nProvider instance={createI18n(locale)}>
      <form onSubmit={(event) => event.preventDefault()}>
        {fields}
        <Button type="submit">Send</Button>
      </form>
    </I18nProvider>,
  );
}

test.each([
  ['uk', 'Заповніть це поле.', 'Введіть адресу електронної пошти.'],
  ['en', 'Fill in this field.', 'Enter an email address.'],
] as const)(
  'V6 names native constraint errors in the page language (%s)',
  async (locale, required, email) => {
    const user = userEvent.setup();
    form(
      locale,
      <>
        <TextField label="Name" name="name" isRequired />
        <TextField label="Email" name="email" type="email" />
        <TextArea label="Notes" name="notes" isRequired />
      </>,
    );
    await user.type(screen.getByLabelText('Email'), 'not-an-address');
    await user.click(screen.getByRole('button', { name: 'Send' }));
    const name = screen.getByLabelText('Name');
    const describedBy = (element: HTMLElement) =>
      (element.getAttribute('aria-describedby') ?? '')
        .split(' ')
        .map((id) => document.getElementById(id)?.textContent ?? '')
        .join(' ');
    expect(describedBy(name)).toContain(required);
    expect(describedBy(screen.getByLabelText('Email'))).toContain(email);
    expect(describedBy(screen.getByLabelText('Notes'))).toContain(required);
  },
);

test('V6 keeps an explicit error message over the constraint text', async () => {
  const user = userEvent.setup();
  form('uk', <TextField label="Name" name="name" isRequired errorMessage="Вкажіть ім’я" />);
  await user.click(screen.getByRole('button', { name: 'Send' }));
  const name = screen.getByLabelText('Name');
  const text = document.getElementById(name.getAttribute('aria-describedby') ?? '')?.textContent;
  expect(text).toBe('Вкажіть ім’я');
});

test('V6 keeps the text of a validate function and of form-level errors', async () => {
  const user = userEvent.setup();
  render(
    <I18nProvider instance={createI18n('uk')}>
      <Form validationErrors={{ server: 'Сервер каже ні' }}>
        <TextField
          label="Name"
          name="name"
          validate={(value) => (value.length > 3 ? 'Не більше трьох символів' : null)}
        />
        <TextField label="Server" name="server" />
      </Form>
    </I18nProvider>,
  );
  await user.type(screen.getByLabelText('Name'), 'abcd');
  await user.tab();
  const described = (label: string) =>
    (screen.getByLabelText(label).getAttribute('aria-describedby') ?? '')
      .split(' ')
      .map((id) => document.getElementById(id)?.textContent ?? '')
      .join(' ');
  expect(described('Name')).toContain('Не більше трьох символів');
  expect(described('Server')).toContain('Сервер каже ні');
});

test('V6 names a length constraint', async () => {
  const user = userEvent.setup();
  form('uk', <TextField label="Code" name="code" minLength={6} />);
  await user.type(screen.getByLabelText('Code'), '123');
  await user.click(screen.getByRole('button', { name: 'Send' }));
  const code = screen.getByLabelText('Code');
  const text = (code.getAttribute('aria-describedby') ?? '')
    .split(' ')
    .map((id) => document.getElementById(id)?.textContent ?? '')
    .join(' ');
  expect(text).toContain('Перевірте довжину значення.');
});
