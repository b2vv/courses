import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import type { ResendOutcome } from '../domain/identity';
import { Button } from '../../../presentation/ui';
import { ResendResult } from './ResendResult';

export function EmailVerificationBanner({
  isPending,
  outcome,
  onResend,
}: {
  readonly isPending: boolean;
  readonly outcome?: ResendOutcome | undefined;
  readonly onResend: () => void;
}) {
  const { t } = useTranslation();
  // Deliberately in-memory: a reload must surface the still-unverified account again.
  const [dismissed, setDismissed] = useState(false);
  if (dismissed) return null;
  return (
    <aside className="verification-banner" aria-label={t('verificationBanner.label')}>
      <p>{t('verificationBanner.message')}</p>
      {/* isPending, not isDisabled: a native `disabled` on the focused button drops focus. */}
      <Button variant="secondary" isPending={isPending} onPress={onResend}>
        {t(isPending ? 'verificationBanner.resending' : 'verificationBanner.resend')}
      </Button>
      <Button variant="secondary" onPress={() => setDismissed(true)}>
        {t('common.close')}
      </Button>
      <ResendResult outcome={outcome} />
    </aside>
  );
}
