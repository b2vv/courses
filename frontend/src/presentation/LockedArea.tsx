import { useTranslation } from 'react-i18next';
export function LockedArea() {
  const { t } = useTranslation();
  return (
    <section className="notice">
      <h1>{t('locked.title')}</h1>
      <p>{t('locked.text')}</p>
      <a href="/login">{t('locked.action')}</a>
    </section>
  );
}
