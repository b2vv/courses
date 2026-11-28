import { useState } from 'react';
import type { FormEvent } from 'react';
import { useTranslation } from 'react-i18next';
import type { ResendOutcome, SessionErrorCode } from '../domain/identity';
import { isUncertainWriteError } from '../domain/identity';
import { Button, Link, TextField } from '../../../presentation/ui';
import { ResendResult } from './ResendResult';

/**
 * What `/verify-email` shows (design-pass ACCT, plan §18 V2 rows S0–S5). The route decides the
 * state from the link attempt and one session read started after its result; the view only draws.
 */
export type VerifyEmailState =
  /** S0: the link request is paused until the network returns; it has not reached the server. */
  | { readonly kind: 'waiting' }
  /** S1: the link request is in flight. */
  | { readonly kind: 'checking' }
  /** S2: the link worked. `next` waits for the session read; `offline` while it is paused. */
  | {
      readonly kind: 'verified';
      readonly next: 'pending' | 'learn' | 'login';
      readonly offline: boolean;
    }
  /** S3: the link failed, but the signed-in viewer's own address is already verified. */
  | { readonly kind: 'already-verified' }
  /** S4: the link is missing, unknown or used. `action` waits for the session read. */
  | {
      readonly kind: 'link-failed';
      readonly action: 'pending' | 'guest' | 'signed-in' | 'reload';
      readonly offline: boolean;
    }
  /** S5: any other failure of the link request. */
  | { readonly kind: 'failed'; readonly code: SessionErrorCode };

type Props = {
  readonly state: VerifyEmailState;
  readonly resendPending: boolean;
  readonly resendOutcome: ResendOutcome | undefined;
  readonly retryPending: boolean;
  readonly onResend: (email?: string) => void;
  readonly onRetry: () => void;
};

function reloadPage() {
  window.location.reload();
}

export function VerifyEmailView({
  state,
  resendPending,
  resendOutcome,
  retryPending,
  onResend,
  onRetry,
}: Props) {
  const { t } = useTranslation();
  const [email, setEmail] = useState('');
  function submitEmail(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    if (event.currentTarget.checkValidity()) onResend(email);
  }
  const busy = state.kind === 'waiting' || state.kind === 'checking';
  const title = {
    waiting: 'verify.title',
    checking: 'verify.title',
    verified: 'verify.successTitle',
    'already-verified': 'verify.alreadyTitle',
    'link-failed': 'verify.linkFailedTitle',
    failed: 'verify.failedTitle',
  }[state.kind];
  return (
    <section className="notice" aria-busy={busy}>
      <p className="eyebrow">{t('verify.label')}</p>
      <h1>{t(title)}</h1>
      {state.kind === 'waiting' && <p>{t('verify.waiting')}</p>}
      {state.kind === 'checking' && <p>{t('verify.checking')}</p>}
      {state.kind === 'verified' && (
        <>
          {state.offline && <p>{t('verify.waiting')}</p>}
          {state.next === 'learn' && (
            <Link href="/learn" className="ui-button ui-button--primary">
              {t('verify.toLearning')}
            </Link>
          )}
          {state.next === 'login' && (
            // A full load: the link is already spent, and the sign-in page must read the
            // session afresh rather than a cached viewer.
            <Link
              href="/login"
              routerOptions={{ reloadDocument: true }}
              className="ui-button ui-button--primary"
            >
              {t('common.login')}
            </Link>
          )}
        </>
      )}
      {state.kind === 'already-verified' && (
        <>
          <p>{t('verify.alreadyText')}</p>
          <Link href="/learn" className="ui-button ui-button--primary">
            {t('verify.toLearning')}
          </Link>
        </>
      )}
      {state.kind === 'link-failed' && (
        <>
          <p>{t('verify.expired')}</p>
          {state.offline && <p>{t('verify.waiting')}</p>}
          {state.action === 'guest' && (
            <form className="notice-action" onSubmit={submitEmail} noValidate>
              <TextField
                label={t('login.email')}
                name="email"
                type="email"
                autoComplete="email"
                maxLength={254}
                value={email}
                onChange={setEmail}
                isRequired
              />
              <Button type="submit" isPending={resendPending}>
                {t(resendPending ? 'verify.resending' : 'verify.resend')}
              </Button>
            </form>
          )}
          {state.action === 'signed-in' && (
            <Button isPending={resendPending} onPress={() => onResend()}>
              {t(resendPending ? 'verify.resending' : 'verify.resend')}
            </Button>
          )}
          {state.action === 'reload' && <Button onPress={reloadPage}>{t('verify.reload')}</Button>}
          <ResendResult outcome={resendOutcome} />
        </>
      )}
      {state.kind === 'failed' && (
        <>
          <p role="alert">
            {t(`account.error.${state.code}`)}
            {isUncertainWriteError(state.code) && ` ${t('verify.maybeVerified')}`}
          </p>
          {state.code === 'origin_forbidden' ? (
            // The server checks the origin before the token: the same request from this page
            // would be refused again, so there is no retry (plan §18 S5).
            <Link href="/" className="ui-button ui-button--primary">
              {t('page.goHome')}
            </Link>
          ) : (
            <Button aria-disabled={retryPending} onPress={onRetry}>
              {t('verify.retry')}
            </Button>
          )}
        </>
      )}
    </section>
  );
}
