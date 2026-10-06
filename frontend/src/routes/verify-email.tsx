import { useEffect } from 'react';
import { createFileRoute } from '@tanstack/react-router';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { SessionQueryError } from '../app/composition/identity-queries';
import { parseTokenSearch } from '../app/router/token-search';
import { VerifyEmailView } from '../presentation/identity/VerifyEmailView';
export const Route = createFileRoute('/verify-email')({
  validateSearch: parseTokenSearch,
  component: VerifyEmailRoute,
});
function VerifyEmailRoute() {
  const { token } = Route.useSearch();
  const { identityDependencies } = Route.useRouteContext();
  const client = useQueryClient();
  const verify = useMutation({
    mutationFn: (value: string) =>
      identityDependencies.session.verify({ token: value }).then(identityDependencies.unwrap),
    onSuccess: () => identityDependencies.refresh(client),
  });
  const resend = useMutation({
    mutationFn: () => identityDependencies.session.resend({}).then(identityDependencies.unwrap),
  });
  useEffect(() => {
    if (token !== undefined && verify.status === 'idle') verify.mutate(token);
  }, [token, verify]);
  const error = verify.error ?? resend.error;
  return (
    <VerifyEmailView
      isPending={verify.isPending || resend.isPending}
      completed={verify.isSuccess}
      errorCode={
        error instanceof SessionQueryError
          ? error.code
          : token === undefined
            ? 'invalid_account_token'
            : undefined
      }
      onResend={() => resend.mutate()}
    />
  );
}
