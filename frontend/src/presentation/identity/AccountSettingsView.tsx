import { useState, type FormEvent } from 'react';
import { useTranslation } from 'react-i18next';
import type { SessionErrorCode } from '../../domain/identity/identity';
import { Button, TextField } from '../ui';

type Props = {
  readonly passwordPending: boolean;
  readonly emailPending: boolean;
  readonly passwordError: SessionErrorCode | undefined;
  readonly emailError: SessionErrorCode | undefined;
  readonly passwordSaved: boolean;
  readonly emailRequested: boolean;
  readonly onChangePassword: (input: { oldPassword: string; newPassword: string }) => void;
  readonly onChangeEmail: (input: { newEmail: string }) => void;
};

export function AccountSettingsView(props: Props) {
  const { t } = useTranslation();
  const [oldPassword, setOldPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [email, setEmail] = useState('');
  function submitPassword(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!props.passwordPending && event.currentTarget.checkValidity())
      props.onChangePassword({ oldPassword, newPassword });
  }
  function submitEmail(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!props.emailPending && event.currentTarget.checkValidity())
      props.onChangeEmail({ newEmail: email });
  }
  return (
    <section className="account-settings">
      <div>
        <p className="eyebrow">{t('account.label')}</p>
        <h1>{t('account.title')}</h1>
      </div>
      <section className="account-settings-card" aria-labelledby="change-password-title">
        <h2 id="change-password-title">{t('account.password.title')}</h2>
        <form onSubmit={submitPassword} noValidate aria-busy={props.passwordPending}>
          {props.passwordSaved && <output>{t('account.password.success')}</output>}
          {props.passwordError && (
            <p className="ui-field-error" role="alert">
              {t(`account.error.${props.passwordError}`)}
            </p>
          )}
          <TextField
            label={t('account.password.old')}
            name="old-password"
            type="password"
            autoComplete="current-password"
            value={oldPassword}
            onChange={setOldPassword}
            isRequired
          />
          <TextField
            label={t('account.password.new')}
            name="new-password"
            type="password"
            autoComplete="new-password"
            value={newPassword}
            onChange={setNewPassword}
            isRequired
          />
          <Button type="submit" isDisabled={props.passwordPending}>
            {t(props.passwordPending ? 'account.password.saving' : 'account.password.submit')}
          </Button>
        </form>
      </section>
      <section className="account-settings-card" aria-labelledby="change-email-title">
        <h2 id="change-email-title">{t('account.email.title')}</h2>
        <form onSubmit={submitEmail} noValidate aria-busy={props.emailPending}>
          {props.emailRequested && <output>{t('account.email.success')}</output>}
          {props.emailError && (
            <p className="ui-field-error" role="alert">
              {t(`account.error.${props.emailError}`)}
            </p>
          )}
          <TextField
            label={t('account.email.new')}
            name="email"
            type="email"
            autoComplete="email"
            value={email}
            onChange={setEmail}
            isRequired
          />
          <Button type="submit" isDisabled={props.emailPending}>
            {t(props.emailPending ? 'account.email.saving' : 'account.email.submit')}
          </Button>
        </form>
      </section>
    </section>
  );
}
