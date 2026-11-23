import { execFile } from 'node:child_process';
import { randomUUID } from 'node:crypto';
import { fileURLToPath } from 'node:url';
import { promisify } from 'node:util';
import { test, expect } from '@playwright/test';
import { deleteE2ECourse } from './course-cleanup';

const execute = promisify(execFile);
const root = fileURLToPath(new URL('../../', import.meta.url));

// Fixture-only SQL through Compose psql, as in the other specs: a publication date 31 days in
// the past cannot be made through the product. What is checked - the catalog list, its count
// and the course page - goes through the production HTTP and UI path.
async function sql(statement: string) {
  await execute(
    'docker',
    [
      'compose',
      'exec',
      '-T',
      'postgres',
      'psql',
      '-v',
      'ON_ERROR_STOP=1',
      '-U',
      'courses',
      '-d',
      'courses',
      '-c',
      statement,
    ],
    { cwd: root, timeout: 60_000 },
  );
}

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
// run-scoped fixture matches `React` or `Go` in its title or summary (the catalog's full-text
// search covers both, `courses.sql.go:37-38`), so the claim holds whatever else the catalog
// contains. Row H04 asks
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
  // The one way out leads to the last page that has courses, with the same query; no page
  // links are offered from an empty page (design-pass C2 #1, #7).
  await page.goto('/courses?q=React&limit=1&offset=1&sort=recent');
  await expect(page.getByText('На цій сторінці немає курсів')).toBeVisible();
  await expect(page.getByRole('navigation', { name: 'Сторінки каталогу' })).toHaveCount(0);
  await page.getByRole('link', { name: 'До останньої сторінки' }).click();
  await expect(page.getByRole('heading', { name: 'React: доступні інтерфейси' })).toBeVisible();
  expect(Number(new URL(page.url()).searchParams.get('offset'))).toBe(0);

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
  await page.getByRole('link', { name: 'До останньої сторінки' }).click();
  await expect(page.locator('.course-list').getByRole('listitem')).toHaveCount(1);

  // A search with no result offers to clear it, and no pages.
  await page.goto('/courses?q=zzz-no-such-course');
  await expect(page.getByText('Курсів не знайдено')).toBeVisible();
  await expect(page.getByRole('navigation', { name: 'Сторінки каталогу' })).toHaveCount(0);
  await page.getByRole('link', { name: 'Скинути пошук' }).click();
  await expect(page.getByLabel('Знайти курс')).toHaveValue('');
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

test('changing the sort applies at once, keeps focus and follows Back', async ({ page }) => {
  await page.goto('/courses?q=React&sort=recent&limit=1&offset=0');
  const field = page.getByLabel('Знайти курс');
  const sort = page.getByLabel('Сортування');
  // The typed text goes with the new order, from the first page.
  await field.fill('Go');
  await sort.focus();
  await sort.selectOption('title');
  await expect(page).toHaveURL(/[?&]q=Go&sort=title&limit=1&offset=0$/);
  await expect(
    page.getByRole('heading', { name: 'Go: фундамент надійного бекенду' }),
  ).toBeVisible();
  await expect(sort).toBeFocused();
  // Back restores the earlier address, and the controls follow it.
  await page.goBack();
  await expect(page).toHaveURL(/[?&]q=React&sort=recent/);
  await expect(page.getByRole('heading', { name: 'React: доступні інтерфейси' })).toBeVisible();
  await expect(field).toHaveValue('React');
  await expect(sort).toHaveValue('recent');
  // Interacting with the field re-renders it; the earlier text must not come back.
  await field.hover();
  await field.click();
  await expect(field).toHaveValue('React');
  await page.goForward();
  await expect(page).toHaveURL(/[?&]q=Go&sort=title/);
  await expect(field).toHaveValue('Go');
  await expect(sort).toHaveValue('title');
  await field.hover();
  await field.click();
  await expect(field).toHaveValue('Go');
});

test('a course with no lessons leaves the catalog 30 days after publication (C2b)', async ({
  page,
  request,
}) => {
  const id = randomUUID();
  // One alphanumeric token, so the catalog search scopes the check to this fixture alone.
  const token = `c2bwindow${id.replaceAll('-', '').slice(0, 12)}`;
  const slug = `c2b-window-${id.slice(0, 8)}`;
  await sql(
    `INSERT INTO courses (id,instructor_id,slug,title,summary,description,status,created_at,updated_at,published_at)
     VALUES ('${id}','40000000-0000-4000-8000-000000000002','${slug}','${token} course','${token} summary','Fixture description','published',now() - interval '40 days',now(),now() - interval '31 days')`,
  );
  try {
    const hidden = await request.get(`/api/v1/courses?q=${token}`);
    expect(hidden.status()).toBe(200);
    expect(await hidden.json()).toMatchObject({ items: [], pagination: { total: 0 } });
    await page.goto(`/courses?q=${token}`);
    await expect(page.getByText('Курсів не знайдено')).toBeVisible();
    // Leaving the list does not close the course's own page.
    const detail = await page.goto(`/courses/${slug}`);
    expect(detail?.status()).toBe(200);
    await expect(page.getByRole('heading', { level: 1 })).toHaveText(`${token} course`);

    await sql(`UPDATE courses SET published_at = now() - interval '29 days' WHERE id = '${id}'`);
    await page.goto(`/courses?q=${token}`);
    const card = page.locator('.course-list').getByRole('listitem');
    await expect(card).toHaveCount(1);
    await expect(card.getByText('Скоро буде')).toBeVisible();
    await expect(page.getByText('1 курс', { exact: true })).toBeVisible();
  } finally {
    await deleteE2ECourse(id);
  }
});
