import { useEffect, useRef, useState, type FormEvent } from 'react';
import { useTranslation } from 'react-i18next';
import { QRCodeSVG } from 'qrcode.react';
import type { SessionErrorCode, TwoFactorAction } from '../domain/identity';
import type { RecoveryCodes, TwoFactorEnrollment } from '../domain/identity';
import { isUncertainWriteError } from '../domain/identity';
import { Button, TextArea, TextField } from '../../../presentation/ui';

type Props = {
  readonly currentEmail: string | undefined;
  readonly passwordPending: boolean;
  readonly emailPending: boolean;
  readonly passwordError: SessionErrorCode | undefined;
  readonly emailError: SessionErrorCode | undefined;
  /** When the last successful change was submitted; a new value clears the form (plan §18 V10). */
  readonly passwordSavedAt: number | undefined;
  readonly emailRequestedAt: number | undefined;
  readonly twoFactorEnabled: boolean;
  readonly twoFactorPending: boolean;
  readonly twoFactorError:
    | { readonly action: TwoFactorAction; readonly code: SessionErrorCode }
    | undefined;
  /** An uncertain or conflicting 2FA result: the shown state may be out of date (plan §18 V5a). */
  readonly twoFactorStateUnknown: boolean;
  readonly enrollment: TwoFactorEnrollment | undefined;
  readonly recovery:
    | { readonly codes: RecoveryCodes; readonly kind: 'enabled' | 'regenerated' }
    | undefined;
  readonly onChangePassword: (input: { oldPassword: string; newPassword: string }) => void;
  readonly onChangeEmail: (input: { newEmail: string }) => void;
  readonly onEnrollTwoFactor: () => void;
  readonly onConfirmTwoFactor: (code: string) => void;
  readonly onCancelTwoFactor: () => void;
  readonly onDisableTwoFactor: (input: { password: string; code: string }) => void;
  readonly onRegenerateRecoveryCodes: (code: string) => void;
  readonly onDismissRecoveryCodes: () => void;
};

type Phase = 'recovery' | 'setup' | 'enabled' | 'disabled';
type UserStep = 'enroll' | 'confirm' | 'done' | 'cancel' | 'disable' | 'regenerate';

function reloadPage() {
  window.location.reload();
}

