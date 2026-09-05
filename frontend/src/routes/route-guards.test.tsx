import { afterAll, afterEach, beforeAll, expect, test } from '@rstest/core';
import { focusManager, QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { createMemoryHistory, createRouter, RouterProvider } from '@tanstack/react-router';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { http, HttpResponse } from 'msw';
import { setupServer } from 'msw/node';
import { apiOrigin } from '../adapters/api/api-origin';
import { createApiSessionRepository } from '../adapters/api/session-repository';
import { createCatalog } from '../application/catalog/catalog';
import { createSession } from '../application/identity/session';
import { createCatalogQueries } from '../app/composition/catalog-queries';
import { createSessionQueries, sessionQueryKey } from '../app/composition/identity-queries';
import type { Locale } from '../app/composition/locale';
import { createI18n, I18nProvider } from '../app/providers/i18n';
import type { CourseRepository } from '../ports/CourseRepository';
import type { AuthenticatedUser, Role } from '../domain/identity/identity';
import { routeTree } from '../routeTree.gen';

const currentUserUrl = `${apiOrigin()}/api/v1/auth/me`;
const logoutUrl = `${apiOrigin()}/api/v1/auth/logout`;
const authenticationRequired = {
  error: { code: 'authentication_required', message: 'x', requestId: 'a'.repeat(32) },
};
const dependencyUnavailable = {
  error: { code: 'dependency_unavailable', message: 'x', requestId: 'b'.repeat(32) },
};
const server = setupServer(
  http.get(currentUserUrl, () => HttpResponse.json(authenticationRequired, { status: 401 })),
);

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

function createTestRouter(
  href: string,
  cachedSession: AuthenticatedUser | null,
  locale: Locale = 'uk',
) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } });
  const i18n = createI18n(locale);
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
    Wrap: ({ children }) => <I18nProvider instance={i18n}>{children}</I18nProvider>,
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

test.each(['student', 'instructor', 'admin'] satisfies Role[])(
  'G03 opens the same /learn empty state for the %s role',
  async (role) => {
    const signedIn = user(role);
    answerWith(signedIn);
    const router = createTestRouter('/learn', signedIn);

    await router.load();
    renderRouter(router);

    expect(router.state.location.href).toBe('/learn');
    expect(screen.getByRole('main').textContent).toBe(
      'Ваше навчанняКурси з’являться тут після відкриття навчального доступу.Перейти до каталогу',
    );
    expect(screen.getByRole('link', { name: 'Перейти до каталогу' }).getAttribute('href')).toBe(
      '/courses',
    );
    expect(screen.queryByText('Сторінку не знайдено')).toBeNull();
    expect(screen.queryByText('Ваші викладацькі курси')).toBeNull();
  },
);

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

test('G40 keeps the /admin shell empty until the later phase', async () => {
  const admin = user('admin');
  answerWith(admin);
  const router = createTestRouter('/admin', admin);

  await router.load();
  renderRouter(router);

  const main = screen.getByRole('main');
  expect(main.textContent).toBe(
    'АдмініструванняКерування курсами з’явиться в наступній фазі.Перейти до каталогу',
  );
  expect(main.querySelector('table')).toBeNull();
  expect(main.querySelector('ul, ol')).toBeNull();
});

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

test('G24 signs out, clears the cached identity and navigates to the catalog', async () => {
  const admin = user('admin');
  let current: AuthenticatedUser | null = admin;
  let currentUserRequests = 0;
  let logoutRequests = 0;
  server.use(
    http.get(currentUserUrl, () => {
      currentUserRequests += 1;
      return current === null
        ? HttpResponse.json(authenticationRequired, { status: 401 })
        : HttpResponse.json(current);
    }),
    http.post(logoutUrl, () => {
      logoutRequests += 1;
      current = null;
      return new HttpResponse(null, { status: 204 });
    }),
  );
  const router = createTestRouter('/learn', admin);
  await router.load();
  renderRouter(router);
  expect(currentUserRequests).toBe(1);

  await userEvent.click(screen.getByRole('button', { name: 'Вийти' }));

  await waitFor(() => expect(router.state.location.href).toBe('/courses'));
  await waitFor(() => expect(router.state.status).toBe('idle'));
  expect(logoutRequests).toBe(1);
  expect(router.options.context.queryClient.getQueryData(sessionQueryKey)).toBe(null);
  expect(screen.getByRole('link', { name: 'Увійти' })).toBeTruthy();
  expect(screen.queryByText('admin')).toBeNull();

  await router.navigate({ to: '/learn' });
  expect(router.state.location.href).toBe('/login?returnTo=%2Flearn');
  expect(currentUserRequests).toBe(2);
});

test('G25 ignores a second sign-out while the first request is pending', async () => {
  const admin = user('admin');
  answerWith(null);
  const browser = userEvent.setup();
  let logoutRequests = 0;
  let release: (() => void) | undefined;
  const responseGate = new Promise<void>((resolve) => {
    release = resolve;
  });
  server.use(
    http.post(logoutUrl, async () => {
      logoutRequests += 1;
      await responseGate;
      return new HttpResponse(null, { status: 204 });
    }),
  );
  const router = createTestRouter('/', admin);
  await router.load();
  renderRouter(router);

  await browser.click(screen.getByRole('button', { name: 'Вийти' }));
  const pendingButton = await screen.findByRole('button', { name: 'Виходимо…' });
  expect(pendingButton.hasAttribute('disabled')).toBe(true);
  await browser.click(pendingButton);
  expect(logoutRequests).toBe(1);

  if (release === undefined) throw new Error('The delayed logout response was not initialized');
  release();
  await waitFor(() => expect(router.state.location.href).toBe('/courses'));
  await waitFor(() => expect(router.state.status).toBe('idle'));
});

test.each([
  {
    locale: 'uk',
    signOut: 'Вийти',
    login: 'Увійти',
    message: 'Сеанс у браузері завершено, але сервер не підтвердив вихід.',
  },
  {
    locale: 'en',
    signOut: 'Sign out',
    login: 'Log in',
    message: 'Your browser session ended, but the server did not confirm sign-out.',
  },
] satisfies ReadonlyArray<{ locale: Locale; signOut: string; login: string; message: string }>)(
  'G26/G31 ends the $locale local session and localizes the revocation failure',
  async ({ locale, signOut, login, message }) => {
    const admin = user('admin');
    answerWith(null);
    server.use(
      http.post(logoutUrl, () => HttpResponse.json(dependencyUnavailable, { status: 503 })),
    );
    const router = createTestRouter('/', admin, locale);
    await router.load();
    renderRouter(router);

    await userEvent.click(screen.getByRole('button', { name: signOut }));

    await waitFor(() => expect(router.state.location.href).toBe('/courses'));
    await waitFor(() => expect(router.state.status).toBe('idle'));
    expect(router.options.context.queryClient.getQueryData(sessionQueryKey)).toBe(null);
    expect(screen.getByRole('status').textContent).toContain(message);
    expect(screen.getByRole('link', { name: login })).toBeTruthy();
  },
);
