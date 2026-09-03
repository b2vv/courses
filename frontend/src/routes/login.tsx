import { createFileRoute } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
export const Route = createFileRoute('/login')({ component: Login });
function Login() {
  const { t } = useTranslation();
  return (
    <section className="notice">
      <p className="eyebrow">{t('login.label')}</p>
      <h1>{t('login.title')}</h1>
      <p>{t('login.text')}</p>
      <a href="/courses">{t('login.action')}</a>
    </section>
  );
}
