import {
  createRootRouteWithContext,
  HeadContent,
  Outlet,
  Scripts,
  Link,
  useRouter,
  useRouterState,
} from '@tanstack/react-router';
import { useMutation, useQuery } from '@tanstack/react-query';
import type { createCatalogDependencies } from '../app/composition/catalog';
import type { AdminDependencies } from '../app/composition/admin';
import type { LearnDependencies } from '../app/composition/learn';
import type { TeachDependencies } from '../app/composition/teach';
import type { createEnrollmentDependencies } from '../app/composition/enrollment';
import { useEffect, useState, type ReactNode } from 'react';
import { useTranslation } from 'react-i18next';
import { forgetOtherAuthors } from '../modules/transcript';
import { useSignOut } from '../app/composition/use-sign-out';
import { canAccessAdministration, canAdministerAccounts } from '../modules/identity';
import { AdminShell, type AdminNavSection } from '../modules/admin';
import type { TranscriptHistoryStore } from '../modules/transcript';
import { LanguageSwitcher } from '../presentation/LanguageSwitcher';
import { IdentityHeader } from '../modules/identity';
import { EmailVerificationBanner } from '../modules/identity';
import { MenuCard, StatusPage, Toast } from '../presentation/ui';
import { resolveRootIdentity, type RootIdentityContext } from '../app/composition/root-identity';
import type { createI18n } from '../app/providers/i18n';
import { pageTitle } from '../app/composition/page-title';
import '../app/styles/main.css';

type RootRouteContext = RootIdentityContext & {
  i18n: ReturnType<typeof createI18n>;
  admin: AdminDependencies;
  catalog: ReturnType<typeof createCatalogDependencies>;
  learn: LearnDependencies;
  teach: TeachDependencies;
  enrollment: ReturnType<typeof createEnrollmentDependencies>;
  transcriptHistory: TranscriptHistoryStore;
};

export const Route = createRootRouteWithContext<RootRouteContext>()({
  beforeLoad: resolveRootIdentity,
  head: ({ match, matches }) => {
    const { i18n } = match.context;
    // A page that resolved to "not found" has no title of its own, so the root names it: either
    // no route below the root matched the path, or a matched route threw `notFound()`.
    // `_notFound` marks the boundary that renders it (typed on RouteMatch in router-core 1.171);
    // the root is that boundary for the `/admin/$` catch-all.
    const notFound =
      matches.length === 1 ||
      matches.some((candidate) => candidate.status === 'notFound' || candidate._notFound === true);
    return {
      meta: [
        { charSet: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        { title: notFound ? pageTitle(i18n, 'notFound') : i18n.t('pageTitle.appTitle') },
        { name: 'description', content: i18n.t('pageTitle.appDescription') },
      ],
    };
  },
  shellComponent: Document,
  component: Layout,
  notFoundComponent: NotFound,
  errorComponent: PageError,
});

function NotFound() {
  const { t } = useTranslation();
  const { identityDependencies } = Route.useRouteContext();
  const session = useQuery({ ...identityDependencies.current(), refetchOnMount: false });
  const onAdminPath = useRouterState({
    select: (state) =>
      state.location.pathname === '/admin' || state.location.pathname.startsWith('/admin/'),
  });
  // Plain anchors: a not-found answer can sit under a shell that is not the destination's (a
  // refused viewer on an admin path), so the way back is a full page load.
  return (
    <StatusPage label="404" title={t('page.notFound')} hint={t('page.notFoundHint')}>
      {onAdminPath ? (
        <a className="ui-button ui-button--primary" href="/admin">
          {t('page.goToAdmin')}
        </a>
      ) : session.data ? (
        <>
          <a className="ui-button ui-button--primary" href="/learn">
            {t('common.myLearning')}
          </a>
          <a className="ui-button ui-button--secondary" href="/courses">
            {t('common.catalog')}
          </a>
        </>
      ) : (
        <>
          <a className="ui-button ui-button--primary" href="/courses">
            {t('catalog.goToCatalog')}
          </a>
          <a className="ui-button ui-button--secondary" href="/">
            {t('page.goHome')}
          </a>
        </>
      )}
    </StatusPage>
  );
}
function PageError() {
  const { t } = useTranslation();
  // A form, so the retry works before hydration and after a failed client bundle.
  return (
    <StatusPage title={t('page.failure')} hint={t('page.failureHint')} isAlert>
      <form method="get">
        <button type="submit" className="ui-button ui-button--primary">
          {t('common.retry')}
        </button>
      </form>
      <a className="ui-button ui-button--secondary" href="/">
        {t('page.goHome')}
      </a>
    </StatusPage>
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
  // The compact header's menu closes on every navigation: a link followed inside it, or
  // Back/Forward while it is open (a remembered address would reopen it on return).
  const router = useRouter();
  const [isMenuOpen, setMenuOpen] = useState(false);
  useEffect(() => router.subscribe('onBeforeNavigate', () => setMenuOpen(false)), [router]);
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
          {/* Two layouts of the same header; CSS shows exactly one by the header's width in em,
              so enlarged text switches to the compact one too (G0 #1). The hidden one is
              display:none and so out of the accessibility tree. */}
          <nav className="client-primary-nav header-wide" aria-label={t('common.mainNavigation')}>
            <Link to="/courses">{t('common.catalog')}</Link>
            {session.data ? <Link to="/learn">{t('common.myLearning')}</Link> : null}
          </nav>
          <div className="client-header-actions header-wide">
            <IdentityHeader
              user={session.data ?? null}
              isSigningOut={isSigningOut}
              onSignOut={signOut}
            />
            <LanguageSwitcher />
          </div>
          <div className="header-compact">
            {/* A guest's sign-in stays one tap away beside «Меню» where the row has room; on a
                narrower header or with enlarged text CSS hides it, and the copy inside the menu
                is the way in (owner 2026-11-22, variant «Б, де вміщається»). */}
            {session.data ? null : (
              <Link to="/login" className="ui-link header-login header-compact-login">
                {t('common.login')}
              </Link>
            )}
            <MenuCard
              triggerLabel={t('common.menu')}
              title={t('common.menu')}
              closeLabel={t('common.closeMenu')}
              isOpen={isMenuOpen}
              onOpenChange={setMenuOpen}
            >
              {/* A press closes the menu even on the current page's own link, where the
                  address, and so the open state, would not change. */}
              <nav className="client-menu-nav" aria-label={t('common.mainNavigation')}>
                <Link to="/courses" onClick={() => setMenuOpen(false)}>
                  {t('common.catalog')}
                </Link>
                {session.data ? (
                  <Link to="/learn" onClick={() => setMenuOpen(false)}>
                    {t('common.myLearning')}
                  </Link>
                ) : null}
              </nav>
              <IdentityHeader
                user={session.data ?? null}
                isSigningOut={isSigningOut}
                onSignOut={signOut}
              />
              <LanguageSwitcher />
            </MenuCard>
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
      {/* One footer on every client page (G0 #5); signed-in links only for a signed-in viewer (#8). */}
      <footer className="site-footer">
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
            {session.data ? <Link to="/learn">{t('common.myLearning')}</Link> : null}
          </nav>
        </div>
      </footer>
    </div>
  );
}
