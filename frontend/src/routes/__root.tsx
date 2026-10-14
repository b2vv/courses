import {
  createRootRouteWithContext,
  HeadContent,
  Outlet,
  Scripts,
  Link,
  useRouterState,
} from '@tanstack/react-router';
import { useMutation, useQuery } from '@tanstack/react-query';
import type { createCatalogDependencies } from '../app/composition/catalog';
import type { AdminDependencies } from '../app/composition/admin';
import type { LearnDependencies } from '../app/composition/learn';
import type { TeachDependencies } from '../app/composition/teach';
import type { createEnrollmentDependencies } from '../app/composition/enrollment';
import { useEffect, type ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import { forgetOtherAuthors } from '../application/transcript/recovery';
import { useSignOut } from '../app/composition/use-sign-out';
import { canAccessAdministration, canAdministerAccounts } from '../domain/identity/identity';
import { AdminShell, type AdminNavSection } from '../presentation/admin/AdminShell';
import type { TranscriptHistoryStore } from '../ports/TranscriptHistoryStore';
import { LanguageSwitcher } from '../presentation/LanguageSwitcher';
import { IdentityHeader } from '../presentation/identity/IdentityHeader';
import { EmailVerificationBanner } from '../presentation/identity/EmailVerificationBanner';
import { Toast } from '../presentation/ui';
import { resolveRootIdentity, type RootIdentityContext } from '../app/composition/root-identity';
import '../app/styles/main.css';

type RootRouteContext = RootIdentityContext & {
  admin: AdminDependencies;
  catalog: ReturnType<typeof createCatalogDependencies>;
  learn: LearnDependencies;
  teach: TeachDependencies;
  enrollment: ReturnType<typeof createEnrollmentDependencies>;
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
  const session = useQuery({ ...identityDependencies.current(), refetchOnMount: false });
  const isStorefront = useRouterState({
    select: (state) =>
      state.location.pathname === '/' || state.location.pathname.startsWith('/courses'),
  });
  // Client vs admin surface: the admin and instructor workspaces are a separate
  // app with their own shell (AdminShell) and the NorthWind theme; the client
  // header/footer are suppressed there. The admin surface needs BOTH an admin
  // path and a viewer the guards let in: a refused student gets the not-found
  // answer inside the client app, because AdminShell never mounts for them and
  // an admin-surface page would have no chrome at all.
  const onAdminPath = useRouterState({
    select: (state) =>
      state.location.pathname.startsWith('/admin') || state.location.pathname.startsWith('/teach'),
  });
  const onTeachPath = useRouterState({
    select: (state) => state.location.pathname.startsWith('/teach'),
  });
  const surface =
    onAdminPath && session.data != null && canAccessAdministration(session.data)
      ? 'admin'
      : 'client';

  // React Aria portals mount under body, outside .app-shell. Keep admin dialogs
  // and tooltips in the same token scope, and restore it when leaving the area.
  useEffect(() => {
    if (surface !== 'admin') return;
    const previousSurface = document.body.getAttribute('data-surface');
    document.body.setAttribute('data-surface', 'admin');
    return () => {
      if (previousSurface === null) document.body.removeAttribute('data-surface');
      else document.body.setAttribute('data-surface', previousSurface);
    };
  }, [surface]);

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
  const { signOut, isSigningOut, signOutFailed, dismissSignOutFailure } =
    useSignOut(identityDependencies);
  const resendVerification = useMutation({
    mutationFn: () => identityDependencies.session.resend({}).then(identityDependencies.unwrap),
  });

  // The admin side is its own root layout: AdminShell (sidebar + top bar + the
  // #main skip target) replaces the client header and footer. It lives here, not
  // in the /admin and /teach routes, because those are client-only (ssr: false):
  // rendered at the root, the signed-in chrome is in the server HTML, so the
  // viewer never sees an anonymous frame before hydration.
  if (surface === 'admin' && session.data != null) {
    const viewer = session.data;
    const navSections: AdminNavSection[] = onTeachPath
      ? [
          {
            items: [
              { to: '/teach', label: t('common.navModeration'), exact: true, icon: 'moderation' },
            ],
          },
          {
            label: t('common.navGroupPeople'),
            items: [{ to: '/teach/learners', label: t('common.navLearners'), icon: 'people' }],
          },
        ]
      : [
          {
            items: [
              { to: '/admin', label: t('common.navCourses'), exact: true, icon: 'overview' },
              ...(canAdministerAccounts(viewer)
                ? [
                    {
                      to: '/admin/analytics',
                      label: t('common.navAnalytics'),
                      icon: 'analytics' as const,
                    },
                  ]
                : []),
            ],
          },
          {
            label: t('common.navGroupPeople'),
            items: [
              { to: '/admin/learners', label: t('common.navLearners'), icon: 'people' },
              ...(canAdministerAccounts(viewer)
                ? [
                    {
                      to: '/admin/users',
                      label: t('adminUsers.title'),
                      icon: 'accounts' as const,
                    },
                  ]
                : []),
            ],
          },
          {
            label: t('common.navModeration'),
            items: [
              {
                to: '/admin/participant-reports',
                label: t('common.navParticipantReports'),
                icon: 'moderation',
              },
              {
                to: '/admin/refunds',
                label: t('common.navRefunds'),
                icon: 'refunds',
              },
            ],
          },
        ];
    return (
      <div className="app-shell" data-surface="admin">
        <a className="skip-link" href="#main">
          {t('common.skip')}
        </a>
        <AdminShell
          title={onTeachPath ? t('common.teachArea') : t('administration.title')}
          navLabel={onTeachPath ? t('common.teachArea') : t('common.adminArea')}
          navSections={navSections}
          user={viewer}
          identityDependencies={identityDependencies}
        >
          <Outlet />
        </AdminShell>
      </div>
    );
  }

  return (
    <div className="app-shell" data-surface="client">
      <a className="skip-link" href="#main">
        {t('common.skip')}
      </a>
      <header className="site-header">
        <div className="shell header-inner">
          <Link to="/" className="wordmark client-wordmark" aria-label={t('common.homeLabel')}>
            <span className="client-wordmark-mark" aria-hidden="true">
              {'</>'}
            </span>
            <span className="client-wordmark-text">CourseForge</span>
          </Link>
          <nav className="client-primary-nav" aria-label={t('common.mainNavigation')}>
            <Link to="/courses">{t('common.catalog')}</Link>
          </nav>
          <div className="client-header-actions">
            <IdentityHeader
              user={session.data ?? null}
              isSigningOut={isSigningOut}
              onSignOut={signOut}
            />
            <LanguageSwitcher />
          </div>
        </div>
      </header>
      {session.data?.emailVerified === false && (
        <div className="shell">
          <EmailVerificationBanner
            isPending={resendVerification.isPending}
            onResend={() => resendVerification.mutate()}
          />
        </div>
      )}
      {signOutFailed && (
        <div className="shell sign-out-message">
          <Toast message={t('identity.signOutFailure')} onDismiss={dismissSignOutFailure} />
        </div>
      )}
      <main id="main" className="shell main-content" tabIndex={-1}>
        <Outlet />
      </main>
      {isStorefront ? (
        <footer className="site-footer site-footer--storefront">
          <div className="shell footer-grid">
            <div className="footer-brand">
              {/* Named by its visible text, not homeLabel: the header wordmark
                  already owns "CourseForge — home", and two links sharing one
                  accessible name make the header link ambiguous. */}
              <Link to="/" className="wordmark">
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
    </div>
  );
}
