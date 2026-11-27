import { useIsMutating, useMutation, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { resendOutcomeOf } from '../../modules/identity';
import type { ResendOutcome } from '../../modules/identity';
import { SessionQueryError } from './identity-queries';
import type { IdentityDependencies } from './identity';

/** All resend buttons share this key, so one press anywhere holds every one of them. */
export const resendVerificationKey = ['identity', 'resend-verification'] as const;

/**
 * Resend the verification email (design-pass ACCT V1). The verify page, the banner and the
 * lesson gate can be on screen together: a press is refused at dispatch while any resend is in
 * flight (the mutation is pending in the cache synchronously, before React renders), and each
 * surface shows only the result of its own press.
 */
export function useResendVerification(
  identityDependencies: IdentityDependencies,
  /**
   * The signed-in viewer, where the surface outlives sign-out and sign-in (the root banner).
   * A result is shown only to the viewer who pressed, so neither it nor a late answer reaches
   * the next account.
   */
  viewerId?: string,
) {
  const client = useQueryClient();
  const [pressedBy, setPressedBy] = useState<string>();
  const mutation = useMutation({
    mutationKey: resendVerificationKey,
    mutationFn: (email: string | undefined) =>
      identityDependencies.session
        .resend(email === undefined ? {} : { email })
        .then(identityDependencies.unwrap),
  });
  const inFlight = useIsMutating({ mutationKey: resendVerificationKey, exact: true }) > 0;
  const resend = (email?: string) => {
    if (client.isMutating({ mutationKey: resendVerificationKey, exact: true }) > 0) return;
    setPressedBy(viewerId);
    mutation.mutate(email);
  };
  const outcome: ResendOutcome | undefined =
    pressedBy !== viewerId
      ? undefined
      : mutation.isSuccess
        ? 'accepted'
        : mutation.error === null
          ? undefined
          : resendOutcomeOf(
              mutation.error instanceof SessionQueryError
                ? mutation.error.code
                : 'invalid_response',
            );
  return { resend, isPending: inFlight, outcome };
}
