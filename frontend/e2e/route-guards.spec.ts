import { expect, test, type Page } from '@playwright/test';
import { authFixtures as fixtures, authStatePath, recordRequest, trackPageRequests } from './auth';

let finishRequestTracking = () => Promise.resolve();

test.beforeEach(({ page }) => {
  finishRequestTracking = trackPageRequests(page);
});

test.afterEach(async () => {
  await finishRequestTracking();
});

async function submitLogin(page: Page, role: keyof typeof fixtures) {
  const loginResponse = page.waitForResponse((response) => {
    return (
      new URL(response.url()).pathname === '/api/v1/auth/login' &&
      response.request().method() === 'POST'
    );
  });
  await page.getByLabel('Електронна пошта').fill(fixtures[role].email);
  await page.getByLabel('Пароль').fill(fixtures[role].password);
  await page.getByRole('button', { name: 'Увійти', exact: true }).click();
  return loginResponse;
}

test.describe('catalog learning entry', () => {
  test.use({ storageState: authStatePath('student') });

  test('R36 keeps /learn on the signed-in learning overview', async ({ page }) => {
    await page.goto('/learn');

    await expect(page).toHaveURL((url) => url.pathname === '/learn');
    await expect(page.getByRole('heading', { name: 'Моє навчання' })).toBeVisible();
  });
});

test('anonymous protected roots and nested addresses preserve their return URL', async ({
  page,
}) => {
  for (const route of ['/admin', '/admin/courses', '/learn', '/learn/go/intro']) {
    const response = await page.goto(route);
    expect(response?.status()).toBe(200);
    await expect(page).toHaveURL((url) => {
      return url.pathname === '/login' && url.searchParams.get('returnTo') === route;
    });
    await expect(page.getByRole('heading', { name: 'Вхід у кабінет' })).toBeVisible();
  }
});

test('student returnTo reaches the not-found admin answer and the learning shell', async ({
  page,
}) => {
  await page.goto('/admin');
  await expect(page).toHaveURL((url) => url.pathname === '/login');
  const loginResponse = await submitLogin(page, 'student');
  expect(loginResponse.status()).toBe(200);

  await expect(page).toHaveURL((url) => url.pathname === '/admin');
  await expect(page.getByRole('heading', { name: 'Сторінку не знайдено' })).toBeVisible();
  await expect(page.getByText('Студентка Курсу')).toBeVisible();

  await page.goto('/learn');
  await expect(page).toHaveURL((url) => url.pathname === '/learn');
  await expect(page.getByRole('heading', { name: 'Моє навчання' })).toBeVisible();

  await page.goto('/learn/go/intro');
  // The splat lesson route renders the learning shell («Ваше навчання»), not the /learn index
  // list («Моє навчання») — only the /learn index became the my-learning list.
  await expect(page.getByRole('heading', { name: 'Ваше навчання' })).toBeVisible();
  await page.goto('/admin/courses');
  await expect(page.getByRole('heading', { name: 'Сторінку не знайдено' })).toBeVisible();
});

test('admin identity survives SSR and hydration with one protected-boundary revalidation', async ({
  page,
}) => {
  await page.goto('/admin');
  await expect(page).toHaveURL((url) => url.pathname === '/login');
  const loginResponse = await submitLogin(page, 'admin');
  expect(loginResponse.status()).toBe(200);
  expect((await loginResponse.allHeaders())['set-cookie']).toContain('courseforge_session=');
  await expect(page).toHaveURL((url) => url.pathname === '/admin');
  await expect(page.getByRole('heading', { name: 'Курси' })).toBeVisible();

  const rawAdminResponse = await page.request.get('/admin');
  expect(rawAdminResponse.status()).toBe(200);
  const adminHtml = await rawAdminResponse.text();
  expect(adminHtml).toContain(fixtures.admin.name);
  expect(adminHtml).toContain('Вийти');
  // The session query is dehydrated and already resolved, so the client has the
  // viewer before it renders. The serializer deduplicates by reference — the user
  // object is emitted once and the entry holds a pointer to it — so the name is
  // asserted above, on the document, rather than inside this entry.
  expect(adminHtml).toMatch(
    /dehydratedData.*?status:"success".{0,200}?queryKey:\$R\[\d+\]=\["session"\]/s,
  );

  const rawCoursesResponse = await page.request.get('/courses');
  expect(rawCoursesResponse.status()).toBe(200);
  expect(rawCoursesResponse.headers()['cache-control']).toBe('private, no-store');
  expect(rawCoursesResponse.headers()['vary']?.split(/\s*,\s*/)).toContain('Cookie');
  const coursesHtml = await rawCoursesResponse.text();
  expect(coursesHtml).toContain(fixtures.admin.name);
  expect(coursesHtml).toContain('Вийти');

  await page.addInitScript((name) => {
    Reflect.set(window, '__courseforgeAnonymousIdentityFrameSeen', false);
    const observer = new MutationObserver(() => {
      // The identity lives in the page header (client: its actions group beside the
      // primary nav; admin: the top bar), not inside <nav>, so probe the header.
      const banner = document.querySelector('header');
      if (banner && (!banner.textContent?.includes(name) || banner.textContent.includes('Увійти')))
        Reflect.set(window, '__courseforgeAnonymousIdentityFrameSeen', true);
    });
    observer.observe(document, { childList: true, subtree: true, characterData: true });
  }, fixtures.admin.name);

  let currentUserRequests = 0;
  page.on('request', (request) => {
    if (new URL(request.url()).pathname === '/api/v1/auth/me') currentUserRequests += 1;
  });
  await page.goto('/admin', { waitUntil: 'networkidle' });
  await expect(page.getByText(fixtures.admin.name)).toBeVisible();
  await expect(page.getByRole('heading', { name: 'Курси' })).toBeVisible();
  expect(currentUserRequests).toBe(1);

  currentUserRequests = 0;
  await page.goto('/courses', { waitUntil: 'networkidle' });
  await expect(page.getByText(fixtures.admin.name)).toBeVisible();
  expect(currentUserRequests).toBe(0);
  expect(
    await page.evaluate(() => Reflect.get(window, '__courseforgeAnonymousIdentityFrameSeen')),
  ).toBe(false);

  await page.goto('/login?returnTo=%2Fadmin');
  await expect(page.getByRole('heading', { name: 'Ви вже увійшли' })).toBeVisible();
  await page.getByRole('button', { name: 'Продовжити' }).click();
  await expect(page).toHaveURL((url) => url.pathname === '/admin');
  await expect(page.getByRole('heading', { name: 'Курси' })).toBeVisible();

  await page.goto('/learn');
  await expect(page).toHaveURL((url) => url.pathname === '/learn');
  await expect(page.getByRole('heading', { name: 'Моє навчання' })).toBeVisible();
});

