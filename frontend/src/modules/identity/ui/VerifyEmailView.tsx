import { useTranslation } from 'react-i18next';
import type { SessionErrorCode } from '../domain/identity';
import { Button } from '../../../presentation/ui';

type Props = {
  readonly isPending: boolean;
  readonly completed: boolean;
  readonly errorCode: SessionErrorCode | undefined;
  readonly onResend: () => void;
};
export function VerifyEmailView({ isPending, completed, errorCode, onResend }: Props) {
  const { t } = useTranslation();
  const success = completed;
  return (
    <section className="notice" aria-busy={isPending}>
      <p className="eyebrow">{t('verify.label')}</p>
      <h1>{success ? t('verify.successTitle') : t('verify.title')}</h1>
      <p>{success ? t('verify.success') : t(errorCode ? 'verify.expired' : 'verify.checking')}</p>
      {errorCode && (
        <Button onPress={onResend} isDisabled={isPending}>
          {t(isPending ? 'verify.resending' : 'verify.resend')}
        </Button>
      )}
    </section>
  );
}
