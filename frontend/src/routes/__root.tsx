import {
  createRootRouteWithContext,
  HeadContent,
  Outlet,
  Scripts,
  Link,
} from '@tanstack/react-router';
import type { createCatalogDependencies } from '../app/composition/catalog';
import type { ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import { LanguageSwitcher } from '../presentation/LanguageSwitcher';
import { resolveRootIdentity, type RootIdentityContext } from '../app/composition/root-identity';
import '../app/styles/main.css';

type RootRouteContext = RootIdentityContext & {
  catalog: ReturnType<typeof createCatalogDependencies>;
};

export const Route = createRootRouteWithContext<RootRouteContext>()({
  beforeLoad: resolveRootIdentity,
  head: () => ({
    meta: [
      { charSet: 'utf-8' },
      { name: 'viewport', content: 'width=device-width, initial-scale=1' },
      { title: 'CourseForge — навчання через практику' },
      {
        name: 'description',
        content: 'Курси для розробників: архітектура, інструменти й практика.',
      },
    ],
  }),
  shellComponent: Document,
  component: Layout,
  notFoundComponent: NotFound,
  errorComponent: PageError,
});

function NotFound() {
  const { t } = useTranslation();
  return (
    <section className="notice">
      <h1>{t('page.notFound')}</h1>
      <a href="/courses">{t('catalog.goToCatalog')}</a>
    </section>
  );
}
function PageError() {
  const { t } = useTranslation();
  return (
    <section className="notice" role="alert">
      <h1>{t('page.failure')}</h1>
      <p>{t('page.failureHint')}</p>
      <form method="get">
        <button type="submit">{t('common.retry')}</button>
      </form>
    </section>
  );
}

function Document({ children }: { children: ReactNode }) {
  const { i18n } = useTranslation();
  return (
    <html lang={i18n.resolvedLanguage === 'en' ? 'en' : 'uk'}>
      <head>
        <HeadContent />
      </head>
      <body>
        {children}
        <Scripts />
      </body>
    </html>
  );
}

function Layout() {
  const { t } = useTranslation();
  return (
    <>
      <a className="skip-link" href="#main">
        {t('common.skip')}
      </a>
      <header className="site-header">
        <div className="shell header-inner">
          <Link to="/" className="wordmark" aria-label={t('common.homeLabel')}>
            <span aria-hidden="true">{'</>'}</span>CourseForge
          </Link>
          <nav aria-label={t('common.mainNavigation')}>
            <Link to="/courses">{t('common.catalog')}</Link>
            <Link to="/login">{t('common.login')}</Link>
            <LanguageSwitcher />
          </nav>
        </div>
      </header>
      <main id="main" className="shell main-content" tabIndex={-1}>
        <Outlet />
      </main>
      <footer className="shell site-footer">
        <span>CourseForge</span>
        <span>{t('common.tagline')}</span>
      </footer>
    </>
  );
}