test('public SPA transitions do not revalidate; learning does, and Back stays signed out', async ({
  page,
  baseURL,
}) => {
  if (baseURL === undefined) throw new Error('Playwright baseURL is required');
  const disposableSession = await page.request.post('/api/v1/auth/login', {
    data: {
      email: fixtures.student.email,
      password: fixtures.student.password,
    },
    headers: { Origin: new URL(baseURL).origin },
  });
  await recordRequest('login');
  expect(disposableSession.status()).toBe(200);

  let currentUserRequests = 0;
  page.on('request', (request) => {
    if (new URL(request.url()).pathname === '/api/v1/auth/me') currentUserRequests += 1;
  });

  await page.goto('/courses', { waitUntil: 'networkidle' });
  await page.getByRole('link', { name: 'CourseForge — головна' }).click();
  await expect(page.getByRole('heading', { level: 1 })).toContainText('Навички, які');
  // The storefront footer also links to the catalog; the header's primary
  // navigation is the one under test.
  await page
    .getByRole('navigation', { name: 'Основна навігація' })
    .getByRole('link', { name: 'Каталог', exact: true })
    .click();
  await expect(page.getByRole('heading', { name: 'Каталог курсів' })).toBeVisible();
  expect(currentUserRequests).toBe(0);
  await expect(page.getByText(fixtures.student.name)).toBeVisible();

  await page.evaluate(async () => {
    const router: unknown = Reflect.get(window, '__TSR_ROUTER__');
    if (typeof router !== 'object' || router === null)
      throw new Error('TanStack Router is unavailable');
    const navigate: unknown = Reflect.get(router, 'navigate');
    if (typeof navigate !== 'function') throw new Error('TanStack Router cannot navigate');
    await Reflect.apply(navigate, router, [{ to: '/learn' }]);
  });
  await expect(page).toHaveURL((url) => url.pathname === '/learn');
  await expect(page.getByRole('heading', { name: 'Моє навчання' })).toBeVisible();
  expect(currentUserRequests).toBe(1);

  await page.getByRole('button', { name: 'Вийти' }).click();
  await expect(page).toHaveURL((url) => url.pathname === '/courses');
  await page.goBack();
  // Back now lands on the real /learn page (the redirect was removed), and signed-out it bounces
  // through requireIdentity to /login — the point of the test holds: Back stays signed out, no
  // signed-in learning view is restored.
  await expect(page).toHaveURL((url) => url.pathname === '/login');
  await expect(page.getByText(fixtures.student.name)).toHaveCount(0);
  await expect(page.getByRole('heading', { name: 'Моє навчання' })).toHaveCount(0);
});

test.describe('with stored non-admin sessions', () => {
  test.describe('as an instructor', () => {
    test.use({ storageState: authStatePath('instructor') });

    test('the admin users route renders not-found before loading the directory', async ({
      page,
    }) => {
      let directoryRequests = 0;
      page.on('request', (request) => {
        if (new URL(request.url()).pathname === '/api/v1/admin/users') directoryRequests += 1;
      });

      await page.goto('/admin/users');

      await expect(page.getByRole('heading', { name: 'Сторінку не знайдено' })).toBeVisible();
      expect(directoryRequests).toBe(0);
    });
  });

  test.describe('as a student', () => {
    test.use({ storageState: authStatePath('student') });

    test('R20 direct API access is refused without mutating the user', async ({
      page,
      browser,
      baseURL,
    }) => {
      if (baseURL === undefined) throw new Error('Playwright baseURL is required');
      const targetID = '40000000-0000-4000-8000-0000000000d3';

      await page.goto('/courses');
      await expect(page.getByText(fixtures.student.name, { exact: true })).toBeVisible();

      const readResponse = await page.request.get('/api/v1/admin/users');
      expect(readResponse.status()).toBe(403);
      expect(await readResponse.json()).toMatchObject({ error: { code: 'forbidden' } });

      const mutationResponse = await page.request.post(`/api/v1/admin/users/${targetID}/status`, {
        data: { status: 'blocked' },
        headers: { Origin: new URL(baseURL).origin },
      });
      expect(mutationResponse.status()).toBe(403);
      expect(await mutationResponse.json()).toMatchObject({ error: { code: 'forbidden' } });

      const adminContext = await browser.newContext({
        baseURL,
        storageState: authStatePath('admin'),
      });
      try {
        const unchangedResponse = await adminContext.request.get(
          '/api/v1/admin/users?q=reset-me%40courseforge.local',
        );
        expect(unchangedResponse.status()).toBe(200);
        expect(await unchangedResponse.json()).toMatchObject({
          items: [{ id: targetID, status: 'active' }],
        });
      } finally {
        await adminContext.close();
      }
    });
  });
});
