import { expect, test, type Page } from '@playwright/test';

const baseURL = process.env.E2E_BASE_URL ?? 'http://localhost:8080';
const fixtures = {
  student: {
    email: 'student@courseforge.local',
    password: 'courseforge-student-demo',
    name: 'Студентка Курсу',
  },
  admin: {
    email: 'admin@courseforge.local',
    password: 'courseforge-admin-demo',
    name: 'Адміністратор Курсу',
  },
} as const;

async function authenticate(page: Page, role: keyof typeof fixtures) {
  const { email, password } = fixtures[role];
  const response = await page.request.post('/api/v1/auth/login', {
    data: { email, password },
    headers: { Origin: new URL(baseURL).origin },
  });
  expect(response.status()).toBe(200);
}

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
  await expect(page.getByRole('heading', { name: 'Ваше навчання' })).toBeVisible();
  await expect(page.getByRole('link', { name: 'Перейти до каталогу' })).toBeVisible();

  await page.goto('/learn/go/intro');
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
  await expect(page.getByRole('heading', { name: 'Адміністрування' })).toBeVisible();
  await expect(page.getByText('Керування курсами з’явиться в наступній фазі.')).toBeVisible();

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
      const navigation = document.querySelector('header nav');
      if (
        navigation &&
        (!navigation.textContent?.includes(name) || navigation.textContent.includes('Увійти'))
      )
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
  await expect(page.getByRole('heading', { name: 'Адміністрування' })).toBeVisible();
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
  await expect(page.getByRole('heading', { name: 'Адміністрування' })).toBeVisible();

  await page.goto('/learn');
  await expect(page.getByRole('heading', { name: 'Ваше навчання' })).toBeVisible();
});

test('public SPA transitions do not revalidate; learning does, and Back stays signed out', async ({
  page,
}) => {
  await authenticate(page, 'student');
  let currentUserRequests = 0;
  page.on('request', (request) => {
    if (new URL(request.url()).pathname === '/api/v1/auth/me') currentUserRequests += 1;
  });

  await page.goto('/courses', { waitUntil: 'networkidle' });
  await page.getByRole('link', { name: 'CourseForge — головна' }).click();
  await expect(page.getByRole('heading', { level: 1 })).toContainText('Навички, які');
  await page.getByRole('link', { name: 'Каталог', exact: true }).click();
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
  await expect(page.getByRole('heading', { name: 'Ваше навчання' })).toBeVisible();
  expect(currentUserRequests).toBe(1);

  await page.getByRole('button', { name: 'Вийти' }).click();
  await expect(page).toHaveURL((url) => url.pathname === '/courses');
  await page.goBack();
  await expect(page).toHaveURL((url) => {
    return url.pathname === '/login' && url.searchParams.get('returnTo') === '/learn';
  });
  await expect(page.getByText(fixtures.student.name)).toHaveCount(0);
  await expect(page.getByRole('heading', { name: 'Ваше навчання' })).toHaveCount(0);
});
