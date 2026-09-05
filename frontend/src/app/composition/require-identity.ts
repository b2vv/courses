import { useEffect, useRef } from 'react';
import { useQuery } from '@tanstack/react-query';
import { redirect, useLocation, useNavigate } from '@tanstack/react-router';
import type { AuthenticatedUser } from '../../domain/identity/identity';
import { SessionQueryError } from './identity-queries';
import type { IdentityDependencies } from './identity';
import type { RootIdentityContext } from './root-identity';

type GuardLocation = { readonly pathname: string; readonly searchStr: string };

export async function requireIdentity(
  context: RootIdentityContext,
  location: GuardLocation,
): Promise<AuthenticatedUser> {
  let user: AuthenticatedUser | null;
  try {
    user = await context.queryClient.fetchQuery(context.identityDependencies.current());
  } catch {
    throw new SessionQueryError({ code: 'dependency_unavailable' });
  }
  if (user === null) {
    throw redirect({
      to: '/login',
      search: { returnTo: `${location.pathname}${location.searchStr}` },
    });
  }
  return user;
}

export function useRedirectWhenSessionEnds(identityDependencies: IdentityDependencies) {
  const session = useQuery({ ...identityDependencies.current(), refetchOnMount: false });
  const returnTo = useLocation({
    select: (location) => `${location.pathname}${location.searchStr}`,
  });
  const protectedReturnTo = useRef(returnTo);
  const navigate = useNavigate();

  useEffect(() => {
    if (session.data !== null) {
      protectedReturnTo.current = returnTo;
      return;
    }
    void navigate({ to: '/login', search: { returnTo: protectedReturnTo.current } });
  }, [navigate, returnTo, session.data]);
}