export function AccountSettingsView(props: Props) {
  const { t } = useTranslation();
  const [oldPassword, setOldPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [email, setEmail] = useState('');
  const [twoFactorCode, setTwoFactorCode] = useState('');
  const [disablePassword, setDisablePassword] = useState('');
  const [disableCode, setDisableCode] = useState('');
  const [regenerationCode, setRegenerationCode] = useState('');

  // The fields are read-only while a write is in flight, so clearing after it cannot erase
  // anything typed later (plan §18 V10). Cleared during render when a new success arrives.
  const [passwordClearedAt, setPasswordClearedAt] = useState(props.passwordSavedAt);
  if (props.passwordSavedAt !== passwordClearedAt) {
    setPasswordClearedAt(props.passwordSavedAt);
    if (props.passwordSavedAt !== undefined) {
      setOldPassword('');
      setNewPassword('');
    }
  }
  const [emailClearedAt, setEmailClearedAt] = useState(props.emailRequestedAt);
  if (props.emailRequestedAt !== emailClearedAt) {
    setEmailClearedAt(props.emailRequestedAt);
    if (props.emailRequestedAt !== undefined) setEmail('');
  }

  // Focus follows the card's content when a user action replaces it (plan §18 V5). The first
  // render and errors that leave the form in place move nothing.
  const phase: Phase = props.recovery
    ? 'recovery'
    : props.enrollment
      ? 'setup'
      : props.twoFactorEnabled
        ? 'enabled'
        : 'disabled';
  const step = useRef<UserStep | undefined>(undefined);
  useEffect(() => {
    const last = step.current;
    if (last === undefined) return;
    const target =
      phase === 'setup'
        ? 'two-factor-setup'
        : phase === 'disabled' && last === 'cancel'
          ? 'two-factor-enable'
          : 'two-factor-status';
    const element = document.getElementById(target);
    if (element === null) return;
    step.current = undefined;
    element.focus();
  }, [phase, props.recovery]);
  useEffect(() => {
    // A conflict or an uncertain write that left the form in place: nothing to move.
    if (props.twoFactorError !== undefined) step.current = undefined;
  }, [props.twoFactorError]);

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
    if (props.twoFactorPending || !event.currentTarget.checkValidity()) return;
    step.current = 'confirm';
    props.onConfirmTwoFactor(twoFactorCode);
  }
  function submitDisable(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (props.twoFactorPending || !event.currentTarget.checkValidity()) return;
    step.current = 'disable';
    props.onDisableTwoFactor({ password: disablePassword, code: disableCode });
  }
  function submitRegeneration(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (props.twoFactorPending || !event.currentTarget.checkValidity()) return;
    step.current = 'regenerate';
    props.onRegenerateRecoveryCodes(regenerationCode);
  }
  function errorFor(action: TwoFactorAction) {
    const error = props.twoFactorError;
    if (error === undefined || error.action !== action) return null;
    const regenerationUnknown = action === 'regenerate' && isUncertainWriteError(error.code);
    return (
      <p className="ui-field-error" role="alert">
        {regenerationUnknown
          ? t('account.twoFactor.regenerateUncertain')
          : t(`account.error.${error.code}`)}
      </p>
    );
  }
  const stateNotice = props.twoFactorStateUnknown && (
    <div className="two-factor-unknown">
      <p>{t('account.twoFactor.stateUnknown')}</p>
      <Button variant="secondary" onPress={reloadPage}>
        {t('verify.reload')}
      </Button>
    </div>
  );

  return (
    <section className="account-settings">
      <div>
        <p className="eyebrow">{t('account.label')}</p>
        <h1>{t('account.title')}</h1>
      </div>
      <section className="account-settings-card" aria-labelledby="change-password-title">
        <h2 id="change-password-title">{t('account.password.title')}</h2>
        <form onSubmit={submitPassword} noValidate aria-busy={props.passwordPending}>
          {props.passwordSavedAt !== undefined && (
            <output className="form-success">{t('account.password.success')}</output>
          )}
          {props.passwordError && (
            <p className="ui-field-error" role="alert">
              {t(`account.error.${props.passwordError}`)}
              {isUncertainWriteError(props.passwordError) && ` ${t('account.password.uncertain')}`}
            </p>
          )}
          <TextField
            label={t('account.password.old')}
            name="old-password"
            type="password"
            autoComplete="current-password"
            value={oldPassword}
            onChange={setOldPassword}
            isReadOnly={props.passwordPending}
            isRequired
          />
          <TextField
            label={t('account.password.new')}
            name="new-password"
            type="password"
            autoComplete="new-password"
            value={newPassword}
            onChange={setNewPassword}
            isReadOnly={props.passwordPending}
            isRequired
          />
          <Button type="submit" isPending={props.passwordPending}>
            {t(props.passwordPending ? 'account.password.saving' : 'account.password.submit')}
          </Button>
        </form>
      </section>
      <section className="account-settings-card" aria-labelledby="change-email-title">
        <h2 id="change-email-title">{t('account.email.title')}</h2>
        {props.currentEmail !== undefined && (
          <p className="account-current-email">
            {t('account.email.current')} <strong>{props.currentEmail}</strong>
          </p>
        )}
        <form onSubmit={submitEmail} noValidate aria-busy={props.emailPending}>
          {props.emailRequestedAt !== undefined && (
            <output className="form-success">{t('account.email.success')}</output>
          )}
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
            isReadOnly={props.emailPending}
            isRequired
          />
          <Button type="submit" isPending={props.emailPending}>
            {t(props.emailPending ? 'account.email.saving' : 'account.email.submit')}
          </Button>
        </form>
      </section>
      {/* The 2FA buttons are replaced by their own success, so they are busy through
          aria-disabled and the dispatch guards, not isPending: a focused pending React Aria
          Button announces itself by its id, and the announcer keeps that reference for 7 s
          after the button is gone (axe role-img-alt). */}
      <section className="account-settings-card" aria-labelledby="two-factor-title">
        <h2 id="two-factor-title">{t('account.twoFactor.title')}</h2>
        {props.recovery ? (
          <>
            <p id="two-factor-status" tabIndex={-1}>
              {t(
                props.recovery.kind === 'enabled'
                  ? 'account.twoFactor.enabled'
                  : 'account.twoFactor.newCodes',
              )}
            </p>
            <RecoveryCodesBlock
              codes={props.recovery.codes}
              onDone={() => {
                step.current = 'done';
                props.onDismissRecoveryCodes();
              }}
            />
          </>
        ) : props.enrollment ? (
          <form onSubmit={submitConfirmation} noValidate aria-busy={props.twoFactorPending}>
            <p id="two-factor-setup" tabIndex={-1}>
              {t('account.twoFactor.setup')}
            </p>
            <QRCodeSVG
              value={props.enrollment.otpauthUri}
              aria-label={t('account.twoFactor.qrAlt')}
            />
            <TextArea
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
              isReadOnly={props.twoFactorPending}
              isRequired
            />
            {errorFor('confirm')}
            {stateNotice}
            <div className="cluster">
              <Button type="submit" aria-disabled={props.twoFactorPending}>
                {t('account.twoFactor.confirm')}
              </Button>
              {/* Hidden once the result is uncertain: 2FA may already be on (plan §18 V5a). */}
              {!props.twoFactorStateUnknown && (
                <Button
                  type="button"
                  variant="secondary"
                  onPress={() => {
                    if (props.twoFactorPending) return;
                    step.current = 'cancel';
                    setTwoFactorCode('');
                    props.onCancelTwoFactor();
                  }}
                >
                  {t('account.twoFactor.cancel')}
                </Button>
              )}
            </div>
          </form>
        ) : props.twoFactorEnabled ? (
          <>
            <p id="two-factor-status" tabIndex={-1}>
              {t('account.twoFactor.enabled')}
            </p>
            {stateNotice}
            <form
              onSubmit={submitRegeneration}
              noValidate
              aria-busy={props.twoFactorPending}
              aria-labelledby="two-factor-codes-title"
            >
              <h3 id="two-factor-codes-title">{t('account.twoFactor.codesTitle')}</h3>
              <p>{t('account.twoFactor.codesHint')}</p>
              <TextField
                label={t('account.twoFactor.code')}
                name="two-factor-regenerate-code"
                inputMode="numeric"
                autoComplete="one-time-code"
                value={regenerationCode}
                onChange={setRegenerationCode}
                isReadOnly={props.twoFactorPending}
                isRequired
              />
              {errorFor('regenerate')}
              <Button type="submit" variant="secondary" aria-disabled={props.twoFactorPending}>
                {t('account.twoFactor.regenerate')}
              </Button>
            </form>
            <form
              onSubmit={submitDisable}
              noValidate
              aria-busy={props.twoFactorPending}
              aria-labelledby="two-factor-disable-title"
            >
              <h3 id="two-factor-disable-title">{t('account.twoFactor.disable')}</h3>
              <TextField
                label={t('account.twoFactor.password')}
                name="two-factor-disable-password"
                type="password"
                autoComplete="current-password"
                value={disablePassword}
                onChange={setDisablePassword}
                isReadOnly={props.twoFactorPending}
                isRequired
              />
              <TextField
                label={t('account.twoFactor.code')}
                name="two-factor-disable-code"
                inputMode="numeric"
                autoComplete="one-time-code"
                value={disableCode}
                onChange={setDisableCode}
                isReadOnly={props.twoFactorPending}
                isRequired
              />
              {errorFor('disable')}
              <Button type="submit" variant="danger" aria-disabled={props.twoFactorPending}>
                {t('account.twoFactor.disable')}
              </Button>
            </form>
          </>
        ) : (
          <>
            <p id="two-factor-status" tabIndex={-1}>
              {t('account.twoFactor.disabled')}
            </p>
            {errorFor('enroll')}
            {stateNotice}
            <Button
              id="two-factor-enable"
              type="button"
              onPress={() => {
                if (props.twoFactorPending) return;
                step.current = 'enroll';
                props.onEnrollTwoFactor();
              }}
              aria-disabled={props.twoFactorPending}
            >
              {t('account.twoFactor.enable')}
            </Button>
          </>
        )}
      </section>
    </section>
  );
}

