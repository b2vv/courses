import { useState } from 'react';
import type { FormEvent } from 'react';
import { useTranslation } from 'react-i18next';
import type {
  AuthenticatedUser,
  CredentialsInput,
  SessionErrorCode,
} from '../../domain/identity/identity';
import { Button, TextField } from '../ui';

type LoginViewProps = {
  readonly user: AuthenticatedUser | null;
  readonly isPending: boolean;
  readonly errorCode: SessionErrorCode | undefined;
  readonly onSubmit: (input: CredentialsInput) => void;
  readonly onContinue: () => void;
};

export function LoginView({ user, isPending, errorCode, onSubmit, onContinue }: LoginViewProps) {
  const { t } = useTranslation();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');

  if (user)
    return (
      <section className="notice">
        <p className="eyebrow">{t('login.label')}</p>
        <h1>{t('login.signedInTitle')}</h1>
        <p>{t('login.signedInText', { name: user.displayName })}</p>
        <Button onPress={onContinue}>{t('login.continue')}</Button>
      </section>
    );

  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!isPending) onSubmit({ email, password });
  }

  return (
    <section className="notice">
      <p className="eyebrow">{t('login.label')}</p>
      <h1>{t('login.title')}</h1>
      <p>{t('login.description')}</p>
      <form onSubmit={submit} noValidate aria-busy={isPending}>
        {/*
          Every refusal renders the same message in the same place, so the form
          never reveals whether an address exists.
        */}
        {errorCode && (
          <p className="ui-field-error" role="alert">
            {t(`login.error.${errorCode}`)}
          </p>
        )}
        <TextField
          label={t('login.email')}
          placeholder={t('login.emailPlaceholder')}
          name="email"
          type="email"
          autoComplete="username"
          value={email}
          onChange={setEmail}
          isRequired
        />
        <TextField
          label={t('login.password')}
          name="password"
          type="password"
          autoComplete="current-password"
          value={password}
          onChange={setPassword}
          isRequired
        />
        <Button type="submit" isDisabled={isPending}>
          {isPending ? t('login.submitting') : t('login.submit')}
        </Button>
      </form>
    </section>
  );
}
