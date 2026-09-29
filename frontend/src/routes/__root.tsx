import {
  createRootRouteWithContext,
  HeadContent,
  Outlet,
  Scripts,
  Link,
  useNavigate,
  useRouterState,
} from '@tanstack/react-router';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import type { createCatalogDependencies } from '../app/composition/catalog';
import type { AdminDependencies } from '../app/composition/admin';
import type { LearnDependencies } from '../app/composition/learn';
import type { TeachDependencies } from '../app/composition/teach';
import { useEffect, useState, type ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import { forgetOtherAuthors } from '../application/transcript/recovery';
import { sessionQueryKey } from '../app/composition/identity-queries';
import type { TranscriptHistoryStore } from '../ports/TranscriptHistoryStore';
import { LanguageSwitcher } from '../presentation/LanguageSwitcher';
import { IdentityHeader } from '../presentation/identity/IdentityHeader';
import { Toast } from '../presentation/ui';
import { resolveRootIdentity, type RootIdentityContext } from '../app/composition/root-identity';
import '../app/styles/main.css';

type RootRouteContext = RootIdentityContext & {
  admin: AdminDependencies;
  catalog: ReturnType<typeof createCatalogDependencies>;
  learn: LearnDependencies;
  teach: TeachDependencies;
  transcriptHistory: TranscriptHistoryStore;
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
  const { identityDependencies, transcriptHistory } = Route.useRouteContext();
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const [signOutFailed, setSignOutFailed] = useState(false);
  const session = useQuery({ ...identityDependencies.current(), refetchOnMount: false });
  const isStorefront = useRouterState({
    select: (state) =>
      state.location.pathname === '/' || state.location.pathname.startsWith('/courses'),
  });

  // Unsaved editor lines are kept locally so a reload cannot lose them, and they outlive the
  // session on disk. The author is in their key, so nobody is ever OFFERED another person's text
  // - but on a shared machine "not offered" is not "not there", so a change of viewer removes
  // the rest. Signing out is a change of viewer too, which is why this is not tied to signing in.
  //
  // Only a SETTLED identity may do this. A pending or failed query leaves `data` undefined, and
  // acting on that would let one flaky request delete work the author has not saved.
  const viewerId = session.isSuccess ? (session.data?.id ?? null) : undefined;
  useEffect(() => {
    if (viewerId === undefined) return;
    void forgetOtherAuthors(transcriptHistory, viewerId);
  }, [viewerId, transcriptHistory]);
  const signOut = useMutation({
    mutationFn: () =>
      identityDependencies.session.logOut().then((result) => identityDependencies.unwrap(result)),
    onMutate: () => setSignOutFailed(false),
    onSettled: async (_data, error) => {
      setSignOutFailed(error !== null);
      try {
        await navigate({ to: '/courses' });
      } finally {
        queryClient.setQueryData(sessionQueryKey, null);
      }
    },
  });
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
            <IdentityHeader
              user={session.data ?? null}
              isSigningOut={signOut.isPending}
              onSignOut={() => signOut.mutate()}
            />
            <LanguageSwitcher />
          </nav>
        </div>
      </header>
      {signOutFailed && (
        <div className="shell sign-out-message">
          <Toast message={t('identity.signOutFailure')} onDismiss={() => setSignOutFailed(false)} />
        </div>
      )}
      <main id="main" className="shell main-content" tabIndex={-1}>
        <Outlet />
      </main>
      {isStorefront ? (
        <footer className="site-footer site-footer--storefront">
          <div className="shell footer-grid">
            <div className="footer-brand">
              <Link to="/" className="wordmark" aria-label={t('common.homeLabel')}>
                <span aria-hidden="true">{'</>'}</span>CourseForge
              </Link>
              <p>{t('common.tagline')}</p>
            </div>
            <nav aria-label={t('common.footerLearning')} className="footer-links">
              <strong>{t('common.footerLearning')}</strong>
              <Link to="/courses">{t('common.catalog')}</Link>
              <Link to="/learn">{t('common.myLearning')}</Link>
            </nav>
          </div>
        </footer>
      ) : (
        <footer className="shell site-footer">
          <span>CourseForge</span>
          <span>{t('common.tagline')}</span>
        </footer>
      )}
    </>
  );
}
