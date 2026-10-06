import { useState } from 'react';
import type { FormEvent } from 'react';
import { useTranslation } from 'react-i18next';
import type { SessionErrorCode } from '../../domain/identity/identity';
import { Button, TextField } from '../ui';
type Props = {
  readonly isPending: boolean;
  readonly errorCode: SessionErrorCode | undefined;
  readonly onSubmit: (password: string) => void;
};
export function PasswordResetView({ isPending, errorCode, onSubmit }: Props) {
  const { t } = useTranslation();
  const [password, setPassword] = useState('');
  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!isPending && event.currentTarget.checkValidity()) onSubmit(password);
  }
  return (
    <section className="notice">
      <p className="eyebrow">{t('reset.label')}</p>
      <h1>{t('reset.title')}</h1>
      <form onSubmit={submit} noValidate aria-busy={isPending}>
        {errorCode && (
          <p className="ui-field-error" role="alert">
            {t(`reset.error.${errorCode}`)}
          </p>
        )}
        <TextField
          label={t('reset.password')}
          name="password"
          type="password"
          autoComplete="new-password"
          value={password}
          onChange={setPassword}
          isRequired
        />
        <Button type="submit" isDisabled={isPending}>
          {t(isPending ? 'reset.submitting' : 'reset.submit')}
        </Button>
      </form>
    </section>
  );
}
