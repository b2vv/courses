import { useState } from 'react';
import type { FormEvent } from 'react';
import { useTranslation } from 'react-i18next';
import type { RegisterInput, SessionErrorCode } from '../../domain/identity/identity';
import { Button, Link, TextField } from '../ui';

type Props = {
  readonly isPending: boolean;
  readonly errorCode: SessionErrorCode | undefined;
  readonly onSubmit: (input: RegisterInput) => void;
};

export function RegisterView({ isPending, errorCode, onSubmit }: Props) {
  const { t } = useTranslation();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [displayName, setDisplayName] = useState('');
  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!isPending && event.currentTarget.checkValidity())
      onSubmit({ email, password, displayName });
  }
  return (
    <section className="notice">
      <p className="eyebrow">{t('register.label')}</p>
      <h1>{t('register.title')}</h1>
      <p>{t('register.description')}</p>
      <form onSubmit={submit} noValidate aria-busy={isPending}>
        {errorCode && (
          <p className="ui-field-error" role="alert">
            {t(`register.error.${errorCode}`)}
          </p>
        )}
        <TextField
          label={t('register.name')}
          name="name"
          autoComplete="name"
          value={displayName}
          onChange={setDisplayName}
          isRequired
        />
        <TextField
          label={t('register.email')}
          name="email"
          type="email"
          autoComplete="email"
          value={email}
          onChange={setEmail}
          isRequired
        />
        <TextField
          label={t('register.password')}
          name="password"
          type="password"
          autoComplete="new-password"
          value={password}
          onChange={setPassword}
          isRequired
        />
        {errorCode === 'email_taken' && (
          <p>
            <Link href="/login">{t('register.login')}</Link>
            {' · '}
            <Link href="/password/forgot">{t('register.forgot')}</Link>
          </p>
        )}
        <Button type="submit" isDisabled={isPending}>
          {t(isPending ? 'register.submitting' : 'register.submit')}
        </Button>
      </form>
    </section>
  );
}
