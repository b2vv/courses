import { useTranslation } from 'react-i18next';
import type { ResendOutcome } from '../domain/identity';
import { Link } from '../../../presentation/ui';

/** The one answer to a resend press, shared by the verify page, the banner and the lesson gate. */
export function ResendResult({ outcome }: { readonly outcome: ResendOutcome | undefined }) {
  const { t } = useTranslation();
  if (outcome === undefined) return null;
  if (outcome === 'accepted')
    return <output className="form-success">{t('verify.resendAccepted')}</output>;
  if (outcome === 'session_ended')
    return (
      <p className="ui-field-error" role="alert">
        {t('verify.resendSessionEnded')}{' '}
        {/* A full load: a move inside the app would read the cached viewer, and /login would
            offer «Продовжити» instead of the sign-in form (design-pass ACCT V1). */}
        <Link href="/login" routerOptions={{ reloadDocument: true }}>
          {t('common.login')}
        </Link>
      </p>
    );
  return (
    <p className="ui-field-error" role="alert">
      {t(outcome === 'rate_limited' ? 'verify.resendLimited' : 'verify.resendFailed')}
    </p>
  );
}
