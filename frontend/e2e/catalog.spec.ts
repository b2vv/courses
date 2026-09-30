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

// B-13: this test used to read the second card of the *global* catalog - `?limit=1&offset=1`
// - and name the course it expected there. Any sibling file that publishes a course takes
// position zero with `published_at = now` and shifts that window by one, which is not a product
// defect: `docs/foundation/assumptions.md:21` says offset pagination may shift between requests
// during publication. So the assertion demanded stability the product never promised, and
// `workers: 1` was the only thing holding it up.
//
// The window is now scoped with `q`, and the scope is what makes it deterministic: no
// run-scoped fixture matches `React` or `Go` in its title or summary (the catalog searches both,
// `courses.sql.go:355-356`), so the claim holds whatever else the catalog contains. Row H04 asks
// that reopening a q/sort/limit/offset URL show the same filtered order - which is what is
// asserted here - not that a particular course sit at a particular global offset.
//
// `workers: 1` stays, and not for this reason: five workers under load produced browser timeouts
// and 504s (`playwright.config.ts:9`). That measurement is untouched by this change, so the
// number keeps its own independent cause.
test('search and pagination live in the URL and recover from an empty page', async ({ page }) => {
  await page.goto('/courses?q=React&limit=1&offset=0&sort=recent');
  await expect(page.getByRole('heading', { name: 'React: доступні інтерфейси' })).toBeVisible();
  // The limit is honoured, not merely accepted: the scope holds exactly one course, and one card
  // is what the page renders.
  // Count the cards in the course list, not every h2 on the page: the catalog
  // also carries the value strip's own (visually hidden) heading.
  await expect(page.locator('.course-list').getByRole('listitem')).toHaveCount(1);

  // Offset one past the end of a scoped result is the empty page, and the recovery link works
  // from there - the same claim the ceiling case makes, on a boundary a fixture can reach.
  await page.goto('/courses?q=React&limit=1&offset=1&sort=recent');
  await expect(page.getByText('На цій сторінці немає курсів')).toBeVisible();
  await page.getByRole('link', { name: 'На першу сторінку' }).click();
  await expect(page.getByRole('heading', { name: 'Каталог курсів' })).toBeVisible();

  await page.goto('/courses?q=React&limit=1&offset=1&sort=recent');
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
