import { afterAll, afterEach, beforeAll, expect, test } from '@rstest/core';
import { focusManager, QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { createMemoryHistory, createRouter, RouterProvider } from '@tanstack/react-router';
import { render, screen, waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { setupServer } from 'msw/node';
import { apiOrigin } from '../adapters/api/api-origin';
import { createApiSessionRepository } from '../adapters/api/session-repository';
import { createCatalog } from '../application/catalog/catalog';
import { createSession } from '../application/identity/session';
import { createCatalogQueries } from '../app/composition/catalog-queries';
import { createSessionQueries, sessionQueryKey } from '../app/composition/identity-queries';
import type { CourseRepository } from '../ports/CourseRepository';
import type { AuthenticatedUser, Role } from '../domain/identity/identity';
import { routeTree } from '../routeTree.gen';

const currentUserUrl = `${apiOrigin()}/api/v1/auth/me`;
const server = setupServer();

beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));

afterEach(() => {
  focusManager.setFocused(undefined);
  server.resetHandlers();
});

afterAll(() => server.close());

function user(role: Role): AuthenticatedUser {
  return {
    id: `40000000-0000-4000-8000-00000000000${role === 'student' ? '1' : role === 'instructor' ? '2' : '3'}`,
    email: `${role}@courseforge.local`,
    displayName: role,
    role,
  };
}

const authenticationRequired = {
  error: { code: 'authentication_required', message: 'x', requestId: 'a'.repeat(32) },
};
const dependencyUnavailable = {
  error: { code: 'dependency_unavailable', message: 'x', requestId: 'b'.repeat(32) },
};

function createTestRouter(href: string, cachedSession: AuthenticatedUser | null) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const identityDependencies = {
    ...createSessionQueries(createSession(createApiSessionRepository())),
    shouldLoadCurrentUser: true,
  };
  const catalogRepository: CourseRepository = {
    listPublished: async () => ({ ok: false, error: { code: 'dependency_unavailable' } }),
    getPublished: async () => ({ ok: false, error: { code: 'course_not_found' } }),
  };
  const router = createRouter({
    routeTree,
    history: createMemoryHistory({ initialEntries: [href] }),
    context: {
      queryClient,
      identityDependencies,
      catalog: createCatalogQueries(createCatalog(catalogRepository)),
    },
    defaultPendingMs: 0,
    defaultPendingMinMs: 0,
  });
  queryClient.setQueryData(sessionQueryKey, cachedSession);
  return router;
}

function renderRouter(router: ReturnType<typeof createTestRouter>) {
  return render(
    <QueryClientProvider client={router.options.context.queryClient}>
      <RouterProvider router={router} />
    </QueryClientProvider>,
  );
}

function answerWith(current: AuthenticatedUser | null) {
  server.use(
    http.get(currentUserUrl, () =>
      current === null
        ? HttpResponse.json(authenticationRequired, { status: 401 })
        : HttpResponse.json(current),
    ),
  );
}

test.each([
  ['/learn', '/login?returnTo=%2Flearn'],
  ['/admin', '/login?returnTo=%2Fadmin'],
])('G01/G02/G13 redirects an anonymous visit to %s', async (href, expected) => {
  answerWith(null);
  const router = createTestRouter(href, null);

  await router.load();

  expect(router.state.location.href).toBe(expected);
});

test('G04 gives a student the same not-found answer for /admin as an unknown address', async () => {
  const student = user('student');
  answerWith(student);
  const adminRouter = createTestRouter('/admin', student);
  await adminRouter.load();
  const adminScreen = renderRouter(adminRouter);
  const adminAnswer = screen.getByRole('heading', { name: 'Сторінку не знайдено' }).textContent;
  adminScreen.unmount();

  const unknownRouter = createTestRouter('/unknown', student);
  await unknownRouter.load();
  renderRouter(unknownRouter);

  expect(screen.getByRole('heading', { name: 'Сторінку не знайдено' }).textContent).toBe(
    adminAnswer,
  );
});

