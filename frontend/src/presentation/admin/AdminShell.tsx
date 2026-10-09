import type { ReactNode } from 'react';
import { Link } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import type { AuthenticatedUser } from '../../domain/identity/identity';
import type { IdentityDependencies } from '../../app/composition/identity';
import { useSignOut } from '../../app/composition/use-sign-out';
import { LanguageSwitcher } from '../LanguageSwitcher';
import { Button, Toast } from '../ui';
import { AdminNavIcon, type AdminNavIconKey } from './AdminNavIcon';

export type { AdminNavIconKey } from './AdminNavIcon';

export type AdminNavItem = {
  readonly to: string;
  readonly label: string;
  readonly exact?: boolean;
  readonly icon: AdminNavIconKey;
};

export type AdminNavSection = {
  /** UPPERCASE group label; omit for the primary (unlabelled) section. */
  readonly label?: string;
  readonly items: readonly AdminNavItem[];
};

type AdminShellProps = {
  readonly title: string;
  readonly navLabel: string;
  readonly navSections: readonly AdminNavSection[];
  readonly user: AuthenticatedUser;
  readonly identityDependencies: IdentityDependencies;
  readonly children: ReactNode;
};

/**
 * The admin-side shell: a narrow icon rail, a wide sectioned sidebar and a top
 * bar that replace the client chrome for /admin and /teach. It carries its own
 * navigation, a link back to the client, and sign-out — the two surfaces are
 * separate apps that share one design system (the NorthWind theme comes from the
 * data-surface="admin" scope set on the root .app-shell).
 */
export function AdminShell({
  title,
  navLabel,
  navSections,
  user,
  identityDependencies,
  children,
}: AdminShellProps) {
  const { t } = useTranslation();
  const { signOut, isSigningOut, signOutFailed, dismissSignOutFailure } =
    useSignOut(identityDependencies);
  const railItems = navSections.flatMap((section) => section.items);

  return (
    <div className="admin-shell">
      {/* The rail mirrors the sidebar as a visual shortcut. The sidebar <nav> is the
          canonical navigation; the rail is hidden from assistive tech and the tab
          order (aria-hidden + tabIndex -1) so there are no duplicate landmarks or
          duplicate-named links. Mouse users can still click it. */}
      <div className="admin-rail" aria-hidden="true">
        <Link to="/admin" className="admin-rail-brand" tabIndex={-1}>
          <span aria-hidden="true">{'</>'}</span>
        </Link>
        <div className="admin-rail-items">
          {railItems.map((item) => (
            <Link
              key={item.to}
              to={item.to}
              activeOptions={{ exact: item.exact ?? false }}
              className="admin-rail-item"
              activeProps={{ 'aria-current': 'page' }}
              tabIndex={-1}
              title={item.label}
            >
              <AdminNavIcon name={item.icon} />
            </Link>
          ))}
        </div>
      </div>
      <aside className="admin-sidebar">
        <div className="admin-workspace">
          <span className="admin-workspace-name">CourseForge</span>
          <span className="admin-workspace-area">{navLabel}</span>
        </div>
        <nav className="admin-nav" aria-label={navLabel}>
          {navSections.map((section, index) => (
            <div className="admin-nav-section" key={section.label ?? `primary-${index}`}>
              {section.label != null && (
                <span className="admin-nav-section-label">{section.label}</span>
              )}
              {section.items.map((item) => (
                <Link
                  key={item.to}
                  to={item.to}
                  activeOptions={{ exact: item.exact ?? false }}
                  className="admin-nav-link"
                  activeProps={{ 'aria-current': 'page' }}
                >
                  <AdminNavIcon name={item.icon} />
                  <span>{item.label}</span>
                </Link>
              ))}
            </div>
          ))}
        </nav>
        <Link to="/" className="admin-to-client">
          <span aria-hidden="true">←</span> {t('common.toClient')}
        </Link>
      </aside>
      <div className="admin-canvas">
        <header className="admin-topbar">
          <span className="admin-topbar-title">{title}</span>
          <div className="admin-topbar-actions">
            <span className="admin-topbar-user">{user.displayName}</span>
            <LanguageSwitcher />
            <Button variant="secondary" isDisabled={isSigningOut} onPress={signOut}>
              {isSigningOut ? t('identity.signingOut') : t('identity.signOut')}
            </Button>
          </div>
        </header>
        {signOutFailed && (
          <div className="admin-topbar-message">
            <Toast message={t('identity.signOutFailure')} onDismiss={dismissSignOutFailure} />
          </div>
        )}
        <main id="main" className="admin-content" tabIndex={-1}>
          {children}
        </main>
      </div>
    </div>
  );
}
