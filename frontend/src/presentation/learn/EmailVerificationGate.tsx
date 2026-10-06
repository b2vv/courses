import { useTranslation } from 'react-i18next';
import type { RootIdentity } from '../../app/composition/root-identity';
import { Button, Link } from '../ui';

export function EmailVerificationGate({
  isPending,
  onResend,
}: {
  readonly isPending: boolean;
  readonly onResend: () => void;
}) {
  const { t } = useTranslation();
  return (
    <section className="lesson-verification-gate" aria-labelledby="email-verification-gate-title">
      <h2 id="email-verification-gate-title">{t('learn.video.emailVerificationTitle')}</h2>
      <p>{t('learn.video.emailVerificationHint')}</p>
      <div className="cluster">
        <Button isDisabled={isPending} onPress={onResend}>
          {t(isPending ? 'verificationBanner.resending' : 'verificationBanner.resend')}
        </Button>
        <Link href="/settings/account">{t('learn.video.account')}</Link>
      </div>
    </section>
  );
}

export function shouldGateLessonPlayback(identity: RootIdentity): boolean {
  return identity.status === 'signed-in' && !identity.user.emailVerified;
}