test.each(['instructor', 'admin'] satisfies Role[])(
  'G05 opens /admin for the %s role',
  async (role) => {
    const signedIn = user(role);
    answerWith(signedIn);
    const router = createTestRouter('/admin', signedIn);

    await router.load();
    renderRouter(router);

    expect(router.state.location.href).toBe('/admin');
    expect(screen.getByRole('heading', { name: 'Адміністрування' })).toBeTruthy();
    expect(screen.queryByText('Сторінку не знайдено')).toBeNull();
  },
);

test.each([
  ['/learn/go/intro?lesson=3', null, '/login?returnTo=%2Flearn%2Fgo%2Fintro%3Flesson%3D3'],
  ['/admin/courses', null, '/login?returnTo=%2Fadmin%2Fcourses'],
] as const)('G06 redirects anonymous nested address %s', async (href, cached, expected) => {
  answerWith(null);
  const router = createTestRouter(href, cached);

  await router.load();

  expect(router.state.location.href).toBe(expected);
});

test('G06 applies the parent role matrix to signed-in nested addresses', async () => {
  const student = user('student');
  answerWith(student);
  const learnRouter = createTestRouter('/learn/go/intro', student);
  await learnRouter.load();
  const learnScreen = renderRouter(learnRouter);
  expect(screen.getByRole('heading', { name: 'Ваше навчання' })).toBeTruthy();
  learnScreen.unmount();

  const adminRouter = createTestRouter('/admin/courses', student);
  await adminRouter.load();
  renderRouter(adminRouter);
  expect(screen.getByRole('heading', { name: 'Сторінку не знайдено' })).toBeTruthy();
});

test('G08 shows the route pending state while protected navigation revalidates', async () => {
  const admin = user('admin');
  let release: (() => void) | undefined;
  let requestStarted = false;
  const responseGate = new Promise<void>((resolve) => {
    release = resolve;
  });
  server.use(
    http.get(currentUserUrl, async () => {
      requestStarted = true;
      await responseGate;
      return HttpResponse.json(admin);
    }),
  );
  const router = createTestRouter('/', admin);
  await router.load();
  renderRouter(router);

  const navigation = router.navigate({ to: '/admin' });
  await waitFor(() => expect(requestStarted).toBe(true));
  expect(await screen.findByText('Перевіряємо сеанс')).toBeTruthy();
  if (release === undefined) throw new Error('The delayed identity response was not initialized');
  release();
  await navigation;

  expect(await screen.findByRole('heading', { name: 'Адміністрування' })).toBeTruthy();
});

test('G09 renders a localized retryable route error when identity is unavailable', async () => {
  const admin = user('admin');
  server.use(
    http.get(currentUserUrl, () => HttpResponse.json(dependencyUnavailable, { status: 503 })),
  );
  const router = createTestRouter('/admin', admin);

  await router.load();
  renderRouter(router);

  expect(screen.getByRole('heading', { name: 'Не вдалося перевірити сеанс' })).toBeTruthy();
  expect(screen.getByRole('link', { name: 'Спробувати ще раз' }).getAttribute('href')).toBe(
    '/admin',
  );
  expect(router.state.location.href).toBe('/admin');

  answerWith(admin);
  await router.invalidate({ sync: true });

  expect(await screen.findByRole('heading', { name: 'Адміністрування' })).toBeTruthy();
});

test('G11 revalidates a cached identity and redirects when the session has ended', async () => {
  const admin = user('admin');
  answerWith(null);
  const router = createTestRouter('/admin', admin);

  await router.load();

  expect(router.state.location.href).toBe('/login?returnTo=%2Fadmin');
  expect(router.options.context.queryClient.getQueryData(sessionQueryKey)).toBe(null);
});

test('G12 leaves a mounted protected route when focus revalidation ends the session', async () => {
  const admin = user('admin');
  let current: AuthenticatedUser | null = admin;
  server.use(
    http.get(currentUserUrl, () =>
      current === null
        ? HttpResponse.json(authenticationRequired, { status: 401 })
        : HttpResponse.json(current),
    ),
  );
  const router = createTestRouter('/admin', admin);
  await router.load();
  renderRouter(router);
  expect(screen.getByRole('heading', { name: 'Адміністрування' })).toBeTruthy();

  current = null;
  focusManager.setFocused(false);
  focusManager.setFocused(true);

  await waitFor(() => expect(router.state.location.href).toBe('/login?returnTo=%2Fadmin'));
});
