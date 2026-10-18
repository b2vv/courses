import { useState } from 'react';
import type { FormEvent } from 'react';
import { useTranslation } from 'react-i18next';
import type { SessionErrorCode } from '../domain/identity';
import { Button, TextField } from '../../../presentation/ui';
type Props = {
  readonly isPending: boolean;
  readonly sent: boolean;
  readonly errorCode: SessionErrorCode | undefined;
  readonly onSubmit: (email: string) => void;
};
export function PasswordForgotView({ isPending, sent, errorCode, onSubmit }: Props) {
  const { t } = useTranslation();
  const [email, setEmail] = useState('');
  function submit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!isPending && event.currentTarget.checkValidity()) onSubmit(email);
  }
  if (sent)
    return (
      <section className="notice">
        <h1>{t('forgot.sentTitle')}</h1>
        <p>{t('forgot.sent')}</p>
      </section>
    );
  return (
    <section className="notice">
      <p className="eyebrow">{t('forgot.label')}</p>
      <h1>{t('forgot.title')}</h1>
      <p>{t('forgot.description')}</p>
      <form onSubmit={submit} noValidate aria-busy={isPending}>
        {errorCode && (
          <p className="ui-field-error" role="alert">
            {t(`forgot.error.${errorCode}`)}
          </p>
        )}
        <TextField
          label={t('forgot.email')}
          name="email"
          type="email"
          autoComplete="email"
          value={email}
          onChange={setEmail}
          isRequired
        />
        <Button type="submit" isDisabled={isPending}>
          {t(isPending ? 'forgot.submitting' : 'forgot.submit')}
        </Button>
      </form>
    </section>
  );
}
