import { useState, type FormEvent } from 'react';
import { useTranslation } from 'react-i18next';
import { QRCodeSVG } from 'qrcode.react';
import type { SessionErrorCode } from '../domain/identity';
import type { RecoveryCodes, TwoFactorEnrollment } from '../domain/identity';
import { Button, TextField } from '../../../presentation/ui';

type Props = {
  readonly passwordPending: boolean;
  readonly emailPending: boolean;
  readonly passwordError: SessionErrorCode | undefined;
  readonly emailError: SessionErrorCode | undefined;
  readonly passwordSaved: boolean;
  readonly emailRequested: boolean;
  readonly twoFactorEnabled: boolean;
  readonly twoFactorPending: boolean;
  readonly twoFactorError: SessionErrorCode | undefined;
  readonly enrollment: TwoFactorEnrollment | undefined;
  readonly recoveryCodes: RecoveryCodes | undefined;
  readonly onChangePassword: (input: { oldPassword: string; newPassword: string }) => void;
  readonly onChangeEmail: (input: { newEmail: string }) => void;
  readonly onEnrollTwoFactor: () => void;
  readonly onConfirmTwoFactor: (code: string) => void;
  readonly onDisableTwoFactor: (input: { password: string; code: string }) => void;
  readonly onRegenerateRecoveryCodes: (code: string) => void;
};

export function AccountSettingsView(props: Props) {
  const { t } = useTranslation();
  const [oldPassword, setOldPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [email, setEmail] = useState('');
  const [twoFactorCode, setTwoFactorCode] = useState('');
  const [disablePassword, setDisablePassword] = useState('');
  const [disableCode, setDisableCode] = useState('');
  const [regenerationCode, setRegenerationCode] = useState('');
  const [copied, setCopied] = useState(false);
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
  function submitConfirmation(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!props.twoFactorPending && event.currentTarget.checkValidity())
      props.onConfirmTwoFactor(twoFactorCode);
  }
  function submitDisable(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!props.twoFactorPending && event.currentTarget.checkValidity())
      props.onDisableTwoFactor({ password: disablePassword, code: disableCode });
  }
  function submitRegeneration(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (!props.twoFactorPending && event.currentTarget.checkValidity())
      props.onRegenerateRecoveryCodes(regenerationCode);
  }
  async function copyRecoveryCodes() {
    if (props.recoveryCodes === undefined || !navigator.clipboard) return;
    await navigator.clipboard.writeText(props.recoveryCodes.recoveryCodes.join('\n'));
    setCopied(true);
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
      <section className="account-settings-card" aria-labelledby="two-factor-title">
        <h2 id="two-factor-title">{t('account.twoFactor.title')}</h2>
        {props.recoveryCodes ? (
          <div className="two-factor-recovery" aria-live="polite">
            <p>{t('account.twoFactor.recovery')}</p>
            <ul aria-label={t('account.twoFactor.recovery')}>
              {props.recoveryCodes.recoveryCodes.map((recoveryCode) => (
                <li key={recoveryCode}>
                  <code>{recoveryCode}</code>
                </li>
              ))}
            </ul>
            <Button type="button" variant="secondary" onPress={() => void copyRecoveryCodes()}>
              {t('account.twoFactor.copy')}
            </Button>
            {copied && <output>{t('account.twoFactor.copied')}</output>}
          </div>
        ) : props.enrollment ? (
          <form onSubmit={submitConfirmation} noValidate aria-busy={props.twoFactorPending}>
            <p>{t('account.twoFactor.setup')}</p>
            <QRCodeSVG
              value={props.enrollment.otpauthUri}
              aria-label={t('account.twoFactor.qrAlt')}
            />
            <TextField
              label={t('account.twoFactor.secret')}
              name="two-factor-secret"
              value={props.enrollment.secret}
              isReadOnly
            />
            <TextField
              label={t('account.twoFactor.code')}
              name="two-factor-confirmation-code"
              inputMode="numeric"
              autoComplete="one-time-code"
              value={twoFactorCode}
              onChange={setTwoFactorCode}
              isRequired
            />
            {props.twoFactorError && (
              <p className="ui-field-error" role="alert">
                {t(`account.error.${props.twoFactorError}`)}
              </p>
            )}
            <Button type="submit" isDisabled={props.twoFactorPending}>
              {t('account.twoFactor.confirm')}
            </Button>
          </form>
        ) : props.twoFactorEnabled ? (
          <>
            <p>{t('account.twoFactor.enabled')}</p>
            {props.twoFactorError && (
              <p className="ui-field-error" role="alert">
                {t(`account.error.${props.twoFactorError}`)}
              </p>
            )}
            <form onSubmit={submitDisable} noValidate aria-busy={props.twoFactorPending}>
              <TextField
                label={t('account.twoFactor.password')}
                name="two-factor-disable-password"
                type="password"
                autoComplete="current-password"
                value={disablePassword}
                onChange={setDisablePassword}
                isRequired
              />
              <TextField
                label={t('account.twoFactor.code')}
                name="two-factor-disable-code"
                inputMode="numeric"
                autoComplete="one-time-code"
                value={disableCode}
                onChange={setDisableCode}
                isRequired
              />
              <Button type="submit" variant="danger" isDisabled={props.twoFactorPending}>
                {t('account.twoFactor.disable')}
              </Button>
            </form>
            <form onSubmit={submitRegeneration} noValidate aria-busy={props.twoFactorPending}>
              <TextField
                label={t('account.twoFactor.code')}
                name="two-factor-regenerate-code"
                inputMode="numeric"
                autoComplete="one-time-code"
                value={regenerationCode}
                onChange={setRegenerationCode}
                isRequired
              />
              <Button type="submit" variant="secondary" isDisabled={props.twoFactorPending}>
                {t('account.twoFactor.regenerate')}
              </Button>
            </form>
          </>
        ) : (
          <>
            <p>{t('account.twoFactor.disabled')}</p>
            {props.twoFactorError && (
              <p className="ui-field-error" role="alert">
                {t(`account.error.${props.twoFactorError}`)}
              </p>
            )}
            <Button
              type="button"
              onPress={props.onEnrollTwoFactor}
              isDisabled={props.twoFactorPending}
            >
              {t('account.twoFactor.enable')}
            </Button>
          </>
        )}
      </section>
    </section>
  );
}