/**
 * The one-time code list. «Скопійовано» lives here, so «Готово» (which unmounts the block) also
 * drops it, and a new set never starts as already copied (plan §18 V14).
 */
function RecoveryCodesBlock({
  codes,
  onDone,
}: {
  readonly codes: RecoveryCodes;
  readonly onDone: () => void;
}) {
  const { t } = useTranslation();
  const [copy, setCopy] = useState<'copied' | 'failed'>();
  const mounted = useRef(true);
  useEffect(() => {
    mounted.current = true;
    return () => {
      mounted.current = false;
    };
  }, []);
  async function copyCodes() {
    try {
      if (!navigator.clipboard) throw new Error('no clipboard');
      await navigator.clipboard.writeText(codes.recoveryCodes.join('\n'));
      if (mounted.current) setCopy('copied');
    } catch {
      if (mounted.current) setCopy('failed');
    }
  }
  return (
    <div className="two-factor-recovery" aria-live="polite">
      <p>{t('account.twoFactor.recovery')}</p>
      <ul aria-label={t('account.twoFactor.recovery')}>
        {codes.recoveryCodes.map((recoveryCode) => (
          <li key={recoveryCode}>
            <code>{recoveryCode}</code>
          </li>
        ))}
      </ul>
      <div className="cluster">
        <Button type="button" variant="secondary" onPress={() => void copyCodes()}>
          {t('account.twoFactor.copy')}
        </Button>
        <Button type="button" onPress={onDone}>
          {t('account.twoFactor.done')}
        </Button>
      </div>
      {copy === 'copied' && (
        <output className="form-success">{t('account.twoFactor.copied')}</output>
      )}
      {copy === 'failed' && (
        <p className="ui-field-error" role="alert">
          {t('account.twoFactor.copyFailed')}
        </p>
      )}
    </div>
  );
}
