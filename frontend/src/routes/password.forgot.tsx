import { createFileRoute } from '@tanstack/react-router';
import { useMutation } from '@tanstack/react-query';
import { SessionQueryError } from '../app/composition/identity-queries';
import { PasswordForgotView } from '../presentation/identity/PasswordForgotView';
export const Route = createFileRoute('/password/forgot')({ component: PasswordForgotRoute });
function PasswordForgotRoute() {
  const { identityDependencies } = Route.useRouteContext();
  const forgot = useMutation({
    mutationFn: (email: string) =>
      identityDependencies.session.forgot({ email }).then(identityDependencies.unwrap),
  });
  return (
    <PasswordForgotView
      isPending={forgot.isPending}
      sent={forgot.isSuccess}
      errorCode={forgot.error instanceof SessionQueryError ? forgot.error.code : undefined}
      onSubmit={(email) => forgot.mutate(email)}
    />
  );
}
