import { useCallback, useEffect, useRef } from 'react';
import { createFileRoute } from '@tanstack/react-router';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { SessionQueryError } from '../app/composition/identity-queries';
import { parseTokenSearch } from '../app/router/token-search';
import { ConfirmEmailChangeView } from '../modules/identity';
import type { ConfirmEmailChangeState, SessionErrorCode } from '../modules/identity';
import { pageTitleHead } from '../app/composition/page-title';

export const Route = createFileRoute('/email/change/confirm')({
  validateSearch: parseTokenSearch,
  head: ({ match }) => pageTitleHead(match.context.i18n, 'emailChangeConfirm'),
  component: ConfirmEmailChangeRoute,
});

const linkFailedCodes: ReadonlySet<SessionErrorCode> = new Set([
  'invalid_account_token',
  'token_used',
  'invalid_request',
]);

function ConfirmEmailChangeRoute() {
  const { token } = Route.useSearch();
  return <ConfirmEmailChangeScreen key={token ?? ''} token={token} />;
}

function ConfirmEmailChangeScreen({ token }: { readonly token: string | undefined }) {
  const { identityDependencies } = Route.useRouteContext();
  const client = useQueryClient();
  const confirm = useMutation({
    mutationFn: (value: string) =>
      identityDependencies.session
        .confirmEmailChange({ token: value })
        .then(identityDependencies.unwrap),
    // Started, not returned: a returned promise would hold the mutation pending until `/me`
    // answers, and E2 with its action must show right after the 204 (plan §18 V3).
    onSuccess: () => {
      void identityDependencies.refresh(client);
    },
  });
  // This screen's own attempt in flight refuses another press at dispatch, before a render
  // (ACCT-7). Per screen: a request left by an earlier visit to the same token must not hold
  // this screen's retry (Л3).
  const attempting = useRef(false);
  const mutate = confirm.mutate;
  const attempt = useCallback(
    (value: string) => {
      if (attempting.current) return;
      attempting.current = true;
      mutate(value, {
        onSettled: () => {
          attempting.current = false;
        },
      });
    },
    [mutate],
  );
  const started = useRef(false);
  useEffect(() => {
    if (token === undefined || started.current) return;
    started.current = true;
    attempt(token);
  }, [token, attempt]);

  const errorCode: SessionErrorCode | undefined =
    confirm.error === null
      ? undefined
      : confirm.error instanceof SessionQueryError
        ? confirm.error.code
        : 'invalid_response';
  let state: ConfirmEmailChangeState;
  if (token === undefined || (errorCode !== undefined && linkFailedCodes.has(errorCode)))
    state = { kind: 'link-failed' };
  else if (errorCode === 'email_taken') state = { kind: 'taken' };
  else if (errorCode !== undefined) state = { kind: 'failed', code: errorCode };
  else if (confirm.isSuccess) state = { kind: 'changed' };
  else state = { kind: confirm.isPaused ? 'waiting' : 'checking' };

  return (
    <ConfirmEmailChangeView
      state={state}
      retryPending={confirm.isPending}
      onRetry={() => {
        if (token !== undefined) attempt(token);
      }}
    />
  );
}
