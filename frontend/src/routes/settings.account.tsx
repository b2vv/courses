import { createFileRoute } from '@tanstack/react-router';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import type { MutationKey } from '@tanstack/react-query';
import { requireIdentity, useRedirectWhenSessionEnds } from '../app/composition/require-identity';
import { SessionQueryError, sessionQueryKey } from '../app/composition/identity-queries';
import { AccountSettingsView, isUncertainWriteError } from '../modules/identity';
import type {
  AuthenticatedUser,
  RecoveryCodes,
  SessionErrorCode,
  TwoFactorAction,
  TwoFactorEnrollment,
} from '../modules/identity';
import { pageTitleHead } from '../app/composition/page-title';

export const Route = createFileRoute('/settings/account')({
  ssr: false,
  beforeLoad: ({ context, location }) => requireIdentity(context, location),
  head: ({ match }) => pageTitleHead(match.context.i18n, 'accountSettings'),
  component: AccountSettingsRoute,
});

const passwordKey = ['identity', 'change-password'] as const;
const emailKey = ['identity', 'change-email'] as const;
/** The four 2FA writes share one key: one 2FA action at a time (plan §18 V5a). */
const twoFactorKey = ['identity', 'two-factor'] as const;

function codeOf(error: unknown): SessionErrorCode {
  return error instanceof SessionQueryError ? error.code : 'invalid_response';
}

function AccountSettingsRoute() {
  const { identityDependencies } = Route.useRouteContext();
  useRedirectWhenSessionEnds(identityDependencies);
  const client = useQueryClient();
  const currentUser = useQuery({ ...identityDependencies.current(), refetchOnMount: false });
  const [enrollment, setEnrollment] = useState<TwoFactorEnrollment>();
  const [recovery, setRecovery] = useState<{
    readonly codes: RecoveryCodes;
    readonly kind: 'enabled' | 'regenerated';
  }>();
  const [twoFactorError, setTwoFactorError] = useState<{
    readonly action: TwoFactorAction;
    readonly code: SessionErrorCode;
  }>();
  // Set by an uncertain or conflicting 2FA result; only a successful confirm or a reload clears
  // it, so a later rejection cannot bring back «Скасувати» over a state that may have changed.
  const [stateUnknown, setStateUnknown] = useState(false);

  /** A session that ended mid-write: publish it, and the redirect hook takes the viewer to sign-in. */
  function onWriteError(error: unknown) {
    if (codeOf(error) === 'authentication_required') client.setQueryData(sessionQueryKey, null);
  }
  function setTwoFactorFlag(enabled: boolean) {
    client.setQueryData<AuthenticatedUser | null>(sessionQueryKey, (user) =>
      user == null ? user : { ...user, twoFactorEnabled: enabled },
    );
    void identityDependencies.refresh(client);
  }
  function twoFactorFailed(action: TwoFactorAction, error: unknown) {
    onWriteError(error);
    const code = codeOf(error);
    setTwoFactorError({ action, code });
    const conflict = code === 'twofactor_already_enabled' || code === 'twofactor_not_enabled';
    const uncertain = (action === 'confirm' || action === 'disable') && isUncertainWriteError(code);
    if (conflict || uncertain) setStateUnknown(true);
  }

  const password = useMutation({
    mutationKey: passwordKey,
    mutationFn: (input: { oldPassword: string; newPassword: string }) =>
      identityDependencies.session.changePassword(input).then(identityDependencies.unwrap),
    onError: onWriteError,
  });
  const email = useMutation({
    mutationKey: emailKey,
    mutationFn: (input: { newEmail: string }) =>
      identityDependencies.session.requestEmailChange(input).then(identityDependencies.unwrap),
    onError: onWriteError,
  });
  const enroll = useMutation({
    mutationKey: twoFactorKey,
    mutationFn: () =>
      identityDependencies.session.enrollTwoFactor().then(identityDependencies.unwrap),
    onMutate: () => setTwoFactorError(undefined),
    onSuccess: (result) => setEnrollment(result),
    onError: (error) => twoFactorFailed('enroll', error),
  });
  const confirm = useMutation({
    mutationKey: twoFactorKey,
    mutationFn: (code: string) =>
      identityDependencies.session.confirmTwoFactor(code).then(identityDependencies.unwrap),
    onMutate: () => setTwoFactorError(undefined),
    onSuccess: (result) => {
      setEnrollment(undefined);
      setStateUnknown(false);
      setRecovery({ codes: result, kind: 'enabled' });
      setTwoFactorFlag(true);
    },
    onError: (error) => twoFactorFailed('confirm', error),
  });
  const disable = useMutation({
    mutationKey: twoFactorKey,
    mutationFn: (input: { password: string; code: string }) =>
      identityDependencies.session
        .disableTwoFactor(input.password, input.code)
        .then(identityDependencies.unwrap),
    onMutate: () => setTwoFactorError(undefined),
    onSuccess: () => {
      setRecovery(undefined);
      setTwoFactorFlag(false);
    },
    onError: (error) => twoFactorFailed('disable', error),
  });
  const regenerate = useMutation({
    mutationKey: twoFactorKey,
    mutationFn: (code: string) =>
      identityDependencies.session.regenerateRecoveryCodes(code).then(identityDependencies.unwrap),
    onMutate: () => setTwoFactorError(undefined),
    // Ends with its own answer: codes are not part of the session (plan §18 V5a).
    onSuccess: (result) => setRecovery({ codes: result, kind: 'regenerated' }),
    onError: (error) => twoFactorFailed('regenerate', error),
  });

  /** Refuse a second press at dispatch: the mutation is pending in the cache before any render. */
  function idle(key: MutationKey) {
    return client.isMutating({ mutationKey: key, exact: true }) === 0;
  }
  const twoFactorPending =
    enroll.isPending || confirm.isPending || disable.isPending || regenerate.isPending;

  return (
    <AccountSettingsView
      currentEmail={currentUser.data?.email}
      passwordPending={password.isPending}
      emailPending={email.isPending}
      passwordError={password.error === null ? undefined : codeOf(password.error)}
      emailError={email.error === null ? undefined : codeOf(email.error)}
      passwordSavedAt={password.isSuccess ? password.submittedAt : undefined}
      emailRequestedAt={email.isSuccess ? email.submittedAt : undefined}
      twoFactorEnabled={currentUser.data?.twoFactorEnabled ?? false}
      twoFactorPending={twoFactorPending}
      twoFactorError={twoFactorError}
      twoFactorStateUnknown={stateUnknown}
      enrollment={enrollment}
      recovery={recovery}
      onChangePassword={(input) => {
        if (idle(passwordKey)) password.mutate(input);
      }}
      onChangeEmail={(input) => {
        if (idle(emailKey)) email.mutate(input);
      }}
      onEnrollTwoFactor={() => {
        if (idle(twoFactorKey)) enroll.mutate();
      }}
      onConfirmTwoFactor={(code) => {
        if (idle(twoFactorKey)) confirm.mutate(code);
      }}
      onCancelTwoFactor={() => {
        if (!idle(twoFactorKey) || stateUnknown) return;
        setEnrollment(undefined);
        setTwoFactorError(undefined);
      }}
      onDisableTwoFactor={(input) => {
        if (idle(twoFactorKey)) disable.mutate(input);
      }}
      onRegenerateRecoveryCodes={(code) => {
        if (idle(twoFactorKey)) regenerate.mutate(code);
      }}
      onDismissRecoveryCodes={() => setRecovery(undefined)}
    />
  );
}
