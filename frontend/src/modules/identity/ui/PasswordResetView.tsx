import { useState } from 'react';
import type { FormEvent } from 'react';
import { useTranslation } from 'react-i18next';
import type { SessionErrorCode } from '../domain/identity';
import { Button, Link, TextField } from '../../../presentation/ui';
type Props = {
  readonly isPending: boolean;
  readonly errorCode: SessionErrorCode | undefined;
  readonly onSubmit: (password: string) => void;
};
// A missing, unknown or spent token cannot be retried with another password: the form would
// only fail again. These codes switch the screen to the way out — asking for a fresh link.
const deadLinkCodes: ReadonlySet<SessionErrorCode> = new Set([
  'invalid_account_token',
  'token_used',
]);
export function PasswordResetView({ isPending, errorCode, onSubmit }: Props) {
  const { t } = useTranslation();
  const [password, setPassword] = useState('');
  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!isPending && event.currentTarget.checkValidity()) onSubmit(password);
  }
  if (errorCode !== undefined && deadLinkCodes.has(errorCode))
    return (
      <section className="notice">
        <p className="eyebrow">{t('reset.label')}</p>
        <h1>{t('reset.invalidTitle')}</h1>
        <p role="alert">{t('reset.invalidText')}</p>
        <Link href="/password/forgot" className="ui-button ui-button--primary">
          {t('reset.requestNew')}
        </Link>
      </section>
    );
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
