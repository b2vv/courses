import { useEffect } from 'react';
import { createFileRoute } from '@tanstack/react-router';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { SessionQueryError } from '../app/composition/identity-queries';
import { parseTokenSearch } from '../app/router/token-search';
import { ConfirmEmailChangeView } from '../modules/identity';

export const Route = createFileRoute('/email/change/confirm')({
  validateSearch: parseTokenSearch,
  component: ConfirmEmailChangeRoute,
});
function ConfirmEmailChangeRoute() {
  const { token } = Route.useSearch();
  const { identityDependencies } = Route.useRouteContext();
  const client = useQueryClient();
  const confirm = useMutation({
    mutationFn: (value: string) =>
      identityDependencies.session
        .confirmEmailChange({ token: value })
        .then(identityDependencies.unwrap),
    onSuccess: () => identityDependencies.refresh(client),
  });
  useEffect(() => {
    if (token !== undefined && confirm.status === 'idle') confirm.mutate(token);
  }, [token, confirm]);
  return (
    <ConfirmEmailChangeView
      isPending={confirm.isPending}
      completed={confirm.isSuccess}
      errorCode={
        confirm.error instanceof SessionQueryError
          ? confirm.error.code
          : token === undefined
            ? 'invalid_account_token'
            : undefined
      }
    />
  );
}
