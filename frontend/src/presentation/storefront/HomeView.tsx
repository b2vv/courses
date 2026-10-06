import { Link } from '../ui';
import { ValueStrip } from './ValueStrip';
import { useTranslation } from 'react-i18next';

export function HomeView({ notice }: { readonly notice?: 'verify-email' | undefined }) {
  const { t } = useTranslation();
  return (
    <div className="storefront-home">
      <section className="home-hero">
        <div className="hero-copy">
          {notice === 'verify-email' && <output>{t('home.verifyEmailNotice')}</output>}
          <p className="ui-pill ui-pill--accent">{t('home.label')}</p>
          <h1>
            {t('home.title')} <em>production</em>
          </h1>
          <p className="lede">{t('home.description')}</p>
          <div className="hero-actions">
            <Link href="/courses" className="primary-link">
              {t('home.action')} <span aria-hidden="true">↗</span>
            </Link>
            <Link href="#values" className="secondary-link">
              <span aria-hidden="true">▶</span> {t('home.secondaryAction')}
            </Link>
          </div>
          <dl className="hero-facts" aria-label={t('home.factsLabel')}>
            <div>
              <dt>{t('home.factPracticeLabel')}</dt>
              <dd>{t('home.factPracticeValue')}</dd>
            </div>
            <div>
              <dt>{t('home.factPaceLabel')}</dt>
              <dd>{t('home.factPaceValue')}</dd>
            </div>
          </dl>
        </div>
        <div className="hero-art" aria-hidden="true">
          <div className="hero-preview-card">
            <div className="hero-preview-cover">
              <p className="hero-preview-label">Популярний курс</p>
              <p className="hero-preview-title">React Architecture Pro</p>
            </div>
          </div>
          <div className="hero-progress-card">
            <div className="hero-progress-header">
              <span>Ваш прогрес</span>
              <strong>68%</strong>
            </div>
            <div className="hero-progress-bar">
              <div className="hero-progress-fill" />
            </div>
            <p className="hero-progress-meta">Модуль 7 з 10 · Domain boundaries</p>
          </div>
        </div>
      </section>
      <ValueStrip />
    </div>
  );
}
