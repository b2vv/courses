import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Button } from '../../../presentation/ui';

export function EmailVerificationBanner({
  isPending,
  onResend,
}: {
  readonly isPending: boolean;
  readonly onResend: () => void;
}) {
  const { t } = useTranslation();
  // Deliberately in-memory: a reload must surface the still-unverified account again.
  const [dismissed, setDismissed] = useState(false);
  if (dismissed) return null;
  return (
    <aside className="verification-banner" aria-label={t('verificationBanner.label')}>
      <p>{t('verificationBanner.message')}</p>
      <Button variant="secondary" isDisabled={isPending} onPress={onResend}>
        {t(isPending ? 'verificationBanner.resending' : 'verificationBanner.resend')}
      </Button>
      <Button variant="secondary" onPress={() => setDismissed(true)}>
        {t('common.close')}
      </Button>
    </aside>
  );
}
