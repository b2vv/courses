import { useTranslation } from 'react-i18next';
import { canAccessAdministration, type AuthenticatedUser } from '../../domain/identity/identity';
import { Button, Link } from '../ui';

type IdentityHeaderProps = {
  readonly user: AuthenticatedUser | null;
  readonly isSigningOut: boolean;
  readonly onSignOut: () => void;
};

export function IdentityHeader({ user, isSigningOut, onSignOut }: IdentityHeaderProps) {
  const { t } = useTranslation();

  if (user === null) return <Link href="/login">{t('common.login')}</Link>;

  return (
    <>
      <span className="header-identity-name">{user.displayName}</span>
      {canAccessAdministration(user) && <Link href="/admin">{t('administration.title')}</Link>}
      <Button variant="secondary" isDisabled={isSigningOut} onPress={onSignOut}>
        {isSigningOut ? t('identity.signingOut') : t('identity.signOut')}
      </Button>
    </>
  );
}
