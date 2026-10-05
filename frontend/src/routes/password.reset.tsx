import { createFileRoute, useNavigate } from '@tanstack/react-router';
import { useMutation } from '@tanstack/react-query';
import { SessionQueryError } from '../app/composition/identity-queries';
import { parseTokenSearch } from '../app/router/token-search';
import { PasswordResetView } from '../presentation/identity/PasswordResetView';
export const Route = createFileRoute('/password/reset')({
  validateSearch: parseTokenSearch,
  component: PasswordResetRoute,
});
function PasswordResetRoute() {
  const { token } = Route.useSearch();
  const { identityDependencies } = Route.useRouteContext();
  const navigate = useNavigate();
  const reset = useMutation({
    mutationFn: (password: string) =>
      token === undefined
        ? Promise.reject(new SessionQueryError({ code: 'invalid_account_token' }))
        : identityDependencies.session.reset({ token, password }).then(identityDependencies.unwrap),
    onSuccess: () => void navigate({ to: '/login', search: { notice: 'password-reset' } }),
  });
  return (
    <PasswordResetView
      isPending={reset.isPending}
      errorCode={
        reset.error instanceof SessionQueryError
          ? reset.error.code
          : token === undefined
            ? 'invalid_account_token'
            : undefined
      }
      onSubmit={(password) => reset.mutate(password)}
    />
  );
}
