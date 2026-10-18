import { useEffect, useState, type ReactNode } from 'react';
import * as Aria from 'react-aria-components';
import { Link } from '@tanstack/react-router';
import { useTranslation } from 'react-i18next';
import type { AuthenticatedUser } from '../../identity';
import type { IdentityDependencies } from '../../../app/composition/identity';
import { useSignOut } from '../../../app/composition/use-sign-out';
import { LanguageSwitcher } from '../../../presentation/LanguageSwitcher';
import { Button, Toast, Tooltip } from '../../../presentation/ui';
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
 * The admin-side shell: one sectioned sidebar (a modal drawer on narrow screens)
 * and a top bar that replace the client chrome for /admin and /teach. It carries its own
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
  const [isDrawerOpen, setDrawerOpen] = useState(false);

  useEffect(() => {
    const narrow = window.matchMedia('(max-width: 52rem)');
    const closeOnDesktop = () => {
      if (!narrow.matches) setDrawerOpen(false);
    };
    narrow.addEventListener('change', closeOnDesktop);
    return () => narrow.removeEventListener('change', closeOnDesktop);
  }, []);

  const navigation = (
    <>
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
                onClick={() => setDrawerOpen(false)}
                activeProps={{ 'aria-current': 'page' }}
              >
                <AdminNavIcon name={item.icon} />
                <span>{item.label}</span>
              </Link>
            ))}
          </div>
        ))}
      </nav>
      <Link to="/" className="admin-to-client" onClick={() => setDrawerOpen(false)}>
        <span aria-hidden="true">←</span> {t('common.toClient')}
      </Link>
    </>
  );

  return (
    <div className="admin-shell">
      {!isDrawerOpen && <aside className="admin-sidebar">{navigation}</aside>}
      <div className="admin-canvas">
        <header className="admin-topbar">
          <Aria.DialogTrigger isOpen={isDrawerOpen} onOpenChange={setDrawerOpen}>
            <Tooltip content={t('common.mainNavigation')}>
              <Aria.Button
                className="ui-button ui-button--secondary admin-nav-toggle"
                aria-label={t('common.mainNavigation')}
                aria-expanded={isDrawerOpen}
              >
                <AdminNavIcon name="menu" />
              </Aria.Button>
            </Tooltip>
            <Aria.ModalOverlay className="admin-drawer-overlay" isDismissable>
              <Aria.Modal className="admin-drawer">
                <Aria.Dialog
                  className="admin-sidebar admin-drawer-dialog"
                  aria-label={t('common.mainNavigation')}
                >
                  {({ close }) => (
                    <>
                      <div className="admin-drawer-close">
                        <Button variant="secondary" onPress={close}>
                          {t('common.close')}
                        </Button>
                      </div>
                      {navigation}
                    </>
                  )}
                </Aria.Dialog>
              </Aria.Modal>
            </Aria.ModalOverlay>
          </Aria.DialogTrigger>
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
