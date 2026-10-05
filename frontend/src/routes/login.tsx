import { createFileRoute, useNavigate } from '@tanstack/react-router';
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { sessionQueryKey, SessionQueryError } from '../app/composition/identity-queries';
import { parseLoginSearch } from '../app/router/login-search';
import { LoginView } from '../presentation/identity/LoginView';
import { safeReturnTo } from '../domain/identity/identity';
import type { AuthenticatedUser, CredentialsInput } from '../domain/identity/identity';

export const Route = createFileRoute('/login')({
  validateSearch: parseLoginSearch,
  head: () => ({ meta: [{ title: 'Вхід — CourseForge' }] }),
  component: LoginRoute,
  pendingComponent: LoginPending,
  errorComponent: LoginFailure,
});

function LoginPending() {
  const { t } = useTranslation();
  return (
    <section className="notice" aria-busy="true">
      <p>{t('login.checking')}</p>
    </section>
  );
}

function LoginFailure() {
  const { t } = useTranslation();
  return (
    <section className="notice" role="alert">
      <h1>{t('login.failure')}</h1>
      <p>{t('login.failureHint')}</p>
      <form method="get">
        <button type="submit">{t('common.retry')}</button>
      </form>
    </section>
  );
}

function LoginRoute() {
  const { returnTo, notice } = Route.useSearch();
  const destination = safeReturnTo(returnTo);
  const { identity, identityDependencies } = Route.useRouteContext();
  const client = useQueryClient();
  const navigate = useNavigate();
  const user = identity.status === 'signed-in' ? identity.user : null;
  const goToReturnTo = () => void navigate({ href: destination });
  const logIn = useMutation({
    mutationFn: (input: CredentialsInput) =>
      identityDependencies.session
        .logIn(input)
        .then((result) => identityDependencies.unwrap(result)),
    onSuccess: (signedIn: AuthenticatedUser) => {
      // Replace the identity outright rather than refetching: the answer is
      // already authoritative and a refetch would race the navigation.
      client.setQueryData(sessionQueryKey, signedIn);
      goToReturnTo();
    },
  });
  const failure = logIn.error;
  return (
    <LoginView
      user={user}
      isPending={logIn.isPending}
      errorCode={failure instanceof SessionQueryError ? failure.code : undefined}
      notice={notice}
      onSubmit={(input) => logIn.mutate(input)}
      onContinue={goToReturnTo}
    />
  );
}
