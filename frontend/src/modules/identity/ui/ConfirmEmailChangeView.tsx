import { useTranslation } from 'react-i18next';
import type { SessionErrorCode } from '../domain/identity';
import { isUncertainWriteError } from '../domain/identity';
import { Button, Link } from '../../../presentation/ui';

/** What `/email/change/confirm` shows (design-pass ACCT, plan §18 V3 rows E0–E5). */
export type ConfirmEmailChangeState =
  | { readonly kind: 'waiting' }
  | { readonly kind: 'checking' }
  | { readonly kind: 'changed' }
  | { readonly kind: 'taken' }
  | { readonly kind: 'link-failed' }
  | { readonly kind: 'failed'; readonly code: SessionErrorCode };

export function ConfirmEmailChangeView({
  state,
  retryPending,
  onRetry,
}: {
  readonly state: ConfirmEmailChangeState;
  readonly retryPending: boolean;
  readonly onRetry: () => void;
}) {
  const { t } = useTranslation();
  const title = {
    waiting: 'account.confirm.title',
    checking: 'account.confirm.title',
    changed: 'account.confirm.successTitle',
    taken: 'account.confirm.takenTitle',
    'link-failed': 'verify.linkFailedTitle',
    failed: 'verify.failedTitle',
  }[state.kind];
  const toSettings = (
    <Link href="/settings/account" className="ui-button ui-button--primary">
      {t(state.kind === 'changed' ? 'learn.video.account' : 'account.confirm.toSettings')}
    </Link>
  );
  return (
    <section className="notice" aria-busy={state.kind === 'waiting' || state.kind === 'checking'}>
      <p className="eyebrow">{t('verify.label')}</p>
      <h1>{t(title)}</h1>
      {state.kind === 'waiting' && <p>{t('verify.waiting')}</p>}
      {state.kind === 'checking' && <p>{t('account.confirm.checking')}</p>}
      {state.kind === 'changed' && (
        <>
          <p>{t('account.confirm.success')}</p>
          {toSettings}
        </>
      )}
      {state.kind === 'taken' && (
        <>
          <p role="alert">{t('account.confirm.takenText')}</p>
          {toSettings}
        </>
      )}
      {state.kind === 'link-failed' && (
        <>
          <p>{t('account.confirm.linkFailedText')}</p>
          {toSettings}
        </>
      )}
      {state.kind === 'failed' && (
        <>
          <p role="alert">
            {t(`account.error.${state.code}`)}
            {isUncertainWriteError(state.code) && ` ${t('account.confirm.maybeChanged')}`}
          </p>
          <div className="cluster">
            {/* The origin is checked before the token: a retry from this page would be refused
                again (plan §18 E5). */}
            {state.code !== 'origin_forbidden' && (
              <Button variant="secondary" aria-disabled={retryPending} onPress={onRetry}>
                {t('verify.retry')}
              </Button>
            )}
            {toSettings}
          </div>
        </>
      )}
    </section>
  );
}
