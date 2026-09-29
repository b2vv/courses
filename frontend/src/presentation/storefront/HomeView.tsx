import { Link } from '../ui';
import { useTranslation } from 'react-i18next';

export function HomeView() {
  const { t } = useTranslation();
  return (
    <div className="storefront-home">
      <section className="home-hero">
        <div className="hero-copy">
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
        <div className="hero-art ui-grid-surface" aria-hidden="true">
          <div className="hero-code-card ui-card ui-card--floating ui-card--glow">
            <div className="hero-code-heading">
              <span className="ui-icon-tile">{'</>'}</span>
              <span className="ui-mono">course.ts</span>
            </div>
            <code>
              <span>learn(</span>
              <strong>{t('home.codePractice')}</strong>
              <span>)</span>
              <span>build(</span>
              <strong>{t('home.codeSystem')}</strong>
              <span>)</span>
              <span>understand(</span>
              <strong>{t('home.codeWhy')}</strong>
              <span>)</span>
            </code>
            <p>{t('home.art')}</p>
          </div>
        </div>
      </section>
      <section id="values" className="value-strip" aria-labelledby="values-heading">
        <h2 id="values-heading" className="ui-visually-hidden">
          {t('home.valuesTitle')}
        </h2>
        {(['practice', 'review', 'access'] as const).map((value) => (
          <article key={value} className="value-item">
            <span className="ui-icon-tile" aria-hidden="true">
              {value === 'practice' ? '⌘' : value === 'review' ? '✓' : '∞'}
            </span>
            <div>
              <h3>{t(`home.values.${value}.title`)}</h3>
              <p>{t(`home.values.${value}.description`)}</p>
            </div>
          </article>
        ))}
      </section>
    </div>
  );
}
