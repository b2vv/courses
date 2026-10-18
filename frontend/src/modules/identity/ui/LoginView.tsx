import { useState } from 'react';
import type { FormEvent } from 'react';
import { useTranslation } from 'react-i18next';
import type { AuthenticatedUser, CredentialsInput, SessionErrorCode } from '../domain/identity';
import { Button, Checkbox, TextField } from '../../../presentation/ui';

type LoginViewProps = {
  readonly user: AuthenticatedUser | null;
  readonly isPending: boolean;
  readonly errorCode: SessionErrorCode | undefined;
  readonly notice: 'password-reset' | undefined;
  readonly onSubmit: (input: CredentialsInput) => void;
  readonly pendingId?: string;
  readonly onComplete?: (code: string) => void;
  readonly onContinue: () => void;
};

export function LoginView({
  user,
  isPending,
  errorCode,
  notice,
  onSubmit,
  pendingId,
  onComplete,
  onContinue,
}: LoginViewProps) {
  const { t } = useTranslation();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [rememberMe, setRememberMe] = useState(false);
  const [code, setCode] = useState('');
  const [recovery, setRecovery] = useState(false);

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
    if (!isPending && event.currentTarget.checkValidity())
      onSubmit({ email, password, rememberMe });
  }
  function submitSecondFactor(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!isPending && event.currentTarget.checkValidity()) onComplete?.(code);
  }

  return (
    <section className="notice">
      <p className="eyebrow">{t('login.label')}</p>
      <h1>{t('login.title')}</h1>
      <p>{pendingId === undefined ? t('login.description') : t('login.twoFactor.description')}</p>
      {notice === 'password-reset' && <output>{t('login.passwordReset')}</output>}
      {pendingId !== undefined ? (
        <form onSubmit={submitSecondFactor} noValidate aria-busy={isPending}>
          <h2>{t('login.twoFactor.title')}</h2>
          {errorCode && (
            <p className="ui-field-error" role="alert">
              {/* The server tries the code as both TOTP and recovery and reports one invalid_totp;
                  in recovery mode name it as a recovery-code failure, matching what the user entered. */}
              {t(
                `login.error.${recovery && errorCode === 'invalid_totp' ? 'recovery_code_invalid' : errorCode}`,
              )}
            </p>
          )}
          <TextField
            label={t(recovery ? 'login.twoFactor.recoveryCode' : 'login.twoFactor.code')}
            name="two-factor-code"
            inputMode={recovery ? 'text' : 'numeric'}
            autoComplete="one-time-code"
            value={code}
            onChange={setCode}
            isRequired
          />
          <Button
            type="button"
            variant="secondary"
            onPress={() => {
              setRecovery((value) => !value);
              setCode('');
            }}
          >
            {t(recovery ? 'login.twoFactor.useAuthenticator' : 'login.twoFactor.useRecovery')}
          </Button>
          <Button type="submit" isDisabled={isPending}>
            {t(isPending ? 'login.twoFactor.submitting' : 'login.twoFactor.submit')}
          </Button>
        </form>
      ) : (
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
          <Checkbox isSelected={rememberMe} onChange={setRememberMe}>
            {t('login.rememberMe')}
          </Checkbox>
          <Button type="submit" isDisabled={isPending}>
            {isPending ? t('login.submitting') : t('login.submit')}
          </Button>
        </form>
      )}
    </section>
  );
}
