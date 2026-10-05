import { createFileRoute } from '@tanstack/react-router';
import { useMutation } from '@tanstack/react-query';
import { requireIdentity, useRedirectWhenSessionEnds } from '../app/composition/require-identity';
import { SessionQueryError } from '../app/composition/identity-queries';
import { AccountSettingsView } from '../presentation/identity/AccountSettingsView';

export const Route = createFileRoute('/settings/account')({
  ssr: false,
  beforeLoad: ({ context, location }) => requireIdentity(context, location),
  component: AccountSettingsRoute,
});
function AccountSettingsRoute() {
  const { identityDependencies } = Route.useRouteContext();
  useRedirectWhenSessionEnds(identityDependencies);
  const password = useMutation({
    mutationFn: (input: { oldPassword: string; newPassword: string }) =>
      identityDependencies.session.changePassword(input).then(identityDependencies.unwrap),
  });
  const email = useMutation({
    mutationFn: (input: { newEmail: string }) =>
      identityDependencies.session.requestEmailChange(input).then(identityDependencies.unwrap),
  });
  return (
    <AccountSettingsView
      passwordPending={password.isPending}
      emailPending={email.isPending}
      passwordError={password.error instanceof SessionQueryError ? password.error.code : undefined}
      emailError={email.error instanceof SessionQueryError ? email.error.code : undefined}
      passwordSaved={password.isSuccess}
      emailRequested={email.isSuccess}
      onChangePassword={(input) => password.mutate(input)}
      onChangeEmail={(input) => email.mutate(input)}
    />
  );
}
