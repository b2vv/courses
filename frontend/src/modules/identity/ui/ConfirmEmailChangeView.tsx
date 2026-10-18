import { useTranslation } from 'react-i18next';
import type { SessionErrorCode } from '../domain/identity';
import { Link } from '../../../presentation/ui';

export function ConfirmEmailChangeView({
  isPending,
  completed,
  errorCode,
}: {
  readonly isPending: boolean;
  readonly completed: boolean;
  readonly errorCode: SessionErrorCode | undefined;
}) {
  const { t } = useTranslation();
  return (
    <section className="notice" aria-busy={isPending}>
      <p className="eyebrow">{t('account.label')}</p>
      <h1>{completed ? t('account.confirm.successTitle') : t('account.confirm.title')}</h1>
      <p>
        {completed
          ? t('account.confirm.success')
          : t(errorCode === undefined ? 'account.confirm.checking' : `account.error.${errorCode}`)}
      </p>
      {completed && <Link href="/settings/account">{t('learn.video.account')}</Link>}
    </section>
  );
}
