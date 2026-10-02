import type { ReactNode } from 'react';
import { Link } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import type { AuthenticatedUser } from '../../domain/identity/identity';
import type { IdentityDependencies } from '../../app/composition/identity';
import { useSignOut } from '../../app/composition/use-sign-out';
import { LanguageSwitcher } from '../LanguageSwitcher';
import { Button, Toast } from '../ui';

export type AdminNavItem = {
  readonly to: string;
  readonly label: string;
  readonly exact?: boolean;
};

type AdminShellProps = {
  readonly title: string;
  readonly navLabel: string;
  readonly navItems: readonly AdminNavItem[];
  readonly user: AuthenticatedUser;
  readonly identityDependencies: IdentityDependencies;
  readonly children: ReactNode;
};

/**
 * The admin-side shell: a dedicated left sidebar and top bar that replace the
 * client chrome for /admin and /teach. It carries its own navigation, a link
 * back to the client, and sign-out — the two surfaces are separate apps that
 * share one design system (the NorthWind editorial theme comes from the
 * data-surface="admin" scope set on the root .app-shell).
 */
export function AdminShell({
  title,
  navLabel,
  navItems,
  user,
  identityDependencies,
  children,
}: AdminShellProps) {
  const { t } = useTranslation();
  const { signOut, isSigningOut, signOutFailed, dismissSignOutFailure } =
    useSignOut(identityDependencies);

  return (
    <div className="admin-shell">
      <aside className="admin-sidebar">
        <Link to="/admin" className="admin-brand" aria-label={t('common.homeLabel')}>
          <span aria-hidden="true">{'</>'}</span>
          <span className="admin-brand-name">CourseForge</span>
        </Link>
        <nav className="admin-nav" aria-label={navLabel}>
          {navItems.map((item) => (
            <Link
              key={item.to}
              to={item.to}
              activeOptions={{ exact: item.exact ?? false }}
              className="admin-nav-link"
              activeProps={{ 'aria-current': 'page' }}
            >
              {item.label}
            </Link>
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
