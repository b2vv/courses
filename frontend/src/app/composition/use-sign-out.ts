import { useState } from 'react';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from '@tanstack/react-router';
import type { IdentityDependencies } from './identity';
import { sessionQueryKey } from './identity-queries';

/**
 * Sign-out shared by the client header and the admin shell. A settled sign-out
 * always lands the viewer on the public catalogue and clears the cached session,
 * whether or not the request itself failed (a failed log-out still ends the
 * local session view); `signOutFailed` lets the caller surface the failure.
 */
export function useSignOut(identityDependencies: IdentityDependencies) {
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const [signOutFailed, setSignOutFailed] = useState(false);
  const mutation = useMutation({
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
  return {
    signOut: () => mutation.mutate(),
    isSigningOut: mutation.isPending,
    signOutFailed,
    dismissSignOutFailure: () => setSignOutFailed(false),
  };
}
