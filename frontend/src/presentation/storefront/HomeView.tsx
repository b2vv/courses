import { Link } from '../ui';
import { useTranslation } from 'react-i18next';

export function HomeView() {
  const { t } = useTranslation();
  return (
    <section className="home-hero">
      <div className="hero-copy">
        <p className="hero-label">{t('home.label')}</p>
        <h1>
          {t('home.title')} <em>production</em>
        </h1>
        <p className="lede">{t('home.description')}</p>
        <Link href="/courses" className="primary-link">
          {t('home.action')} <span aria-hidden="true">↗</span>
        </Link>
      </div>
      <div className="hero-art" aria-hidden="true">
        <span className="hero-code">{'</>'}</span>
        <p>
          {t('home.art')
            .split('\n')
            .map((line) => (
              <span key={line}>
                {line}
                <br />
              </span>
            ))}
        </p>
        <span className="hero-art-caption">CODE · BUILD · UNDERSTAND</span>
      </div>
    </section>
  );
}
