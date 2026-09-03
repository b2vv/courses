import { test, expect } from '@playwright/test';

test('catalog and detail include PostgreSQL course data in initial HTML', async ({
  browser,
  baseURL,
}) => {
  const context = await browser.newContext({
    javaScriptEnabled: false,
    baseURL: baseURL ?? 'http://localhost:8080',
  });
  const page = await context.newPage();
  const response = await page.goto('/courses');
  expect(response?.status()).toBe(200);
  await expect(
    page.getByRole('heading', { name: 'Go: фундамент надійного бекенду' }),
  ).toBeVisible();
  await expect(page.getByText('Чернетка курсу')).toHaveCount(0);
  await page
    .getByRole('heading', { name: 'Go: фундамент надійного бекенду' })
    .getByRole('link')
    .click();
  await expect(page.getByRole('heading', { level: 1 })).toHaveText(
    'Go: фундамент надійного бекенду',
  );
  await expect(page.getByRole('heading', { name: 'Про курс' })).toBeVisible();
  await context.close();
});

test('search and pagination live in the URL and recover from an empty page', async ({ page }) => {
  await page.goto('/courses?limit=1&offset=1&sort=recent');
  await expect(page.getByRole('heading', { name: 'React: доступні інтерфейси' })).toBeVisible();
  await page.getByLabel('Знайти курс').fill('Go');
  await page.getByRole('button', { name: 'Знайти', exact: true }).click();
  await expect(
    page.getByRole('heading', { name: 'Go: фундамент надійного бекенду' }),
  ).toBeVisible();
  expect(Number(new URL(page.url()).searchParams.get('offset') ?? 0)).toBe(0);
  expect(new URL(page.url()).searchParams.get('q')).toBe('Go');
  await page.goto('/courses?limit=1&offset=10000');
  await expect(page.getByText('На цій сторінці немає курсів')).toBeVisible();
  await page.getByRole('link', { name: 'На першу сторінку' }).click();
  await expect(page.getByRole('heading', { name: 'Каталог курсів' })).toBeVisible();
});

test('hidden and absent courses return 404', async ({ page }) => {
  for (const slug of ['draft-course', 'archived-course', 'does-not-exist']) {
    const response = await page.goto(`/courses/${slug}`);
    expect(response?.status()).toBe(404);
    await expect(page.getByText('Курс не знайдено')).toBeVisible();
  }
});

test('catalog hydrates without duplicate data requests or errors', async ({ page }) => {
  const errors: string[] = [];
  const requests: string[] = [];
  page.on('pageerror', (error) => errors.push(error.message));
  page.on('console', (message) => {
    if (message.type() === 'error') errors.push(message.text());
  });
  page.on('request', (request) => {
    if (request.url().includes('/api/v1/courses')) requests.push(request.url());
  });
  await page.goto('/courses');
  await page.getByLabel('Знайти курс').fill('test hydration');
  await expect(page.getByLabel('Знайти курс')).toHaveValue('test hydration');
  expect(errors).toEqual([]);
  expect(requests).toEqual([]);
});
