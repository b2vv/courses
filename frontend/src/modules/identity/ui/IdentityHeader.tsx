import { useTranslation } from 'react-i18next';
import type { AuthenticatedUser } from '../domain/identity';
import { Button, Link } from '../../../presentation/ui';

type IdentityHeaderProps = {
  readonly user: AuthenticatedUser | null;
  readonly isSigningOut: boolean;
  readonly onSignOut: () => void;
};

export function IdentityHeader({ user, isSigningOut, onSignOut }: IdentityHeaderProps) {
  const { t } = useTranslation();

  if (user === null)
    return (
      <div className="header-identity">
        <Link href="/login" className="header-login">
          {t('common.login')}
        </Link>
      </div>
    );

  return (
    <div className="header-identity">
      <span className="header-identity-name">{user.displayName}</span>
      <Link href="/settings/account" className="header-cross-nav">
        {t('common.footerAccount')}
      </Link>
      {user.role === 'admin' && (
        <Link href="/admin" className="header-cross-nav">
          {t('common.toAdmin')}
        </Link>
      )}
      {user.role === 'instructor' && (
        <Link href="/teach" className="header-cross-nav">
          {t('common.toTeach')}
        </Link>
      )}
      <Button variant="secondary" isDisabled={isSigningOut} onPress={onSignOut}>
        {isSigningOut ? t('identity.signingOut') : t('identity.signOut')}
      </Button>
    </div>
  );
}
