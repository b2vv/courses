import { createFileRoute, useNavigate } from '@tanstack/react-router';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { SessionQueryError, sessionQueryKey } from '../app/composition/identity-queries';
import { RegisterView } from '../modules/identity';
import type { AuthenticatedUser, RegisterInput } from '../modules/identity';
export const Route = createFileRoute('/register')({ component: RegisterRoute });
function RegisterRoute() {
  const { identityDependencies } = Route.useRouteContext();
  const client = useQueryClient();
  const navigate = useNavigate();
  const register = useMutation({
    mutationFn: (input: RegisterInput) =>
      identityDependencies.session.register(input).then(identityDependencies.unwrap),
    onSuccess: (user: AuthenticatedUser) => {
      client.setQueryData(sessionQueryKey, user);
      void navigate({ to: '/', search: { notice: 'verify-email' } });
    },
  });
  return (
    <RegisterView
      isPending={register.isPending}
      errorCode={register.error instanceof SessionQueryError ? register.error.code : undefined}
      onSubmit={(input) => register.mutate(input)}
    />
  );
}
