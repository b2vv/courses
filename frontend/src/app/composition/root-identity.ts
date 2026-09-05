import type { QueryClient } from '@tanstack/react-query';
import type { AuthenticatedUser } from '../../domain/identity/identity';
import type { IdentityDependencies } from './identity';
import { sessionQueryKey } from './identity-queries';

export type RootIdentity =
  | { readonly status: 'anonymous' }
  | { readonly status: 'signed-in'; readonly user: AuthenticatedUser }
  | { readonly status: 'unavailable' };

export type RootIdentityContext = {
  queryClient: QueryClient;
  identityDependencies: IdentityDependencies;
};

export async function resolveRootIdentity({
  context,
}: {
  context: RootIdentityContext;
}): Promise<{ identity: RootIdentity }> {
  if (!context.identityDependencies.shouldLoadCurrentUser) {
    context.queryClient.setQueryData(sessionQueryKey, null);
    return { identity: { status: 'anonymous' } };
  }

  try {
    const user = await context.queryClient.ensureQueryData(context.identityDependencies.current());
    return {
      identity: user === null ? { status: 'anonymous' } : { status: 'signed-in', user },
    };
  } catch {
    return { identity: { status: 'unavailable' } };
  }
}
