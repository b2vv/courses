import { useEffect, useRef } from 'react';
import { createFileRoute } from '@tanstack/react-router';
import { useMutation } from '@tanstack/react-query';
import { SessionQueryError } from '../app/composition/identity-queries';
import { useResendVerification } from '../app/composition/resend-verification';
import { useOnline, useSessionReadAfter } from '../app/composition/session-after';
import { parseTokenSearch } from '../app/router/token-search';
import { VerifyEmailView } from '../modules/identity';
import type { SessionErrorCode, VerifyEmailState } from '../modules/identity';
import { pageTitleHead } from '../app/composition/page-title';

export const Route = createFileRoute('/verify-email')({
  validateSearch: parseTokenSearch,
  head: ({ match }) => pageTitleHead(match.context.i18n, 'verifyEmail'),
  component: VerifyEmailRoute,
});

/** A missing, unknown or used link: the server will not take it (plan §18 S3/S4). */
const linkFailedCodes: ReadonlySet<SessionErrorCode> = new Set([
  'invalid_account_token',
  'token_used',
  'invalid_request',
]);

function VerifyEmailRoute() {
  const { token } = Route.useSearch();
  // A new token is a new screen: its own attempt and its own session read (plan §18 V4a).
  return <VerifyEmailScreen key={token ?? ''} token={token} />;
}

function VerifyEmailScreen({ token }: { readonly token: string | undefined }) {
  const { identityDependencies } = Route.useRouteContext();
  const verify = useMutation({
    mutationFn: (value: string) =>
      identityDependencies.session.verify({ token: value }).then(identityDependencies.unwrap),
  });
  // One automatic attempt per screen, also under StrictMode's repeated effect.
  const started = useRef(false);
  useEffect(() => {
    if (token === undefined || started.current) return;
    started.current = true;
    verify.mutate(token);
  }, [token, verify]);
  const resend = useResendVerification(identityDependencies);
  const online = useOnline();

  const errorCode: SessionErrorCode | undefined =
    verify.error === null
      ? undefined
      : verify.error instanceof SessionQueryError
        ? verify.error.code
        : 'invalid_response';
  const result =
    token === undefined || (errorCode !== undefined && linkFailedCodes.has(errorCode))
      ? 'link-failed'
      : verify.isSuccess
        ? 'verified'
        : errorCode !== undefined
          ? 'failed'
          : 'in-flight';
  // The read is tied to this result: a retry that ends the same way reads again.
  const read = useSessionReadAfter(
    identityDependencies,
    result === 'verified' || result === 'link-failed'
      ? `${result}:${verify.submittedAt}`
      : undefined,
  );
  const offline = read?.status === 'pending' && !online;

  let state: VerifyEmailState;
  if (result === 'in-flight') state = { kind: verify.isPaused ? 'waiting' : 'checking' };
  else if (result === 'failed') state = { kind: 'failed', code: errorCode ?? 'invalid_response' };
  else if (result === 'verified')
    state = {
      kind: 'verified',
      next:
        read === undefined || read.status === 'pending'
          ? 'pending'
          : read.status === 'done' && read.user !== null
            ? 'learn'
            : 'login',
      offline,
    };
  else if (read?.status === 'done' && read.user?.emailVerified === true)
    state = { kind: 'already-verified' };
  else
    state = {
      kind: 'link-failed',
      action:
        read === undefined || read.status === 'pending'
          ? 'pending'
          : read.status === 'failed'
            ? 'reload'
            : read.user === null
              ? 'guest'
              : 'signed-in',
      offline,
    };

  return (
    <VerifyEmailView
      state={state}
      resendPending={resend.isPending}
      resendOutcome={resend.outcome}
      retryPending={verify.isPending}
      onResend={(email) => resend.resend(email)}
      onRetry={() => {
        if (token !== undefined && !verify.isPending) verify.mutate(token);
      }}
    />
  );
}
