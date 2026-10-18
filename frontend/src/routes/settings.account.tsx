import { createFileRoute } from '@tanstack/react-router';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { requireIdentity, useRedirectWhenSessionEnds } from '../app/composition/require-identity';
import { SessionQueryError } from '../app/composition/identity-queries';
import { AccountSettingsView } from '../modules/identity';
import type { RecoveryCodes, TwoFactorEnrollment } from '../modules/identity';

export const Route = createFileRoute('/settings/account')({
  ssr: false,
  beforeLoad: ({ context, location }) => requireIdentity(context, location),
  component: AccountSettingsRoute,
});
function AccountSettingsRoute() {
  const { identityDependencies } = Route.useRouteContext();
  useRedirectWhenSessionEnds(identityDependencies);
  const client = useQueryClient();
  const currentUser = useQuery({ ...identityDependencies.current(), refetchOnMount: false });
  const [enrollment, setEnrollment] = useState<TwoFactorEnrollment>();
  const [recoveryCodes, setRecoveryCodes] = useState<RecoveryCodes>();
  const password = useMutation({
    mutationFn: (input: { oldPassword: string; newPassword: string }) =>
      identityDependencies.session.changePassword(input).then(identityDependencies.unwrap),
  });
  const email = useMutation({
    mutationFn: (input: { newEmail: string }) =>
      identityDependencies.session.requestEmailChange(input).then(identityDependencies.unwrap),
  });
  const enroll = useMutation({
    mutationFn: () =>
      identityDependencies.session.enrollTwoFactor().then(identityDependencies.unwrap),
    onSuccess: (result) => setEnrollment(result),
  });
  const confirm = useMutation({
    mutationFn: (code: string) =>
      identityDependencies.session.confirmTwoFactor(code).then(identityDependencies.unwrap),
    onSuccess: async (result) => {
      setEnrollment(undefined);
      setRecoveryCodes(result);
      await identityDependencies.refresh(client);
    },
  });
  const disable = useMutation({
    mutationFn: (input: { password: string; code: string }) =>
      identityDependencies.session
        .disableTwoFactor(input.password, input.code)
        .then(identityDependencies.unwrap),
    onSuccess: async () => {
      setRecoveryCodes(undefined);
      await identityDependencies.refresh(client);
    },
  });
  const regenerate = useMutation({
    mutationFn: (code: string) =>
      identityDependencies.session.regenerateRecoveryCodes(code).then(identityDependencies.unwrap),
    onSuccess: async (result) => {
      setRecoveryCodes(result);
      await identityDependencies.refresh(client);
    },
  });
  const twoFactorPending =
    enroll.isPending || confirm.isPending || disable.isPending || regenerate.isPending;
  const twoFactorError = [enroll.error, confirm.error, disable.error, regenerate.error].find(
    (error): error is SessionQueryError => error instanceof SessionQueryError,
  );
  return (
    <AccountSettingsView
      passwordPending={password.isPending}
      emailPending={email.isPending}
      passwordError={password.error instanceof SessionQueryError ? password.error.code : undefined}
      emailError={email.error instanceof SessionQueryError ? email.error.code : undefined}
      passwordSaved={password.isSuccess}
      emailRequested={email.isSuccess}
      twoFactorEnabled={currentUser.data?.twoFactorEnabled ?? false}
      twoFactorPending={twoFactorPending}
      twoFactorError={twoFactorError?.code}
      enrollment={enrollment}
      recoveryCodes={recoveryCodes}
      onChangePassword={(input) => password.mutate(input)}
      onChangeEmail={(input) => email.mutate(input)}
      onEnrollTwoFactor={() => enroll.mutate()}
      onConfirmTwoFactor={(code) => confirm.mutate(code)}
      onDisableTwoFactor={(input) => disable.mutate(input)}
      onRegenerateRecoveryCodes={(code) => regenerate.mutate(code)}
    />
  );
}
